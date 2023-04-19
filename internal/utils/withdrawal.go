package utils

import (
	"github.com/LorezV/go-diploma.git/internal/models"
	"time"
)

type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func MakeWithdrawalResponse(withdrawal *models.Withdrawal) WithdrawalResponse {
	return WithdrawalResponse{
		Order:       withdrawal.Order,
		Sum:         withdrawal.Sum,
		ProcessedAt: withdrawal.CreatedAt.Format(time.RFC3339),
	}
}

func MakeWithdrawalsResponse(withdrawals []*models.Withdrawal) (result []WithdrawalResponse) {
	for _, withdrawal := range withdrawals {
		result = append(result, MakeWithdrawalResponse(withdrawal))
	}
	return result
}
