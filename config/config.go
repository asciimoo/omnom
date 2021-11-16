package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	App struct {
		Debug bool `yaml:"debug"`
	} `yaml:"app"`
	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
	DB struct {
		Connection string `yaml:"connection"`
		Type       string `yaml:"type"`
	} `yaml:"db"`
}

func Load(filename string) (*Config, error) {
	var c *Config
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(b, &c)
	// TODO validate config
	return c, err
}
