package pkg

import (
	"authz/internal/cfg"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func NewConfig() (err error) {
	if Env == "local" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Error().Msg("Error loading .env file")

			return err
		}
	}

	if err := env.Parse(&cfg.Cfg); err != nil {
		return err
	}

	return nil
}
