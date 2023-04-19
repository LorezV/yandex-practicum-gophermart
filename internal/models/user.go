package models

import "time"

type User struct {
	ID        int
	Login     string
	Password  string
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}
