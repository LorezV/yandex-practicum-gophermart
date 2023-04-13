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
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type accrualClient struct {
	client  *resty.Client
	retryAt *time.Time
}

var AccrualSystemUnavailableError = errors.New("accrual system is unavailable")
var AccrualSystemNoContentError = errors.New("no order in accural")

func MakeAccrualClient() accrualClient {
	return accrualClient{
		client: resty.New().SetBaseURL(config.Config.AccrualSystemAddress),
	}
}

func (ac *accrualClient) FetchOrder(ctx context.Context, number string) (order *AccrualOrder, err error) {
	order = new(AccrualOrder)

	resp, err := ac.client.R().
		SetContext(ctx).
		SetResult(order).
		Get(fmt.Sprintf("/api/orders/%s", number))
	if err != nil {
		return
	}

	if err = ac.isBlocked(resp); err != nil {
		return
	}

	if resp.StatusCode() == http.StatusNoContent {
		err = AccrualSystemNoContentError
		return
	}

	return
}

func (ac *accrualClient) CanRequest() bool {
	if ac.retryAt == nil {
		return true
	}

	if time.Now().After(*ac.retryAt) {
		ac.retryAt = nil
		return true
	}

	return false
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
