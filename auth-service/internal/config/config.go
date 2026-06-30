package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	DB   Postgres
	Salt string
}

type Postgres struct {
	Host    string
	Port    string
	Name    string
	User    string
	Pass    string
	SSLMode string
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	cfg.DB.Host = os.Getenv("APP_DB_HOST")
	cfg.DB.Port = os.Getenv("APP_DB_PORT")
	cfg.DB.Name = os.Getenv("APP_DB_NAME")
	cfg.DB.User = os.Getenv("APP_DB_USER")
	cfg.DB.Pass = os.Getenv("APP_DB_PASS")
	cfg.DB.SSLMode = os.Getenv("APP_DB_SSLMODE")

	cfg.Salt = os.Getenv("APP_HASH_SALT")

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func NewTestConfig() *Config {
	return &Config{
		DB: Postgres{Host: "localhost", Name: "test_db", User: "test_user",
			Pass: "test_pass", Port: "5432", SSLMode: "disable"},
		Salt: "test_salt_0123456789",
	}
}

func (c *Config) GetDatabaseConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		c.DB.Host, c.DB.Port, c.DB.User, c.DB.Name, c.DB.Pass, c.DB.SSLMode)
}

func (c *Config) validate() error {
	if c.DB.Host == "" || c.DB.Port == "" || c.DB.Name == "" ||
		c.DB.User == "" || c.DB.Pass == "" || c.DB.SSLMode == "" {
		return errors.New("missing db configuration")
	}

	if c.Salt == "" {
		return errors.New("missing hash salt")
	}

	return nil
}
