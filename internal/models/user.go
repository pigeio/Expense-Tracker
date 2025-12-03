package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username   string `gorm:"unique" json:"username"`
	Password   string `json:"password"`
	IsPremium  bool   `json:"is_premium" gorm:"default:false"` // <--- This is required
	RazorpayID string `json:"-"`
}
