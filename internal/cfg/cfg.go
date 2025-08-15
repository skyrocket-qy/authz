package cfg

var Cfg Config

type Config struct {
	// MaxCheckDepth int `env:"MAX_CHECK_DEPTH"` // TODO: not implemented but optional
	MaxCheckNodes int `env:"MAX_CHECK_NODES"` // TODO: not implemented

	SchemaPath string `env:"SCHEMA_PATH"`
	Port       int    `env:"PORT"`
	Db         struct {
		User     string `env:"DB_USER"`
		Password string `env:"DB_PASSWORD"`
		Host     string `env:"DB_HOST"`
		Port     int    `env:"DB_PORT"`
		Db       string `env:"DB_DB"`
	}

	Jwt struct {
		Secret string `env:"JWT_SECRET"`
	}

	Redis struct {
		Host     string `env:"REDIS_HOST"`
		Port     string `env:"REDIS_PORT"`
		Password string `env:"REDIS_PASSWORD"`
	}

	Kafka struct {
		Broker string `env:"KAFKA_BROKER"`
	}
}
