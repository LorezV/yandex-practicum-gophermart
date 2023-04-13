package accural

import (
	"context"
	"errors"
	"fmt"
	"github.com/LorezV/go-diploma.git/internal/config"
	"github.com/go-resty/resty/v2"
	"net/http"
	"time"
)

var AccrualClient accrualClient

type AccrualOrder struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual"`
}

type accrualClient struct {
	client  *resty.Client
	retryAt *time.Time
}

var ErrAccrualSystemUnavailable = errors.New("accrual system is unavailable")

func InitAccrualClient() {
	AccrualClient = accrualClient{
		client: resty.New().SetBaseURL(config.Config.AccrualSystemAddress),
	}
}

func (ac *accrualClient) FetchOrder(ctx context.Context, number string) (*AccrualOrder, error) {
	order := new(AccrualOrder)

	resp, err := ac.client.R().
		SetContext(ctx).
		SetResult(order).
		Get(fmt.Sprintf("/api/orders/%s", number))
	if err != nil {
		return nil, ErrAccrualSystemUnavailable
	}

	if err = ac.isBlocked(resp); err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode() != 200 {
		return nil, ErrAccrualSystemUnavailable
	}

	return order, nil
}

func (ac *accrualClient) CanRequest() error {
	if ac.retryAt == nil {
		return nil
	}

	if time.Now().After(*ac.retryAt) {
		ac.retryAt = nil
		return nil
	}

	return fmt.Errorf("accrual client unlocks in %s", time.Until(*ac.retryAt))
}

func (ac *accrualClient) isBlocked(resp *resty.Response) error {
	if !resp.IsError() && resp.StatusCode() != http.StatusTooManyRequests {
		return nil
	}

	delay, err := time.ParseDuration(fmt.Sprintf("%ss", resp.Header().Get("Retry-After")))
	if err != nil {
		return nil
	}

	body := string(resp.Body())
	fmt.Println(body)
	*ac.retryAt = time.Now().Add(delay)

	return fmt.Errorf(body)
}
