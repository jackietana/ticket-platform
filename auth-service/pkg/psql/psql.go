package psql

import (
	"database/sql"

	"github.com/jackietana/ticket-platform/auth-service/internal/config"
	_ "github.com/lib/pq"
)

func NewPostgresConnection(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetDatabaseConnString())
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
