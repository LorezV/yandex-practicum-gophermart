package services

import (
	"context"
	"github.com/LorezV/go-diploma.git/internal/accural"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/LorezV/go-diploma.git/internal/models"
	"github.com/LorezV/go-diploma.git/internal/repository"
)

type User interface {
	Create(ctx context.Context, login, password string) (*models.User, error)
	FindByLogin(ctx context.Context, login string) (*models.User, error)
}

type Auth interface {
	Login(ctx context.Context, login, password string) (string, error)
	GetSecret() string
	GenerateToken(user *models.User) (string, error)
}

type Order interface {
	Create(ctx context.Context, number string, user *models.User) (*models.Order, error)
	FindByNumber(ctx context.Context, number string) (*models.Order, error)
	FindAll(ctx context.Context, user *models.User) ([]*models.Order, error)
	RunPolling(ctx context.Context) error
}

type Withdrawal interface {
	Create(ctx context.Context, withdrawal *models.Withdrawal, user *models.User) error
	Sum(ctx context.Context, user *models.User) (float64, error)
	All(ctx context.Context, user *models.User) ([]*models.Withdrawal, error)
}

type Services struct {
	User
	Auth
	Order
	Withdrawal
}

func MakeServices(repo *repository.Repository, client *clients.AccrualClient, JWTSecret string, db *database.Database) *Services {
	return &Services{
		User:       MakeUserService(repo.Users),
		Auth:       MakeAuthService(repo.Users, JWTSecret),
		Order:      MakeOrderService(repo.Orders, repo.Users, client),
		Withdrawal: NewWithdrawalService(repo.Withdrawals, repo.Users, db),
	}
}
