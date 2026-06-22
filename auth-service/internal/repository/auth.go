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

func (r *Repository) CreateUser(ctx context.Context, usr domain.User) error {
	strExec := "INSERT INTO users (email, password) VALUES ($1, $2)"
	if _, err := r.db.ExecContext(ctx, strExec, usr.Email, usr.Password); err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetUser(ctx context.Context, inp domain.UserInput) (domain.User, error) {
	var usr domain.User
	strExec := "SELECT id, email, password, created_at FROM users WHERE email=$1 AND password=$2"
	err := r.db.QueryRowContext(ctx, strExec, inp.Email, inp.Password).
		Scan(&usr.ID, &usr.Email, &usr.Password, &usr.CreatedAt)
	if err != nil {
		return domain.User{}, err
	}

	return usr, nil
}
