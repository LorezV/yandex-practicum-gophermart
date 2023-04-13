package repository

import (
	"context"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/LorezV/go-diploma.git/internal/models"
)

type Users interface {
	Create(ctx context.Context, login, hashedPassword string) error
	Update(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id int) (*models.User, error)
	FindByLogin(ctx context.Context, login string) (*models.User, error)
}

type Orders interface {
	Create(ctx context.Context, number string, user *models.User) error
	Update(ctx context.Context, order *models.Order) error
	FindByNumber(ctx context.Context, number string) (*models.Order, error)
	FindAll(ctx context.Context, user *models.User) ([]*models.Order, error)
	FindPending(ctx context.Context) ([]*models.Order, error)
}

type Withdrawals interface {
	Create(ctx context.Context, withdrawal *models.Withdrawal) error
	Sum(ctx context.Context, user *models.User) (float64, error)
	FindAll(ctx context.Context, user *models.User) ([]*models.Withdrawal, error)
}

type Repository struct {
	Users
	Orders
	Withdrawals
}

func MakeRepository(db *database.Database) *Repository {
	return &Repository{
		Users:       NewUserRepository(db),
		Orders:      MakeOrderRepository(db),
		Withdrawals: NewWithdrawalRepository(db),
	}
}
