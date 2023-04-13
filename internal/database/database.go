package database

import (
	"context"
	"github.com/jackc/pgx/v5"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/LorezV/go-diploma.git/internal/config"
)

var Connection *pgx.Conn

func InitConnection() (err error) {
	Connection, err = pgx.Connect(context.Background(), config.Config.DatabaseURI)
	if err != nil {
		return
	}

	err = CheckConnection()
	return
}

func CheckConnection() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = Connection.Ping(ctx)
	return
}
