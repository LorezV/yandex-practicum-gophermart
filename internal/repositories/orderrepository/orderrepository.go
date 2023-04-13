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
	Accrual   *float64
}

func CreateOrderTable(ctx context.Context) error {
	_, err := database.Connection.Exec(ctx, `
DROP TABLE "order";
	`)

	return err
}

func FindUnique(ctx context.Context, field string, value interface{}) (order *Order, err error) {
	order = new(Order)
	row := database.Connection.QueryRow(ctx, fmt.Sprintf("SELECT * FROM \"public\".\"order\" WHERE %s=$1", field), value)
	err = scan(row, order)

	fmt.Println(*order.Accrual)
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

		err = scan(rows, order)
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

		err = scan(rows, order)
		if err != nil {
			orders = nil
			return
		}

		orders = append(orders, order)
	}

	return
}

func Create(ctx context.Context, userID int, number string) (*Order, error) {
	_, err := database.Connection.Exec(ctx, `INSERT INTO "public"."order" (number, user_id) VALUES ($1, $2)`, number, userID)
	if err != nil {
		return nil, err
	}

	order, err := FindUnique(ctx, "number", number)
	if err != nil {
		return nil, err
	}

	if _, err = PollStatus(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func Update(ctx context.Context, order *Order) (err error) {
	_, err = database.Connection.Exec(ctx, `UPDATE "public"."order" SET status=$1, accrual=$2 WHERE id=$3`, order.Status, order.Accrual, order.ID)
	return
}

func PollStatus(ctx context.Context, order *Order) (bool, error) {
	resp, err := accural.AccrualClient.FetchOrder(ctx, order.Number)
	if err != nil {
		if errors.Is(err, accural.ErrAccrualSystemNoContent) {
			return false, nil
		}

		return false, err
	}

	if order.Status == resp.Status {
		return false, nil
	}

	order.Status = resp.Status
	order.Accrual = resp.Accrual

	if err = Update(ctx, order); err != nil {
		return false, err
	}

	user, err := userrepository.FindUnique(ctx, "id", order.UserID)
	if err != nil {
		return false, err
	}

	user.Balance += *order.Accrual
	if err = userrepository.Update(ctx, &user); err != nil {
		return false, err
	}

	return true, nil
}

func pollStatuses(ctx context.Context) error {
	if err := accural.AccrualClient.CanRequest(); err != nil {
		return nil
	}

	orders, err := FindPending(ctx)
	if err != nil {
		return err
	}

	for _, order := range orders {
		_, err = PollStatus(ctx, order)
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

func scan(row pgx.Row, order *Order) error {
	return row.Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.CreatedAt,
		&order.Status,
		&order.Accrual)
}
