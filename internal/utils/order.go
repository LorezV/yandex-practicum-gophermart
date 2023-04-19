package utils

import (
	"github.com/LorezV/go-diploma.git/internal/models"
	"time"
)

type OrderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

func MakeOrderResponse(order *models.Order) OrderResponse {
	return OrderResponse{
		Number:     order.Number,
		Status:     order.Status,
		Accrual:    order.Accrual,
		UploadedAt: order.CreatedAt.Format(time.RFC3339),
	}
}

func MakeOrdersResponse(orders []*models.Order) (result []OrderResponse) {
	for _, order := range orders {
		result = append(result, MakeOrderResponse(order))
	}
	return result
}
