package pkgConfig

import (
	"fmt"
	"os"
)

type ServerConfig struct {
	GRPCPort string `yaml:"grpc_port"`
	RESTPort string `yaml:"rest_port"`
}

type PostgresConfig struct {
	Host    string
	Port    string `yaml:"port"`
	Name    string
	User    string
	Pass    string
	SSLMode string
}

func NewPostgresConfig() PostgresConfig {
	cfg := PostgresConfig{}

	cfg.Host = os.Getenv("APP_DB_HOST")
	cfg.Name = os.Getenv("APP_DB_NAME")
	cfg.User = os.Getenv("APP_DB_USER")
	cfg.Pass = os.Getenv("APP_DB_PASS")
	cfg.SSLMode = os.Getenv("APP_DB_SSLMODE")

	return cfg
}

func (c *PostgresConfig) GetDatabaseConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Name, c.Pass, c.SSLMode)
}
