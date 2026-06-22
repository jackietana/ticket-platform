package psql

import (
	"database/sql"
	"fmt"

	"github.com/jackietana/ticket-platform/auth-service/internal/config"
	_ "github.com/lib/pq"
)

func NewPostgresConnection(cfg config.Postgres) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Pass, cfg.SSLMode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
