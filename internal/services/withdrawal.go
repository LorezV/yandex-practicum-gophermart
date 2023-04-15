package services

import (
	"context"
	"errors"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/LorezV/go-diploma.git/internal/models"
	"github.com/LorezV/go-diploma.git/internal/repository"
)

type WithdrawalService struct {
	repo  repository.Withdrawals
	users repository.Users
	db    *database.Database
}

func NewWithdrawalService(repo repository.Withdrawals, users repository.Users, db *database.Database) *WithdrawalService {
	return &WithdrawalService{
		repo:  repo,
		users: users,
		db:    db,
	}
}

func (ws *WithdrawalService) Create(ctx context.Context, withdrawal *models.Withdrawal, user *models.User) error {
	tx, err := ws.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO "public"."withdraw" (user_id, number, sum) VALUES ($1, $2, $3);`, withdrawal.UserID, withdrawal.Order, withdrawal.Sum)
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `UPDATE "public"."user" SET balance = "user".balance - $1 WHERE id = $2 AND balance >= $1;`, user.Balance, user.ID)
	if err != nil {
		return err
	}
	if !(tag.RowsAffected() > 0) {
		return errors.New("insufficient funds")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	updatedUser, err := ws.users.FindByID(ctx, user.ID)
	if err != nil {
		return err
	}

	user.Balance = updatedUser.Balance

	return nil
}

func (ws *WithdrawalService) Sum(ctx context.Context, user *models.User) (float64, error) {
	return ws.repo.Sum(ctx, user)
}

func (ws *WithdrawalService) All(ctx context.Context, user *models.User) ([]*models.Withdrawal, error) {
	return ws.repo.FindAll(ctx, user)
}
