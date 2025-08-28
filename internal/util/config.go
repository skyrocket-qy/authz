package util

import (
	"authz/internal/config"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/skyrocket-qy/erx"
)

func NewConfig() (err error) {
	if config.Env == config.EnvLocal {
		err := godotenv.Load(".env")
		if err != nil {
			return erx.W(err)
		}
	}

	if err := env.Parse(&config.Conf); err != nil {
		return erx.W(err)
	}

	return nil
}
