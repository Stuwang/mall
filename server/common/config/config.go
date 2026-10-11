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

// IsTest 是否测试环境（鉴权走 UserId Header）
func (c *Config) IsTest() bool {
	return c.Env == EnvTest
}

// Load 从 yaml 文件加载配置
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
		c.Server.HTTPAddr = ":8080"
	}
	if c.Database.Filename == "" {
		c.Database.Filename = "mall.db"
	}
	if c.Auth.JWTExpireHours <= 0 {
		c.Auth.JWTExpireHours = 168
	}
}

func (c *Config) validate() error {
	if !c.IsTest() && c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required when env is not %s", EnvTest)
	}
	return nil
}
