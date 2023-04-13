package orderrepository

import (
	"context"
	"errors"
	"fmt"
	"github.com/LorezV/go-diploma.git/internal/accural"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/LorezV/go-diploma.git/internal/repositories/userrepository"
	"github.com/jackc/pgx/v5"
	"time"
)

const (
	NewOrderStatus        = "NEW"
	ProcessingOrderStatus = "PROCESSING"
	InvalidOrderStatus    = "INVALID"
	ProcessedOrderStatus  = "PROCESSED"
)

type Order struct {
	ID        int
	Number    string
	UserID    int
	CreatedAt *time.Time
	Status    string
	Accrual   float64
}

func CreateOrderTable(ctx context.Context) error {
	_, err := database.Connection.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "order" (
			"id" SERIAL PRIMARY KEY,
			"number" VARCHAR(256) NOT NULL UNIQUE,
			"user_id" INTEGER,
			"created_at" timestamp(3) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"status" VARCHAR(128) NOT NULL DEFAULT 'NEW',
			"accrual" FLOAT NOT NULL DEFAULT 0,
			CONSTRAINT user_order_id_fk FOREIGN KEY (user_id)
			REFERENCES "public"."user" (id) MATCH SIMPLE ON DELETE CASCADE ON UPDATE NO ACTION
		);
	`)

	return err
}

func FindUnique(ctx context.Context, field string, value interface{}) (order *Order, err error) {
	order = new(Order)
	row := database.Connection.QueryRow(ctx, fmt.Sprintf("SELECT * FROM \"public\".\"order\" WHERE %s=$1", field), value)
	err = scanOrder(row, order)
	return
}

func FindByUser(ctx context.Context, userID int) (orders []*Order, err error) {
	orders = nil
	rows, err := database.Connection.Query(ctx, `SELECT * FROM "public"."order" WHERE user_id=$1`, userID)
	if err != nil {
		return
	}

	err = rows.Err()
	if err != nil {
		return
	}

	defer rows.Close()

	orders = make([]*Order, 0)

	for rows.Next() {
		order := new(Order)

		err = scanOrder(rows, order)
		if err != nil {
			orders = nil
			return
		}

		orders = append(orders, order)
	}

	return
}

func FindPending(ctx context.Context) (orders []*Order, err error) {
	orders = nil
	rows, err := database.Connection.Query(ctx, `SELECT * FROM "public"."order" WHERE status IN ($1, $2)`, NewOrderStatus, ProcessingOrderStatus)
	if err != nil {
		return
	}

	err = rows.Err()
	if err != nil {
		return
	}

	defer rows.Close()

	orders = make([]*Order, 0)

	for rows.Next() {
		var order = new(Order)

		err = scanOrder(rows, order)
		if err != nil {
			orders = nil
			return
		}

		orders = append(orders, order)
	}

	return
}

func Create(ctx context.Context, userID int, number string) (order *Order, err error) {
	_, err = database.Connection.Exec(ctx, `INSERT INTO "public"."order" (number, user_id) VALUES ($1, $2)`, number, userID)
	if err != nil {
		return
	}

	order, err = FindUnique(ctx, "number", number)
	if err != nil {
		return
	}

	if _, err = PollStatus(ctx, order); err != nil {
		return
	}
	return
}

func Update(ctx context.Context, order *Order) (err error) {
	_, err = database.Connection.Exec(ctx, `UPDATE "public"."order" SET status=$1, accrual=$2 WHERE id=$3`, order.Status, order.Accrual, order.ID)
	return
}

func PollStatus(ctx context.Context, order *Order) (bool, error) {
	resp, err := accural.AccrualClient.FetchOrder(ctx, order.Number)
	if err != nil {
		return false, err
	}

	if resp == nil {
		return false, nil
	}

	if order.Status == resp.Status {
		return false, nil
	}

	order.Status = resp.Status
	order.Accrual = resp.Accrual

	if err = Update(ctx, order); err != nil {
		return false, err
	}

	user, err := userrepository.Get(ctx, "id", order.UserID)
	if err != nil {
		return false, nil
	}

	user.Balance += order.Accrual
	if err := userrepository.Update(ctx, &user); err != nil {
		return false, err
	}

	return true, nil
}

func pollStatuses(ctx context.Context) error {
	if ok := accural.AccrualClient.CanRequest(); !ok {
		return nil
	}

	orders, err := FindPending(ctx)
	if err != nil {
		return err
	}

	for _, order := range orders {
		_, err := PollStatus(ctx, order)
		if err != nil {
			return err
		}
	}

	return nil
}

func RunPollingStatuses(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := pollStatuses(ctx); err != nil && !errors.Is(err, accural.ErrAccrualSystemUnavailable) {
				return err
			}

		}
	}
}

func scanOrder(row pgx.Row, order *Order) error {
	return row.Scan(&order.ID, &order.Number, &order.UserID, &order.CreatedAt, &order.Status, &order.Accrual)
}
