package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	App struct {
		Debug            bool  `yaml:"debug"`
		BookmarksPerPage int64 `yaml:"bookmarks_per_page"`
	} `yaml:"app"`
	Server struct {
		Address string `yaml:"address"`
		BaseURL string `yaml:"base_url"`
	} `yaml:"server"`
	DB struct {
		Connection string `yaml:"connection"`
		Type       string `yaml:"type"`
	} `yaml:"db"`
	Storage struct {
		Type      string `yaml:"type"`
		InitParam string `yaml:"initParam"`
	} `yaml:"storage"`
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
