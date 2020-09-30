package config

type Config struct {
	Pools          []Pool `yaml:"pools"`
	BackendAddress string `yaml:"backend,omitempty"`
	ServerAddress  string `yaml:"server,omitempty"`
}
