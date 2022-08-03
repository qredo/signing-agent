package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
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
	Base      Base         `yaml:"base"`
	HTTP      httpSettings `yaml:"http"`
	Logging   Logging      `yaml:"logging"`
	StoreFile string       `yaml:"store_file"`
}

type Base struct {
	URL                  string `yaml:"url"`
	PIN                  int    `yaml:"int"`
	QredoURL             string `yaml:"qredo_url"`
	QredoAPIDomain       string `yaml:"qredo_api_domain"`
	QredoAPIBasePath     string `yaml:"qredo_api_base_path"`
	RetrySleepGetAgentID int    `yaml:"retry_sleep_get_agent_id"`
	PrivatePEMFilePath   string `yaml:"private_key_path"`
	APIKeyFilePath       string `yaml:"api_key_path"`
}

type Logging struct {
	Format string `yaml:"format"`
	Level  string `yaml:"level"`
}

// Default creates configuration with default values
func (c *Config) Default() {
	c.HTTP = httpSettings{
		Addr:               "127.0.0.1:8007",
		CORSAllowOrigins:   []string{"*"},
		CookieSameSiteMode: "lax",
		CookieSecure:       true,
	}
	c.Base.URL = "http://127.0.0.1:8007"
	c.Base.PIN = 0
	c.Base.QredoURL = "https://play-api.qredo.network"
	c.Logging.Level = "info"
	c.Logging.Format = "json"
	c.StoreFile = "ccstore.db"
	c.Base.PrivatePEMFilePath = "private.pem"
	c.Base.APIKeyFilePath = "apikey"
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
