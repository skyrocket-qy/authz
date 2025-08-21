package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"authz/cmd/service"
	"authz/cmd/tool"
	"authz/internal/pkg"
	"authz/internal/service/logger"
	"authz/internal/wire"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/gox/logx"
	"github.com/skyrocket-qy/protos/gen/authzpb/rbacpb/rbacpbconnect"
	"github.com/skyrocket-qy/protos/gen/authzpb/v1/authzpbv1connect"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var Cmd = &cobra.Command{
	Use:   "",
	Short: "A brief description of your application",
	Long:  `The longer description`,
	Run:   RunServer,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := Cmd.Execute()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func init() {
	Cmd.AddCommand(service.Cmd)
	Cmd.AddCommand(tool.Cmd)

	Cmd.PersistentFlags().StringVarP(&pkg.Env, `env`, "e", "local", `default: local`)
	Cmd.Flags().StringP("engine", `g`, "rbac", "default: rbac")

	Cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		validEnvs := map[string]bool{"local": true, "dev": true, "prod": true, "stage": true}
		if !validEnvs[pkg.Env] {
			return fmt.Errorf(
				"invalid environment value: %s. Must be one of: dev, prod",
				pkg.Env,
			)
		}

		return nil
	}
}

func NewGinEngine() *gin.Engine {
	if pkg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	return gin.Default()
}

func RunServer(cmd *cobra.Command, args []string) {
	if err := pkg.NewConfig(); err != nil {
		logx.Error(err.Error())

		return
	}

	logger.InitLogger()

	lc := pkg.NewLifecycleParallel()

	startConnectServer(lc)
}

var inflight = &sync.WaitGroup{}

func startConnectServer(lc *pkg.LifecycleParallel) {
	connectH, err := wire.NewRbacHandler(context.TODO(), lc)
	if err != nil {
		log.Error().Msg(err.Error())

		return
	}

	inflightInterceptor := connect.UnaryInterceptorFunc(func(
		next connect.UnaryFunc,
	) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			inflight.Add(1)
			defer inflight.Done()

			return next(ctx, req)
		}
	})

	path, handler := authzpbv1connect.NewAuthzServiceHandler(connectH,
		connect.WithCompressMinBytes(512),
		connect.WithInterceptors(inflightInterceptor),
	)

	rbacPath, rbacH := rbacpbconnect.NewRbacServiceHandler(connectH,
		connect.WithCompressMinBytes(512),
		connect.WithInterceptors(inflightInterceptor),
	)
	mux := http.NewServeMux()
	mux.Handle(rbacPath, rbacH)
	mux.Handle(path, handler)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Or "*" for dev
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
	})

	handlerWithCors := c.Handler(mux)
	server := NewHttpServer(lc, handlerWithCors)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Run the server in a goroutine so the main thread can listen for signals.
	go func() {
		log.Info().Msgf("Starting server on %s", server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Could not start server")
		}
	}()

	// Block the main function until an OS signal is received.
	<-stop
	log.Info().Msg("Received interrupt signal, initiating graceful shutdown...")

	// Call the lifecycle shutdown method with a timeout.
	// The server.Shutdown() call is registered in NewHttpServer and will be executed here.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lc.Shutdown(ctx)

	log.Info().Msg("Server gracefully shut down.")
}

func NewHttpServer(lc *pkg.LifecycleParallel, handler http.Handler) *http.Server {
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	lc.Add(server, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return err
		}

		inflight.Wait()

		return nil
	})

	return server
}
