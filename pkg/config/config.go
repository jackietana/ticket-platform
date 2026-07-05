package pkgConfig

import (
	"fmt"
)

type ServerConfig struct {
	GRPCPort string `yaml:"grpc_port"`
	RESTPort string `yaml:"rest_port"`
}

type PostgresConfig struct {
	Host    string `yaml:"host"`
	Port    string `yaml:"port"`
	Name    string `yaml:"dbname"`
	User    string `yaml:"dbuser"`
	Pass    string `yaml:"dbpass"`
	SSLMode string `yaml:"sslmode"`
}

func (c *PostgresConfig) GetDatabaseConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Name, c.Pass, c.SSLMode)
}
