package models

import "time"

const (
	NewOrderStatus        = "NEW"
	ProcessingOrderStatus = "PROCESSING"
	InvalidOrderStatus    = "INVALID"
	ProcessedOrderStatus  = "PROCESSED"
)

type Order struct {
	ID        int
	UserID    int
	Number    string
	Status    string
	Accrual   *float64
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
