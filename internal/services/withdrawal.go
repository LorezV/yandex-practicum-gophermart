package services

import (
	"context"
	"github.com/LorezV/go-diploma.git/internal/models"
	"github.com/LorezV/go-diploma.git/internal/repository"
)

type WithdrawalService struct {
	repo  repository.Withdrawals
	users repository.Users
}

func NewWithdrawalService(repo repository.Withdrawals, users repository.Users) *WithdrawalService {
	return &WithdrawalService{
		repo:  repo,
		users: users,
	}
}

func (ws *WithdrawalService) Create(ctx context.Context, withdrawal *models.Withdrawal, user *models.User) error {
	if err := ws.repo.Create(ctx, withdrawal); err != nil {
		return err
	}

	user.Balance -= withdrawal.Sum
	if err := ws.users.Update(ctx, user); err != nil {
		return err
	}

	return nil
}

func (ws *WithdrawalService) Sum(ctx context.Context, user *models.User) (float64, error) {
	return ws.repo.Sum(ctx, user)
}

func (ws *WithdrawalService) All(ctx context.Context, user *models.User) ([]*models.Withdrawal, error) {
	return ws.repo.FindAll(ctx, user)
}
