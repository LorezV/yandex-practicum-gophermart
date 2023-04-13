package userrepository

import (
	"context"
	"fmt"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID           int
	Login        string
	Password     string
	PasswordSalt string
	Balance      float64
}

func CreateUserTable(ctx context.Context) error {
	_, err := database.Connection.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "user" (
			"id" SERIAL PRIMARY KEY,
			"login" VARCHAR(128) NOT NULL UNIQUE,
			"password" VARCHAR(128) NOT NULL,
			"password_salt" VARCHAR(128) NOT NULL,
			"balance" FLOAT NOT NULL DEFAULT 0
		);
	`)

	return err
}

func Get(ctx context.Context, field string, value interface{}) (user User, err error) {
	row := database.Connection.QueryRow(ctx, fmt.Sprintf("SELECT * FROM \"public\".\"user\" WHERE %s=$1", field), value)
	err = scanUser(row, &user)
	return
}

func Create(ctx context.Context, login string, passwordHash string, passwordSalt string) (err error) {
	_, err = database.Connection.Exec(ctx, `INSERT INTO "public"."user" (login, password, password_salt) VALUES ($1, $2, $3)`, login, passwordHash, passwordSalt)
	return
}

func Update(ctx context.Context, user *User) (err error) {
	_, err = database.Connection.Exec(ctx, `UPDATE "public"."user" SET balance=$1 WHERE id=$2`, user.Balance, user.ID)
	return
}

func Withdraw(ctx context.Context, user User, order string, sum float64) (err error) {
	tx, err := database.Connection.Begin(ctx)
	if err != nil {
		return
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO "public"."withdrawal" (user_id, "order", sum) VALUES ($1, $2, $3);`, user.ID, order, sum)
	if err != nil {
		return
	}

	_, err = tx.Exec(ctx, `UPDATE "public"."user" SET balance = $1 WHERE id=$2;`, user.Balance, user.ID)
	if err != nil {
		return
	}

	tx.Commit(ctx)
	return
}

func scanUser(row pgx.Row, user *User) error {
	return row.Scan(&user.ID, &user.Login, &user.Password, &user.PasswordSalt, &user.Balance)
}
