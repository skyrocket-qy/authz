package cfg

var Cfg Config

type Config struct {
	Port int `env:"PORT"`
	Db   struct {
		User     string `env:"DB_USER"`
		Password string `env:"DB_PASSWORD"`
		Host     string `env:"DB_HOST"`
		Port     int    `env:"DB_PORT"`
		Db       string `env:"DB_DB"`
	}

	Google struct {
		ClientID     string `env:"GOOGLE_CLIENT_ID"`
		ClientSecret string `env:"GOOGLE_CLIENT_SECRET"`
		RedirectURL  string `env:"GOOGLE_REDIRECT_URL"`
	}

	Jwt struct {
		Secret string `env:"JWT_SECRET"`
	}

	Redis struct {
		Host     string `env:"REDIS_HOST"`
		Port     string `env:"REDIS_PORT"`
		Password string `env:"REDIS_PASSWORD"`
	}

	Aws struct {
		Region    string `env:"AWS_REGION"`
		AccessKey string `env:"AWS_ACCESS_KEY"`
		SecretKey string `env:"AWS_SECRET_KEY"`

		Ses struct {
			SenderEmail string `env:"AWS_SES_SENDER_EMAIL"`
		}
	}

	Kafka struct {
		Broker string `env:"KAFKA_BROKER"`
	}
}
