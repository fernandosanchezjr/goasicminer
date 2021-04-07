package config

type Config struct {
	Pools          []Pool     `yaml:"pools"`
	BackendAddress string     `yaml:"backend,omitempty"`
	ServerAddress  string     `yaml:"server,omitempty"`
	R606           []R606     `yaml:"r606,omitempty"`
	DownTime       []DownTime `yaml:"downtime,omitempty"`
}
