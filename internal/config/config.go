package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type DBConfig struct {
	DBHost     string `env:"DB_HOST" env-default:"db"`
	DBUser     string `env:"DB_USER" env-default:"postgres"`
	DBPassword string `env:"DB_PASSWORD" env-default:"postgres"`
	DBName     string `env:"DB_NAME" env-default:"postgres"`
	DBPort     string `env:"DB_PORT" env-default:"5432"`
}

type Config struct {
	HttpServerAddress string        `env:"HTTP_SERVER_ADDRESS" env-default:"localhost:8081"`
	HttpServerTimeout time.Duration `env:"HTTP_SERVER_TIMEOUT" env-default:"5s"`
	LogLevel          string        `env:"LOG_LEVEL" env-default:"DEBUG"`
	DBConfig          DBConfig
}

func MustLoadCfg(configPath string) Config {
	if err := godotenv.Load(configPath); err != nil {
		log.Fatalf("failed to load .env file: %s", err)
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to read environment variables: %s", err)
	}

	return cfg
}
