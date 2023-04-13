package withdrawalrepository

import (
	"context"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/jackc/pgx/v5"
	"time"
)

type Withdrawal struct {
	ID        int        `json:"-"`
	UserID    int        `json:"-"`
	Order     string     `json:"order"`
	Sum       float64    `json:"sum"`
	CreatedAt *time.Time `json:"processed_at"`
}

func CreateWithdrawalTable(ctx context.Context) error {
	_, err := database.Connection.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "withdrawal" (
			"id" SERIAL PRIMARY KEY,
			"user_id" INTEGER,
			"order" VARCHAR(128) NOT NULL DEFAULT 'NEW',
			"sum" FLOAT NOT NULL DEFAULT 0,
			"created_at" timestamp(3) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT user_withdrawal_id_fk FOREIGN KEY (user_id)
			REFERENCES "public"."user" (id) MATCH SIMPLE ON DELETE CASCADE ON UPDATE NO ACTION
		);
	`)

	return err
}

func Sum(ctx context.Context, userID int) (sum float64, err error) {
	err = database.Connection.QueryRow(ctx, `SELECT COALESCE(SUM(sum), 0) FROM "public"."withdrawal" WHERE user_id = $1;`, userID).Scan(&sum)
	return
}

func AllByUser(ctx context.Context, userID int) (withdrawals []Withdrawal, err error) {
	count := 0
	err = database.Connection.QueryRow(ctx, `SELECT COUNT(*) FROM "public"."withdrawal" WHERE user_id = $1;`, userID).Scan(&count)
	if err != nil {
		return
	}

	if count == 0 {
		withdrawals = []Withdrawal{}
		err = nil
		return
	}

	rows, err := database.Connection.Query(ctx, `SELECT ("id", "user_id", "order", "sum", "created_at")FROM "public"."withdrawal" WHERE "user_id" = $1 ORDER BY "created_at" DESC;`, userID)
	if err != nil {
		return
	}

	withdrawals = make([]Withdrawal, 0, count)
	for rows.Next() {
		var withdrawal Withdrawal

		err = scanWithdrawal(rows, &withdrawal)
		if err != nil {
			withdrawals = nil
			return
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	return
}

func scanWithdrawal(row pgx.Row, withdrawal *Withdrawal) error {
	return row.Scan(&withdrawal.ID, &withdrawal.UserID, &withdrawal.Order, &withdrawal.Sum, &withdrawal.CreatedAt)
}
