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
	LogAllRequests       bool     `yaml:"log_all_requests"`
}

// Config is the service configuration
type Config struct {
	Base    Base         `yaml:"base"`
	HTTP    httpSettings `yaml:"http"`
	Logging Logging      `yaml:"logging"`
}

type Base struct {
	PIN              int    `yaml:"int"`
	QredoAPIDomain   string `yaml:"qredo_api_domain"`
	QredoAPIBasePath string `yaml:"qredo_api_base_path"`
	StoreFile        string `yaml:"store_file"`
	AutoApprove      bool   `yaml:"auto_approve"`
	HttpScheme       string `yaml:"http_scheme"`
	WsScheme         string `yaml:"ws_scheme"`
}

type Logging struct {
	Format string `yaml:"format"`
	Level  string `yaml:"level"`
}

// Default creates configuration with default values.
func (c *Config) Default() {
	c.HTTP = httpSettings{
		Addr:             "127.0.0.1:8007",
		CORSAllowOrigins: []string{"*"},
	}
	c.Base.WsScheme = "wss://"
	c.Base.PIN = 0
	c.Base.QredoAPIDomain = "play-api.qredo.network"
	c.Base.QredoAPIBasePath = "/api/v1/p"
	c.Logging.Level = "info"
	c.Logging.Format = "json"
	c.Base.StoreFile = "ccstore.db"
	c.Base.AutoApprove = false
	c.Base.HttpScheme = "https"
}

// Load reads and parses yaml config.
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

// Save saves yaml config.
func (c *Config) Save(fileName string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, b, 0600)
}
