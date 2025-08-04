package cmd

import (
	"authz/api"
	"authz/internal/handler/rest/middleware"
	"authz/internal/pkg"
	"authz/internal/service/logger"
	"authz/internal/wire"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/gox/logx"
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
	Cmd.Flags().
		StringVarP(&pkg.Env, `env`, "e", "local", `default: local`)

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
	lc := pkg.NewSimpleLifecycle()

	h, err := wire.NewHandler(lc)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	e := NewGinEngine()
	api.RegisterAPIHandlers(e, h, middleware.Jwt())
	server := NewHttpServer(lc)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := lc.Shutdown(ctx); err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func NewHttpServer(lc pkg.Lifecycle) *http.Server {
	server := &http.Server{
		Addr:              ":8080",
		Handler:           NewGinEngine(),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	lc.Add(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	})

	return server
}
