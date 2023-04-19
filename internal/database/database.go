package database

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Database struct {
	*pgxpool.Pool
}

func MakeConnection(ctx context.Context, address string) (*Database, error) {
	pool, err := pgxpool.Connect(ctx, address)
	if err != nil {
		return nil, err
	}

	db := &Database{
		pool,
	}

	go func() {
		<-ctx.Done()
		db.Close()
	}()

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}
	if err = db.Migrate(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
