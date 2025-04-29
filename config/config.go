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

	"github.com/asciimoo/omnom/oauth"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App         App          `yaml:"app"`
	Server      Server       `yaml:"server"`
	DB          DB           `yaml:"db"`
	Storage     Storage      `yaml:"storage"`
	SMTP        SMTP         `yaml:"smtp"`
	ActivityPub *ActivityPub `yaml:"activitypub"`
	OAuth       OAuth        `yaml:"oauth"`
}

type App struct {
	Debug                    bool   `yaml:"debug"`
	ResultsPerPage           int64  `yaml:"results_per_page"`
	DisableSignup            bool   `yaml:"disable_signup"`
	TemplateDir              string `yaml:"template_dir"`
	StaticDir                string `yaml:"static_dir"`
	CreateBookmarkFromWebapp bool   `yaml:"create_bookmark_from_webapp"`
	WebappSnapshotterTimeout int    `yaml:"webapp_snapshotter_timeout"`
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
	ClientID         oauth.ClientID         `yaml:"client_id"`
	ClientSecret     oauth.ClientSecret     `yaml:"client_secret"`
	ConfigurationURL oauth.ConfigurationURL `yaml:"configuration_url"`
	AuthURL          oauth.AuthURL          `yaml:"auth_url"`
	TokenURL         oauth.TokenURL         `yaml:"token_url"`
	Icon             oauth.Icon             `yaml:"icon"`
	Scopes           []oauth.ScopeValue     `yaml:"scopes"`
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
		// try $HOME/.config/omnom/config.yml
		b, err = os.ReadFile(filepath.Join(homeDir, ".config/omnom/config.yml"))
		if err == nil {
			return b, nil
		}
	}
	return b, errors.New("configuration file not found. Use --config to specify a custom config file")
}

func Load(filename string) (*Config, error) {
	b, err := readConfigFile(filename)
	if err != nil {
		return nil, err
	}
	c, err := parseConfig(b)
	return c, err
}

func parseConfig(rawConfig []byte) (*Config, error) {
	var c *Config
	err := yaml.Unmarshal(rawConfig, &c)
	if err != nil {
		return nil, err
	}
	for pn, _ := range c.OAuth {
		if _, ok := oauth.Providers[pn]; !ok {
			return nil, errors.New("invalid oauth provider: " + pn)
		}
	}
	return c, nil
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
			Type:  "RSA PRIVATE KEY",
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
			Type:  "RSA PUBLIC KEY",
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
