package config

import "github.com/caarlos0/env"

var Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8000"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS " envDefault:"http://127.0.0.1:8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgres://postgres:admin@localhost:5432/go-diploma"`
	PasswordSalt         string `env:"PASSWORD_SALT" envDefault:"EXAMPLE_PASSWORD_SALT"`
	SecretKey            string `env:"SECRET_KEY" envDefault:"EXAMPLE_SECRET_KEY"`
}

func InitConfig() error {
	return env.Parse(&Config)
}
