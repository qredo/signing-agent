package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gitlab.qredo.com/qredo-server/qredo-core/qconfig"
	"gopkg.in/yaml.v2"
)

type httpSettings struct {
	Addr                 string   `yaml:"addr"`
	CORSAllowOrigins     []string `yaml:"cors_allow_origins"`
	ProxyForwardedHeader string   `yaml:"proxy_forwarded_header"`
	CookieSameSiteMode   string   `yaml:"cookie_same_site"`
	CookieSecure         bool     `yaml:"cookie_secure"`
	LogAllRequests       bool     `yaml:"log_all_requests"`
}

// Config is the service configuration
type Config struct {
	BaseURL        string          `yaml:"base_url"`
	QredoServerURL string          `yaml:"qredo_server_url"`
	HTTP           httpSettings    `yaml:"http"`
	Logging        qconfig.Logging `yaml:"logging"`
	StoreFile      string          `yaml:"store_file"`
	PIN            int             `yaml:"pin"`
}

// Default creates configuration with default values
func (c *Config) Default() {
	c.HTTP = httpSettings{
		Addr:               "127.0.0.1:8007",
		CORSAllowOrigins:   []string{"*"},
		CookieSameSiteMode: "lax",
		CookieSecure:       true,
	}
	c.BaseURL = "http://127.0.0.1:8007"
	c.Logging.Default()
	c.StoreFile = "ccstore.db"
	c.PIN = 0
	c.QredoServerURL = "http://127.0.0.1:8001"
}

// ParseConfigFile parses yaml config
func (c *Config) Load(fileName string) error {
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		return errors.Wrap(err, "read config file")
	}

	c.Default()
	if err := yaml.Unmarshal(f, c); err != nil {
		return errors.Wrap(err, "parse config file")
	}

	return nil
}

// SaveConfigFile saves yaml config
func (c *Config) Save(fileName string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, b, 0600)
}
