package config

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
)

const SecretKey = "EXAMPLE_SECRET_KEY"
const PasswordSalt = "EXAMPLE_PASSWORD_SALT"

var Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8000"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS " envDefault:"http://127.0.0.1:8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgres://postgres:admin@localhost:5432/go-diploma"`
}

func InitConfig() {
	flag.StringVar(&Config.RunAddress, "a", "localhost:8000", "ip:port")
	flag.StringVar(&Config.DatabaseURI, "d", "postgres://postgres:admin@localhost:5432/go-diploma", "postgres://login:password@host:port/database")
	flag.StringVar(&Config.AccrualSystemAddress, "r", "", "An address of the Accrual System")

	flag.Parse()
	if err := env.Parse(&Config); err != nil {
		log.Fatal("can't parse env")
	}
}
