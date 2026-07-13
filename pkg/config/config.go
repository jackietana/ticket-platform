package pkgconfig

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

type RedisConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Pass string `yaml:"pass"`
}

type MinioConfig struct {
	InternalEndpoint string `yaml:"internal_endpoint"`
	ExternalEndpoint string `yaml:"external_endpoint"`
	ServerPort       string `yaml:"server_port"`
	User             string `yaml:"minio_user"`
	Pass             string `yaml:"minio_pass"`
}

type AuthServiceConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (c *PostgresConfig) GetDatabaseConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Name, c.Pass, c.SSLMode)
}

func (c *AuthServiceConfig) GetAuthServiceEndpoint() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
