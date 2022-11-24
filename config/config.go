package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config is the service configuration
type Config struct {
	Base          Base          `yaml:"base"`
	HTTP          HttpSettings  `yaml:"http"`
	Logging       Logging       `yaml:"logging"`
	LoadBalancing LoadBalancing `yaml:"load_balancing"`
	Store         Store         `yaml:"store"`
	AutoApprove   AutoApprove   `yaml:"auto_approval"`
	Websocket     WebSocketConf `yaml:"websocket"`
}

type Base struct {
	PIN      int    `yaml:"pin"`
	QredoAPI string `yaml:"qredo_api"`
}

type AutoApprove struct {
	Enabled          bool `yaml:"enabled"`
	RetryIntervalMax int  `yaml:"retry_interval_max_sec"`
	RetryInterval    int  `yaml:"retry_interval_sec"`
}
type WebSocketConf struct {
	QredoWebsocket    string `yaml:"qredo_websocket"`
	ReconnectTimeOut  int    `yaml:"reconnect_timeout_sec"`
	ReconnectInterval int    `yaml:"reconnect_interval_sec"`
	PingPeriod        int    `yaml:"ping_period_sec"`
	PongWait          int    `yaml:"pong_wait_sec"`
	WriteWait         int    `yaml:"write_wait_sec"`
	ReadBufferSize    int    `yaml:"read_buffer_size"`
	WriteBufferSize   int    `yaml:"write_buffer_size"`
}
type Store struct {
	Type       string    `default:"file" yaml:"type"`
	FileConfig string    `yaml:"file"`
	OciConfig  OciConfig `yaml:"oci"`
	AwsConfig  AWSConfig `yaml:"aws"`
}

type OciConfig struct {
	Compartment         string `yaml:"compartment"`
	Vault               string `yaml:"vault"`
	SecretEncryptionKey string `yaml:"secret_encryption_key"`
	ConfigSecret        string `yaml:"config_secret"`
}

// AWSConfig based signing-agent config. Used when Base StoreType is aws.
type AWSConfig struct {
	Region     string `yaml:"region"`
	SecretName string `yaml:"config_secret"`
}

type HttpSettings struct {
	Addr             string   `yaml:"addr"`
	CORSAllowOrigins []string `yaml:"cors_allow_origins"`
	LogAllRequests   bool     `yaml:"log_all_requests"`
}

type Logging struct {
	Format string `yaml:"format"`
	Level  string `yaml:"level"`
}

type LoadBalancing struct {
	Enable                bool        `yaml:"enable"`
	OnLockErrorTimeOutMs  int         `yaml:"on_lock_error_timeout_ms"`
	ActionIDExpirationSec int         `yaml:"action_id_expiration_sec"`
	RedisConfig           RedisConfig `yaml:"redis"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// Default creates configuration with default values.
func (c *Config) Default() {
	c.HTTP = HttpSettings{
		Addr:             "127.0.0.1:8007",
		CORSAllowOrigins: []string{"*"},
	}

	c.Base.PIN = 0
	c.Base.QredoAPI = "https://play-api.qredo.network/api/v1/p"
	c.AutoApprove = AutoApprove{
		Enabled:          false,
		RetryIntervalMax: 300,
		RetryInterval:    5,
	}
	c.Websocket = WebSocketConf{
		ReconnectTimeOut:  300,
		ReconnectInterval: 5,
		QredoWebsocket:    "wss://play-api.qredo.network/api/v1/p/coreclient/feed",
		PingPeriod:        5,
		PongWait:          10,
		WriteWait:         10,
		ReadBufferSize:    512,
		WriteBufferSize:   1024,
	}
	c.Logging.Level = "info"
	c.Logging.Format = "json"
	c.Store.Type = "file"
	c.Store.FileConfig = "ccstore.db"
	c.LoadBalancing = LoadBalancing{
		Enable:                false,
		OnLockErrorTimeOutMs:  300,
		ActionIDExpirationSec: 6,
		RedisConfig: RedisConfig{
			Host:     "redis",
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}
}

// Load reads and parses yaml config.
func (c *Config) Load(fileName string) error {
	f, err := os.ReadFile(fileName)
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
	return os.WriteFile(fileName, b, 0600)
}
