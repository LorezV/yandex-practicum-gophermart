package clients

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

var ErrAccrualSystemUnavailable = errors.New("accrual system is unavailable")

type Order struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual"`
}

type AccrualClient struct {
	client  *resty.Client
	retryAt *time.Time
}

func MakeAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		client: resty.New().SetBaseURL(baseURL),
	}
}

func (asc *AccrualClient) FetchOrder(ctx context.Context, number string) (*Order, error) {
	order := new(Order)

	resp, err := asc.client.R().
		SetContext(ctx).
		SetResult(order).
		Get(fmt.Sprintf("/api/orders/%s", number))
	if err != nil {
		return nil, ErrAccrualSystemUnavailable
	}
	if err = asc.isBlocked(resp); err != nil {
		return nil, err
	}

	log.Debug().
		Str("status", resp.Status()).
		Str("msg", string(resp.Body())).
		Str("order", number).
		Msg("Polling order status")

	if resp.StatusCode() == http.StatusNoContent {
		return nil, nil
	}

	return order, nil
}

func (asc *AccrualClient) CanRequest() error {
	if asc.retryAt == nil {
		return nil
	}

	if time.Now().After(*asc.retryAt) {
		asc.retryAt = nil
		return nil
	}

	return fmt.Errorf("accural unblocks in %s", time.Until(*asc.retryAt))
}

func (asc *AccrualClient) isBlocked(resp *resty.Response) error {
	if !resp.IsError() && resp.StatusCode() != http.StatusTooManyRequests {
		return nil
	}

	delay, err := time.ParseDuration(fmt.Sprintf("%ss", resp.Header().Get("Retry-After")))
	if err != nil {
		return err
	}

	body := string(resp.Body())
	*asc.retryAt = time.Now().Add(delay)

	return fmt.Errorf(body)
}
