// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	App     App     `yaml:"app"`
	Server  Server  `yaml:"server"`
	DB      DB      `yaml:"db"`
	Storage Storage `yaml:"storage"`
	SMTP    SMTP    `yaml:"smtp"`
	OAuth   OAuth   `yaml:"oauth"`
}

type App struct {
	Debug          bool   `yaml:"debug"`
	ResultsPerPage int64  `yaml:"results_per_page"`
	DisableSignup  bool   `yaml:"disable_signup"`
	TemplateDir    string `yaml:"template_dir"`
	StaticDir      string `yaml:"static_dir"`
}

type Server struct {
	Address      string `yaml:"address"`
	BaseURL      string `yaml:"base_url"`
	SecureCookie bool   `yaml:"secure_cookie"`
}

type DB struct {
	Connection string `yaml:"connection"`
	Type       string `yaml:"type"`
}

type Storage struct {
	Type string `yaml:"type"`
	Root string `yaml:"root"`
}

type SMTP struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	Sender            string `yaml:"sender"`
	TLS               bool   `yaml:"tls"`
	TLSAllowInsecure  bool   `yaml:"tls_allow_insecure"`
	SendTimeout       int    `yaml:"send_timeout"`
	ConnectionTimeout int    `yaml:"connection_timeout"`
}

type OAuth map[string]OAuthEntry

type OAuthEntry struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	AuthURL      string   `yaml:"auth_url"`
	TokenURL     string   `yaml:"token_url"`
	Icon         string   `yaml:"icon"`
	Scopes       []string `yaml:"scopes"`
}

func readConfigFile(filename string) ([]byte, error) {
	// try config file provided by the user
	b, err := os.ReadFile(filename)
	if err == nil {
		return b, nil
	}
	// try $HOME/.omnomrc
	homeDir, err := os.UserHomeDir()
	if err != nil {
		b, err = os.ReadFile(filepath.Join(homeDir, ".omnomrc"))
		if err == nil {
			return b, nil
		}
	}
	// try $HOME/.config/omnom/config.yml
	b, err = os.ReadFile(filepath.Join(homeDir, ".config/omnom/config.yml"))
	if err == nil {
		return b, nil
	}
	return b, errors.New("Configuration file not found. Use --config to specify a custom config file")
}

func Load(filename string) (*Config, error) {
	var c *Config
	b, err := readConfigFile(filename)
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(b, &c)
	// TODO validate config
	return c, err
}
