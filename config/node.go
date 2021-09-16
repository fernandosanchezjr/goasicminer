package config

type Node struct {
	URL        string `yaml:"url"`
	User       string `yaml:"user"`
	Pass       string `yaml:"pass"`
	Wallet     string `yaml:"wallet"`
	ClientOnly bool   `yaml:"clientOnly"`
}
