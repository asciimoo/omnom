// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

// Package config provides configuration management for the Omnom application.
//
// This package handles loading, parsing, and validating application configuration
// from YAML files. It supports configuration for various components including:
//   - Application settings (logging, pagination, snapshots)
//   - Server settings (address, base URL, cookies)
//   - Database configuration (type and connection parameters)
//   - Storage backends (filesystem, future cloud storage)
//   - SMTP email settings
//   - ActivityPub federation (key management)
//   - OAuth provider configuration
//
// The configuration can be loaded from multiple locations in order of precedence:
//  1. Path specified via --config flag
//  2. ./config.yml in the current directory
//  3. ~/.omnomrc in the user's home directory
//  4. ~/.config/omnom/config.yml in the user's config directory
//
// Example usage:
//
//	cfg, err := config.Load("config.yml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Server address:", cfg.Server.Address)
package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/asciimoo/omnom/oauth"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	fname       string
	App         App          `yaml:"app"`
	Server      Server       `yaml:"server"`
	DB          DB           `yaml:"db"`
	Feed        Feed         `yaml:"feed"`
	Storage     Storage      `yaml:"storage"`
	SMTP        SMTP         `yaml:"smtp"`
	ActivityPub *ActivityPub `yaml:"activitypub"`
	OAuth       OAuth        `yaml:"oauth"`
}

// App holds application-specific settings.
type App struct {
	LogLevel                 string `yaml:"log_level"`
	ResultsPerPage           uint   `yaml:"results_per_page"`
	DisableSignup            bool   `yaml:"disable_signup"`
	StaticDir                string `yaml:"static_dir"` // Deprecated: use Storage.Filesystem.RootDir instead
	CreateSnapshotFromWebapp bool   `yaml:"create_snapshot_from_webapp"`
	WebappSnapshotterTimeout int    `yaml:"webapp_snapshotter_timeout"`
	DebugSQL                 bool   `yaml:"debug_sql"`
}

// Server holds server configuration.
type Server struct {
	Address          string `yaml:"address"`
	BaseURL          string `yaml:"base_url"`
	SecureCookie     bool   `yaml:"secure_cookie"`
	RemoteUserHeader string `yaml:"remote_user_header"`
}

// DB holds database configuration.
type DB struct {
	Connection string `yaml:"connection"`
	Type       string `yaml:"type"`
}

// Feed holds feed-related configuration.
type Feed struct {
	ItemsPerPage uint `yaml:"items_per_page"`
}

// Storage holds storage backend configuration.
type Storage struct {
	Filesystem *StorageFilesystem `yaml:"fs"`
}

// StorageFilesystem holds filesystem storage configuration.
type StorageFilesystem struct {
	RootDir string `yaml:"root_dir"`
}

// SMTP holds email server configuration.
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

// ActivityPub holds ActivityPub configuration including key paths.
type ActivityPub struct {
	PubKeyPath  string `yaml:"pubkey"`
	PrivKeyPath string `yaml:"privkey"`
	PubK        *rsa.PublicKey
	PrivK       *rsa.PrivateKey
}

// OAuth maps provider names to their OAuth configurations.
type OAuth map[string]OAuthEntry

// OAuthEntry holds configuration for a single OAuth provider.
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

// Load reads and parses the configuration from the specified file.
func Load(filename string) (*Config, error) {
	b, fn, err := readConfigFile(filename)
	if err != nil {
		log.Debug().Msg("No config file found, using default config")
		c := CreateDefaultConfig()
		c.setDefaultStorage()
		//lint:ignore nilerr // no need to check error
		return c, nil
	}
	c, err := parseConfig(b)
	if err != nil {
		return nil, err
	}
	c.fname = fn
	return c, nil
}

// CreateDefaultConfig returns a new Config with default values.
func CreateDefaultConfig() *Config {
	return &Config{
		App: App{
			ResultsPerPage:           30,
			CreateSnapshotFromWebapp: false,
			WebappSnapshotterTimeout: 15,
			LogLevel:                 "info",
		},
		Server: Server{
			Address:      "127.0.0.1:7331",
			BaseURL:      "http://127.0.0.1:7331",
			SecureCookie: false,
		},
		DB: DB{
			Type:       "sqlite",
			Connection: "db.sqlite3",
		},
		Feed: Feed{
			ItemsPerPage: 20,
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
	if c.Server.BaseURL == "" {
		c.Server.BaseURL = fmt.Sprintf("http://%s", c.Server.Address)
	}
	pu, err := url.Parse(c.Server.BaseURL)
	if err != nil || pu.Scheme == "" || pu.Host == "" {
		return nil, errors.New("invalid Server.BaseURL - use 'https://domain.tld/xy/' format")
	}
	if strings.HasSuffix(c.Server.BaseURL, "/") {
		c.Server.BaseURL = c.Server.BaseURL[:len(c.Server.BaseURL)-1]
	}
	if c.App.StaticDir != "" {
		if c.Storage.Filesystem != nil {
			return nil, errors.New("remove app.static_dir from config, storage.fs is already configured")
		}
		log.Warn().Msg("app.static_dir is deprecated, use storage.fs.root_dir instead to configure where bookmark content is stored")
		c.Storage.Filesystem = &StorageFilesystem{
			RootDir: filepath.Join(c.App.StaticDir, "data"),
		}
	}
	count := 0
	if c.Storage.Filesystem != nil {
		count += 1
	}
	if count > 1 {
		return nil, errors.New("only one storage backend can be configured")
	} else if count == 0 {
		// Default filesystem config
		c.setDefaultStorage()
	}
	return c, nil
}

// Filename returns the path of the loaded configuration file.
func (c *Config) Filename() string {
	if c.fname == "" {
		return "*Default Config*"
	}
	return c.fname
}

func (c *Config) setDefaultStorage() {
	c.Storage.Filesystem = &StorageFilesystem{
		RootDir: "./static/data",
	}
}

// BaseURL constructs a full URL by appending the given path to the server's base URL.
func (c *Config) BaseURL(u string) string {
	if strings.HasPrefix(u, "/") && strings.HasSuffix(c.Server.BaseURL, "/") {
		u = u[1:]
	}
	if !strings.HasPrefix(u, "/") && !strings.HasSuffix(c.Server.BaseURL, "/") {
		u = "/" + u
	}
	return c.Server.BaseURL + u
}

// ExportPrivKey exports the private key in PEM format.
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

// ParsePrivKey parses a private key from PEM format.
func (ap *ActivityPub) ParsePrivKey(privPEM []byte) error {
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return errors.New("failed to parse PEM block containing the key")
	}

	var err error
	ap.PrivK, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	return err
}

// ExportPubKey exports the public key in PEM format.
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

// ParsePubKey parses a public key from PEM format.
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
