package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password" binding:"required,gte=6"`
	CreatedAt time.Time `json:"created_at"`
}

type UserInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,gte=6"`
}

type Session struct {
	Token  string    `json:"token"`
	UserId uuid.UUID `json:"user_id"`
}
