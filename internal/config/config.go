package config

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
)

const SecretKey = "EXAMPLE_SECRET_KEY"
const PasswordSalt = "EXAMPLE_PASSWORD_SALT"

//var Config struct {
//	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8000"`
//	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS " envDefault:"http://127.0.0.1:8080"`
//	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgres://postgres:admin@localhost:5432/go-diploma"`
//}

var Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS "`
	DatabaseURI          string `env:"DATABASE_URI"`
}

func InitConfig() {
	flag.StringVar(&Config.RunAddress, "a", "localhost:8080", "An address and port for server to start")
	flag.StringVar(&Config.DatabaseURI, "d", "", "An address of DB connection")
	flag.StringVar(&Config.AccrualSystemAddress, "r", "", "An address of the Accrual System")

	flag.Parse()
	if err := env.Parse(&Config); err != nil {
		log.Fatal("can't parse env")
	}
}
