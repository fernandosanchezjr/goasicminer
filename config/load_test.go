package config

import "testing"

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(config)
}
