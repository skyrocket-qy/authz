package util

import (
	"authz/internal/config"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/skyrocket-qy/erx"
)

func NewConfig() (err error) {
	if config.Env == config.EnvLocal {
		err := godotenv.Load(".env")
		// In a local environment, it's okay for the .env file to be missing.
		// We should only return an error if one occurred that is NOT a "file does not exist" error.
		if err != nil && !os.IsNotExist(err) {
			return erx.W(err)
		}
	}

	if err := env.Parse(&config.Conf); err != nil {
		return erx.W(err)
	}

	return nil
}
