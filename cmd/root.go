package cmd

import (
	"authz/api"
	"authz/internal/handler/rest"
	"authz/internal/pkg"
	"authz/internal/service/database"
	"authz/internal/service/logger"
	"authz/internal/service/redis"
	"context"
	"fmt"
	"net/http"
	"time"

	validate "authz/internal/service/validate"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/gox/logx"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
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

func StartHTTPServer(lc fx.Lifecycle, r *gin.Engine) {
	server := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTP server ListenAndServe: %v", err)
				}
			}()

			return nil // Allow OnStart to return quickly
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}

func RunServer(cmd *cobra.Command, args []string) {
	if err := pkg.NewConfig(); err != nil {
		logx.Error(err.Error())

		return
	}

	logger.InitLogger()

	var app *fx.App

	app = fx.New(
		fx.Supply(
			fx.Annotate(
				context.TODO(),
				fx.As(new(context.Context)),
			),
		),
		fx.Provide(
			database.New,
			redis.New,
			NewGinEngine,
			rest.NewHandler,
		),
		fx.Invoke(
			validate.New,
			pkg.InitSwagger,
			api.RegisterAPIHandlers,
			StartHTTPServer,
		),
	)
	app.Run()
}
