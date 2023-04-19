package repository

import (
	"context"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/LorezV/go-diploma.git/internal/models"
	"github.com/jackc/pgx/v4"
)

type WithdrawalRepository struct {
	db *database.Database
}

func NewWithdrawalRepository(db *database.Database) *WithdrawalRepository {
	return &WithdrawalRepository{
		db: db,
	}
}

func (wr *WithdrawalRepository) Create(ctx context.Context, withdrawal *models.Withdrawal) error {
	sql := `INSERT INTO "public"."withdraw" (user_id, number, sum) VALUES ($1, $2, $3);`

	if _, err := wr.db.Exec(ctx, sql, withdrawal.UserID, withdrawal.Order, withdrawal.Sum); err != nil {
		return err
	}

	return nil
}

func (wr *WithdrawalRepository) Sum(ctx context.Context, user *models.User) (float64, error) {
	sql := `SELECT COALESCE(SUM(sum), 0) FROM "public"."withdraw" WHERE user_id = $1;`

	var sum float64
	if err := wr.db.QueryRow(ctx, sql, user.ID).Scan(&sum); err != nil {
		return 0, err
	}

	return sum, nil
}

func (wr *WithdrawalRepository) FindAll(ctx context.Context, user *models.User) ([]*models.Withdrawal, error) {
	count := 0
	if err := wr.db.QueryRow(ctx, `SELECT COUNT(*) FROM "public"."withdraw" WHERE user_id = $1;`, user.ID).Scan(&count); err != nil {
		return []*models.Withdrawal{}, err
	}
	if count == 0 {
		return []*models.Withdrawal{}, nil
	}

	sql := `SELECT id, user_id, number, sum, created_at FROM "public"."withdraw" WHERE user_id = $1 ORDER BY created_at DESC;`

	rows, err := wr.db.Query(ctx, sql, user.ID)
	if err != nil {
		return []*models.Withdrawal{}, err
	}

	collection, err := wr.scanWithdrawals(rows, count)
	if err != nil {
		return []*models.Withdrawal{}, err
	}

	return collection, nil
}

func (wr *WithdrawalRepository) scanWithdrawal(row pgx.Row, withdrawal *models.Withdrawal) error {
	if err := row.Scan(
		&withdrawal.ID,
		&withdrawal.UserID,
		&withdrawal.Order,
		&withdrawal.Sum,
		&withdrawal.CreatedAt,
	); err != nil {
		return err
	}

	return nil
}

func (wr *WithdrawalRepository) scanWithdrawals(rows pgx.Rows, count int) ([]*models.Withdrawal, error) {
	withdrawals := make([]*models.Withdrawal, 0, count)

	for rows.Next() {
		withdrawal := new(models.Withdrawal)

		if err := wr.scanWithdrawal(rows, withdrawal); err != nil {
			return []*models.Withdrawal{}, err
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	return withdrawals, nil
}
