package config

import (
	"flag"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
)

var configPath string

func init() {
	configFolder := utils.GetHomeFolder()
	defaultConfigPath := path.Join(configFolder, "config.yaml")
	flag.StringVar(&configPath, "config", defaultConfigPath, "specify config file")
}

func LoadConfig() (*Config, error) {
	c := &Config{}
	var data []byte
	var err error
	log.WithFields(log.Fields{
		"path": configPath,
	}).Println("Loading config")
	if data, err = ioutil.ReadFile(configPath); err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}
