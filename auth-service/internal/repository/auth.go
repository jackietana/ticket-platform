package repository

import (
	"context"
	"database/sql"

	"github.com/jackietana/ticket-platform/auth-service/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db}
}

func (r *Repository) CreateUser(ctx context.Context, usr domain.User) (string, error) {
	var id string
	strExec := "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id"

	if err := r.db.QueryRowContext(ctx, strExec, usr.Email, usr.Password).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (r *Repository) GetUser(ctx context.Context, email string) (domain.User, error) {
	var usr domain.User
	strExec := "SELECT id, email, password, created_at FROM users WHERE email=$1"

	err := r.db.QueryRowContext(ctx, strExec, email).
		Scan(&usr.ID, &usr.Email, &usr.Password, &usr.CreatedAt)
	if err != nil {
		return domain.User{}, err
	}

	return usr, nil
}
