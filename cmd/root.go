package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"authz/cmd/server"
	"authz/cmd/tool"
	"authz/internal/config"
	"authz/internal/handler/connect/middleware"
	"authz/internal/handler/rest"
	"authz/internal/service"
	"authz/internal/service/database"
	"authz/internal/service/logx"
	"authz/internal/util"
	"authz/internal/wire"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/erx"
	"github.com/skyrocket-qy/protos/gen/authzpb/rbacpb/rbacpbconnect"
	"github.com/skyrocket-qy/protos/gen/authzpb/v1/authzpbv1connect"
	"github.com/spf13/cobra"
)

var inflight = &sync.WaitGroup{}

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
		log.Fatal().Err(err).Msg("Failed to execute")
	}
}

func init() {
	Cmd.AddCommand(server.Cmd)
	Cmd.AddCommand(tool.Cmd)

	Cmd.PersistentFlags().StringVarP(&config.Env, `env`, "e", "local", `default: local`)
	Cmd.Flags().StringP("engine", `g`, "rbac", "default: rbac")

	Cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// validEnvs := map[string]bool{"local": true, "dev": true, "prod": true, "stage": true}
		// if !validEnvs[util.Env] {
		// 	return fmt.Errorf(
		// 		"invalid environment value: %s. Must be one of: dev, prod",
		// 		util.Env,
		// 	)
		// }
		return nil
	}
}

func RunServer(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() // Release resources when the function returns

	lc := util.NewLifecycleParallel()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := lc.Shutdown(shutdownCtx); err != nil {
			log.Error().Msg(erx.FullMsg(err))
		}

		log.Info().Msg("Server gracefully shut down.")
	}()

	if err := setupBase(ctx, lc); err != nil {
		log.Error().Msg(erx.FullMsg(err))

		return
	}

	if err := startConnectServer(ctx, lc); err != nil {
		log.Error().Msg(erx.FullMsg(err))

		return
	}

	// 4. Wait for the context to be cancelled (from an OS signal).
	<-ctx.Done()
	log.Info().Msg("Shutdown signal received, initiating graceful shutdown...")
}

func setupBase(ctx context.Context, lc *util.LifecycleParallel) error {
	if err := util.NewConfig(); err != nil {
		return erx.W(err)
	}

	if err := logx.InitLogger(); err != nil {
		return erx.W(err)
	}

	// Use the application context here instead of context.TODO()
	if config.Env != config.EnvProd {
		shutdown, err := service.SetupOTelSDK(ctx)
		if err != nil {
			return erx.W(err)
		}

		lc.Add("otel", shutdown)
	}

	return nil
}

func startConnectServer(ctx context.Context, lc *util.LifecycleParallel) error {
	db, err := database.New(lc)
	if err != nil {
		return erx.W(err, "failed to init db")
	}

	connectH, err := wire.NewRbacHandler(ctx, lc, db)
	if err != nil {
		return erx.W(err, "failed to init connect handler")
	}

	restH := rest.NewHandler(db)

	inflightInterceptor := connect.UnaryInterceptorFunc(func(
		next connect.UnaryFunc,
	) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			inflight.Add(1)
			defer inflight.Done()

			return next(ctx, req)
		}
	})

	handlerOpts := []connect.HandlerOption{
		connect.WithCompressMinBytes(512),
		connect.WithInterceptors(inflightInterceptor),
		connect.WithInterceptors(middleware.NewLogRequest()),
	}

	if config.Env != config.EnvProd {
		otelInterceptor, err := otelconnect.NewInterceptor()
		if err != nil {
			return erx.W(err, "failed to init otel interceptor")
		}

		handlerOpts = append(handlerOpts, connect.WithInterceptors(otelInterceptor))
	}

	path, handler := authzpbv1connect.NewAuthzServiceHandler(connectH, handlerOpts...)
	rbacPath, rbacH := rbacpbconnect.NewRbacServiceHandler(connectH, handlerOpts...)
	mux := http.NewServeMux()
	mux.Handle(rbacPath, rbacH)
	mux.Handle(path, handler)
	restH.RegisterRoutes(mux)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Or "*" for dev
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
	})

	handlerWithCors := c.Handler(mux)
	server := NewHttpServer(lc, handlerWithCors)

	// 2. Run the server in a goroutine so it doesn't block.
	go func() {
		log.Info().Msgf("Starting server on %s", server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Err(err).Msg("Failed to start server")
		}
	}()

	return nil
}

func NewHttpServer(lc *util.LifecycleParallel, handler http.Handler) *http.Server {
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	lc.Add(server, func(c context.Context) error {
		if err := server.Shutdown(c); err != nil {
			return err
		}

		inflight.Wait()

		return nil
	})

	return server
}
