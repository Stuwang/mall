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
	Order    OrderConfig    `yaml:"order"`
	Point    PointConfig    `yaml:"point"`
	// Admins 启动时幂等写入的管理员账号（按 username 去重）
	Admins []AdminAccount `yaml:"admins"`
}

// AdminAccount 初始管理员账号
type AdminAccount struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Nickname string `yaml:"nickname"`
}

type OrderConfig struct {
	// 待支付超时秒数
	PayTimeoutSec int `yaml:"pay_timeout_sec"`
}

type PointConfig struct {
	// 1 元（100 分）可兑换的积分数；会在下单/购积分时快照到订单
	PointsPerYuan int64 `yaml:"points_per_yuan"`
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
	if c.Order.PayTimeoutSec <= 0 {
		c.Order.PayTimeoutSec = 15 * 60
	}
	if c.Point.PointsPerYuan <= 0 {
		c.Point.PointsPerYuan = 100
	}
}

func (c *Config) validate() error {
	if !c.IsTest() && c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required when env is not %s", EnvTest)
	}
	seen := make(map[string]struct{}, len(c.Admins))
	for i, a := range c.Admins {
		if a.Username == "" {
			return fmt.Errorf("admins[%d].username is required", i)
		}
		if a.Password == "" {
			return fmt.Errorf("admins[%d].password is required", i)
		}
		if _, ok := seen[a.Username]; ok {
			return fmt.Errorf("admins[%d].username %q duplicated in config", i, a.Username)
		}
		seen[a.Username] = struct{}{}
	}
	return nil
}
