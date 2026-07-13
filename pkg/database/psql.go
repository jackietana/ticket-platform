package pkgpsql

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	pkgconfig "github.com/jackietana/ticket-platform/pkg/config"
	_ "github.com/lib/pq"
)

var (
	ErrNoNewMigrations = errors.New("no change")
)

func NewPostgresConnection(cfg *pkgconfig.PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetDatabaseConnString())
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func RunUpMigrations(cfg *pkgconfig.PostgresConfig, service string) error {
	migrationsPath := os.Getenv("APP_MIGRATIONS_PATH")
	if migrationsPath == "" {
		_, b, _, _ := runtime.Caller(0)
		migrationsPath = filepath.Join(filepath.Dir(b), "../../../initdb/migrations/", service)
	}
	migrationDir := "file://" + migrationsPath

	db, err := sql.Open("postgres", cfg.GetDatabaseConnString())
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}
	defer driver.Close()

	migrateInst, err := migrate.NewWithDatabaseInstance(migrationDir, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := migrateInst.Up(); err != nil {
		if errors.Is(err, ErrNoNewMigrations) {
			return fmt.Errorf("failed to up db: %w", err)
		}
	}

	migrateInst.Close()
	return nil
}
