package services

import (
	"context"
	"errors"
	"github.com/LorezV/go-diploma.git/internal/accural"
	"github.com/LorezV/go-diploma.git/internal/models"
	"github.com/LorezV/go-diploma.git/internal/repository"
	"time"
)

type OrderService struct {
	repo   repository.Orders
	users  repository.Users
	client *clients.AccrualClient
}

func MakeOrderService(repo repository.Orders, users repository.Users, client *clients.AccrualClient) *OrderService {
	return &OrderService{
		repo:   repo,
		users:  users,
		client: client,
	}
}

func (os *OrderService) Create(ctx context.Context, number string, user *models.User) (*models.Order, error) {
	if err := os.repo.Create(ctx, number, user); err != nil {
		return nil, err
	}

	order, err := os.FindByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	if _, err = os.PollStatus(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (os *OrderService) FindByNumber(ctx context.Context, number string) (*models.Order, error) {
	return os.repo.FindByNumber(ctx, number)
}

func (os *OrderService) FindAll(ctx context.Context, user *models.User) ([]*models.Order, error) {
	return os.repo.FindAll(ctx, user)
}

func (os *OrderService) RunPolling(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := os.pollStatuses(ctx); err != nil && !errors.Is(err, clients.ErrAccrualSystemUnavailable) {
				return err
			}
		}
	}
}

func (os *OrderService) PollStatus(ctx context.Context, order *models.Order) (bool, error) {
	resp, err := os.client.FetchOrder(ctx, order.Number)
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

	if err = os.repo.Update(ctx, order); err != nil {
		return false, err
	}

	user, err := os.users.FindByID(ctx, order.UserID)
	if err != nil {
		return false, err
	}
	user.Balance += *order.Accrual

	if err = os.users.Update(ctx, user); err != nil {
		return false, err
	}

	return true, nil
}

func (os *OrderService) pollStatuses(ctx context.Context) error {
	if err := os.client.CanRequest(); err != nil {
		return nil
	}

	orders, err := os.repo.FindPending(ctx)
	if err != nil {
		return err
	}

	for _, order := range orders {
		_, err := os.PollStatus(ctx, order)
		if err != nil {
			return err
		}
	}

	return nil
}
