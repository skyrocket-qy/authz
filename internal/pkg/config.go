package pkg

import (
	"authz/internal/cfg"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func NewConfig() (err error) {
	if Env == "local" {
		err := godotenv.Load(".env")
		if err != nil {
			return err
		}
	}

	if err := env.Parse(&cfg.Cfg); err != nil {
		return err
	}

	return nil
}
