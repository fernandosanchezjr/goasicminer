package config

type DownTime struct {
	Start string `yaml:"start,omitempty"`
	End   string `yaml:"end,omitempty"`
}
