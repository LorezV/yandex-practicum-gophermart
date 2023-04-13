package models

import "time"

type Withdrawal struct {
	ID        int
	UserID    int
	Order     string
	Sum       float64
	CreatedAt *time.Time
}
