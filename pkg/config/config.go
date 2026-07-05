package pkgConfig

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
