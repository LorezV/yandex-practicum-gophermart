package config

import (
	"github.com/caarlos0/env"
)

const SecretKey = "EXAMPLE_SECRET_KEY"
const PasswordSalt = "EXAMPLE_PASSWORD_SALT"

var Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8000"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS " envDefault:"http://127.0.0.1:8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgres://postgres:admin@localhost:5432/go-diploma"`
}

func InitConfig() error {
	return env.Parse(&Config)
}
