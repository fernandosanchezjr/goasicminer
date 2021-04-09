package config

type PowerControl struct {
	Enabled bool `yaml:"enabled,omitempty"`
	Pin     int  `yaml:"pin,omitempty"`
	High    bool `yaml:"high,omitempty"`
}
