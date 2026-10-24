package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const EnvTest = "test"

type Config struct {
	Env      string         `yaml:"env"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logger   LoggerConfig   `yaml:"logger"`
	Auth     AuthConfig     `yaml:"auth"`
	IM       IMConfig       `yaml:"im"`
}

type ServerConfig struct {
	HTTPAddr string `yaml:"http_addr"`
}

type DatabaseConfig struct {
	Filename string `yaml:"filename"`
}

type LoggerConfig struct {
	Filename string `yaml:"filename"`
}

type AuthConfig struct {
	JWTSecret      string `yaml:"jwt_secret"`
	JWTExpireHours int    `yaml:"jwt_expire_hours"`
}

type IMConfig struct {
	SystemToken     string `yaml:"system_token"`
	HeartbeatSec    int    `yaml:"heartbeat_sec"`
	IdleSec         int    `yaml:"idle_sec"`
	MaxConnsPerUser int    `yaml:"max_conns_per_user"`
	TextMaxRunes    int    `yaml:"text_max_runes"`
}

func (c *Config) IsTest() bool {
	return c.Env == EnvTest
}

func Load(path string) (*Config, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(bs, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Env == "" {
		c.Env = EnvTest
	}
	if c.Server.HTTPAddr == "" {
		c.Server.HTTPAddr = ":8090"
	}
	if c.Database.Filename == "" {
		c.Database.Filename = "im.db"
	}
	if c.Auth.JWTExpireHours <= 0 {
		c.Auth.JWTExpireHours = 168
	}
	if c.IM.HeartbeatSec <= 0 {
		c.IM.HeartbeatSec = 25
	}
	if c.IM.IdleSec <= 0 {
		c.IM.IdleSec = 70
	}
	if c.IM.MaxConnsPerUser <= 0 {
		c.IM.MaxConnsPerUser = 5
	}
	if c.IM.TextMaxRunes <= 0 {
		c.IM.TextMaxRunes = 2000
	}
}

func (c *Config) validate() error {
	if !c.IsTest() && c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required when env is not %s", EnvTest)
	}
	if c.IM.SystemToken == "" {
		return fmt.Errorf("im.system_token is required")
	}
	return nil
}
