package config

var Env Environment

type Environment string

const (
	EnvProd  Environment = "prod"
	EnvStage Environment = "stage"
	EnvQa    Environment = "qa"
	EnvDev   Environment = "dev"
	EnvLocal Environment = "local"
)
