package models

import (
	"time"

	"gorm.io/gorm"
)

type Expense struct {
	gorm.Model
	Title    string    `json:"title"`
	Amount   float64   `json:"amount"`
	Category string    `json:"category"`
	Date     time.Time `json:"date"`
	UserID   uint      `json:"user_id"` // Foreign Key
}
