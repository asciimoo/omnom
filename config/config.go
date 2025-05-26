// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/asciimoo/omnom/oauth"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	fname       string
	App         App          `yaml:"app"`
	Server      Server       `yaml:"server"`
	DB          DB           `yaml:"db"`
	Storage     Storage      `yaml:"storage"`
	SMTP        SMTP         `yaml:"smtp"`
	ActivityPub *ActivityPub `yaml:"activitypub"`
	OAuth       OAuth        `yaml:"oauth"`
}

type App struct {
	LogLevel                 string `yaml:"log_level"`
	ResultsPerPage           int64  `yaml:"results_per_page"`
	DisableSignup            bool   `yaml:"disable_signup"`
	StaticDir                string `yaml:"static_dir"`
	CreateBookmarkFromWebapp bool   `yaml:"create_bookmark_from_webapp"`
	WebappSnapshotterTimeout int    `yaml:"webapp_snapshotter_timeout"`
	DebugSQL                 bool   `yaml:"debug_sql"`
}

type Server struct {
	Address          string `yaml:"address"`
	BaseURL          string `yaml:"base_url"`
	SecureCookie     bool   `yaml:"secure_cookie"`
	RemoteUserHeader string `yaml:"remote_user_header"`
}

type DB struct {
	Connection string `yaml:"connection"`
	Type       string `yaml:"type"`
}

type Storage struct {
	Type string `yaml:"type"`
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

type ActivityPub struct {
	PubKeyPath  string `yaml:"pubkey"`
	PrivKeyPath string `yaml:"privkey"`
	PubK        *rsa.PublicKey
	PrivK       *rsa.PrivateKey
}

type OAuth map[string]OAuthEntry

type OAuthEntry struct {
	ClientID         string   `yaml:"client_id"`
	ClientSecret     string   `yaml:"client_secret"`
	ConfigurationURL string   `yaml:"configuration_url"`
	AuthURL          string   `yaml:"auth_url"`
	TokenURL         string   `yaml:"token_url"`
	Icon             string   `yaml:"icon"`
	Scopes           []string `yaml:"scopes"`
}

func readConfigFile(filename string) ([]byte, string, error) {
	// try config file provided by the user
	b, err := os.ReadFile(filename)
	if err == nil {
		return b, filename, nil
	}
	// try $HOME/.omnomrc
	homeDir, err := os.UserHomeDir()
	if err != nil {
		filename = filepath.Join(homeDir, ".omnomrc")
		b, err = os.ReadFile(filename)
		if err == nil {
			return b, filename, nil
		}
		filename = filepath.Join(homeDir, ".config/omnom/config.yml")
		// try $HOME/.config/omnom/config.yml
		b, err = os.ReadFile(filename)
		if err == nil {
			return b, filename, nil
		}
	}
	return b, "", errors.New("configuration file not found. Use --config to specify a custom config file")
}

func Load(filename string) (*Config, error) {
	b, fn, err := readConfigFile(filename)
	if err != nil {
		log.Debug().Msg("No config file found, using default config")
		//lint:ignore nilerr // no need to check error
		return CreateDefaultConfig(), nil
	}
	c, err := parseConfig(b)
	c.fname = fn
	return c, err
}

func CreateDefaultConfig() *Config {
	return &Config{
		App: App{
			ResultsPerPage:           30,
			CreateBookmarkFromWebapp: false,
			WebappSnapshotterTimeout: 15,
			LogLevel:                 "info",
			StaticDir:                "./static",
		},
		Server: Server{
			Address:      "127.0.0.1:7331",
			SecureCookie: false,
		},
		DB: DB{
			Type:       "sqlite",
			Connection: "db.sqlite3",
		},
		Storage: Storage{
			Type: "fs",
		},
		ActivityPub: &ActivityPub{
			PubKeyPath:  "./public.pem",
			PrivKeyPath: "./private.pem",
		},
		SMTP: SMTP{
			Host:              "",
			Port:              25,
			Username:          "",
			Password:          "",
			Sender:            "Omnom <omnom@127.0.0.1>",
			TLS:               false,
			TLSAllowInsecure:  false,
			SendTimeout:       10,
			ConnectionTimeout: 5,
		},
	}
}

func parseConfig(rawConfig []byte) (*Config, error) {
	c := CreateDefaultConfig()
	err := yaml.Unmarshal(rawConfig, &c)
	if err != nil {
		return nil, err
	}
	if c.Server.RemoteUserHeader != "" {
		if len(c.OAuth) > 0 {
			return nil, errors.New("can't specify OAuth providers when remote user header is enabled")
		}
		if c.App.DisableSignup == false {
			return nil, errors.New("user signups must be disabled when remote user header is enabled")
		}
	}
	for pn, _ := range c.OAuth {
		if _, ok := oauth.Providers[pn]; !ok {
			return nil, errors.New("invalid oauth provider: " + pn)
		}
	}
	if strings.HasSuffix(c.Server.BaseURL, "/") {
		c.Server.BaseURL = c.Server.BaseURL[:len(c.Server.BaseURL)-1]
	}
	return c, nil
}

func (c *Config) Filename() string {
	if c.fname == "" {
		return "*Default Config*"
	}
	return c.fname
}

func (ap *ActivityPub) ExportPrivKey() ([]byte, error) {
	if ap.PrivK == nil {
		var err error
		ap.PrivK, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}
		ap.PubK = &ap.PrivK.PublicKey
	}
	privkeyBytes := x509.MarshalPKCS1PrivateKey(ap.PrivK)
	privkeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privkeyBytes,
		},
	)
	return privkeyPem, nil
}

func (ap *ActivityPub) ParsePrivKey(privPEM []byte) error {
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return errors.New("failed to parse PEM block containing the key")
	}

	var err error
	ap.PrivK, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	return err
}

func (ap *ActivityPub) ExportPubKey() ([]byte, error) {
	pubkeyBytes, err := x509.MarshalPKIXPublicKey(ap.PubK)
	if err != nil {
		return []byte{}, err
	}
	pubkeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubkeyBytes,
		},
	)

	return pubkeyPem, nil
}

func (ap *ActivityPub) ParsePubKey(pubPEM []byte) error {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		ap.PubK = pub
		return nil
	default:
		break // fall through
	}
	return errors.New("key type is not RSA")
}
