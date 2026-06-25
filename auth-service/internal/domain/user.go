package domain

import (
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,gte=6"`
}

type Session struct {
	Token     string `json:"token"`
	UserID    string `json:"user_id"`
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
}
