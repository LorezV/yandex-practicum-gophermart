package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

const SecretKey = "EXAMPLE_SECRET_KEY"
const PasswordSalt = "EXAMPLE_PASSWORD_SALT"

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func MakeConfig() *Config {
	var cfg Config

	flag.StringVar(&cfg.RunAddress, "a", "localhost:8000", "An address and port for server to start")
	flag.StringVar(&cfg.DatabaseURI, "d", "postgres://postgres:admin@localhost:5432/go-diploma", "An address of DB connection")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://127.0.0.1:8080", "An address of the Accrual System")

	flag.Parse()
	if err := env.Parse(&cfg); err != nil {
		log.Fatal().Err(err).Msg("Parsing env")
		return nil
	}

	log.Debug().Interface("config", cfg).Send()
	return &cfg
}
