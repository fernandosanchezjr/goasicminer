package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
)

var configPath string

func init() {
	configFolder := getOrCreateConfigFolder()
	defaultConfigPath := path.Join(configFolder, "config.yaml")
	flag.StringVar(&configPath, "config", defaultConfigPath, "specify config file")
}

func getOrCreateConfigFolder() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Println("could not find home folder")
		return ""
	}
	configFolder := path.Join(home, ".goasicminer")
	if err := os.MkdirAll(configFolder, 0700); err != nil {
		log.Println("Could not create", configFolder)
		return ""
	}
	return configFolder
}

func LoadConfig() (*Config, error) {
	c := &Config{}
	var data []byte
	var err error
	log.Println("Loading config", configPath)
	if data, err = ioutil.ReadFile(configPath); err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}
