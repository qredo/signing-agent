package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config is the service configuration
type Config struct {
	Base          Base          `yaml:"base"`
	HTTP          httpSettings  `yaml:"http"`
	Logging       Logging       `yaml:"logging"`
	Redis         Redis         `yaml:"redis"`
	LoadBalancing LoadBalancing `yaml:"load_balancing"`
}

type Base struct {
	PIN              int       `yaml:"pin"`
	QredoAPIDomain   string    `yaml:"qredo_api_domain"`
	QredoAPIBasePath string    `yaml:"qredo_api_base_path"`
	AutoApprove      bool      `yaml:"auto_approve"`
	HttpScheme       string    `yaml:"http_scheme"`
	WsScheme         string    `yaml:"ws_scheme"`
	StoreType        string    `default:"file" yaml:"store_type"`
	StoreFile        string    `yaml:"store_file"`
	StoreOci         OciConfig `yaml:"store_oci"`
}

type OciConfig struct {
	Compartment         string `yaml:"compartment"`
	Vault               string `yaml:"vault"`
	SecretEncryptionKey string `yaml:"secret_encryption_key"`
	ConfigSecret        string `yaml:"config_secret"`
}

type httpSettings struct {
	Addr                 string   `yaml:"addr"`
	CORSAllowOrigins     []string `yaml:"cors_allow_origins"`
	ProxyForwardedHeader string   `yaml:"proxy_forwarded_header"`
	LogAllRequests       bool     `yaml:"log_all_requests"`
}

type Logging struct {
	Format string `yaml:"format"`
	Level  string `yaml:"level"`
}

type LoadBalancing struct {
	Enable                bool `yaml:"enable"`
	OnLockErrorTimeOutMs  int  `yaml:"on_lock_error_timeout_ms"`
	ActionIDExpirationSec int  `yaml:"action_id_expiration_sec"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
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
	c.Base.AutoApprove = false
	c.Base.HttpScheme = "https"
	c.Base.StoreType = "file"
	c.Base.StoreFile = "ccstore.db"
	c.Redis = Redis{
		Host:     "redis",
		Port:     6379,
		Password: "",
		DB:       0,
	}
	c.LoadBalancing = LoadBalancing{
		Enable:                false,
		OnLockErrorTimeOutMs:  300,
		ActionIDExpirationSec: 6,
	}
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
