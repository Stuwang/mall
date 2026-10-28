package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

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
	Storage  StorageConfig  `yaml:"storage"`
	Coin     CoinConfig     `yaml:"coin"`

	gifts []Gift
}

type StorageConfig struct {
	PublicBase     string      `yaml:"public_base"`
	MaxUploadBytes int64       `yaml:"max_upload_bytes"`
	Local          LocalStorage `yaml:"local"`
}

type LocalStorage struct {
	RootDir string `yaml:"root_dir"`
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

// CoinConfig 金币汇率、支付超时和礼物文件。
type CoinConfig struct {
	CoinsPerYuan  int64  `yaml:"coins_per_yuan"`
	PayTimeoutSec int64  `yaml:"pay_timeout_sec"`
	GiftsFile     string `yaml:"gifts_file"`
}

// Gift 礼物配置。顺序与文件中的顺序一致。
type Gift struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Effect string `json:"effect"`
	Price  int64  `json:"price"`
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
	if err := cfg.loadGifts(path); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Gifts 返回已加载的礼物，顺序与配置文件一致。
func (c *Config) Gifts() []Gift {
	out := make([]Gift, len(c.gifts))
	copy(out, c.gifts)
	return out
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
	if c.Storage.PublicBase == "" {
		c.Storage.PublicBase = "http://localhost:8090"
	}
	if c.Storage.MaxUploadBytes <= 0 {
		c.Storage.MaxUploadBytes = 5 << 20
	}
	if c.Storage.Local.RootDir == "" {
		c.Storage.Local.RootDir = "data/images"
	}
	if c.Coin.CoinsPerYuan <= 0 {
		c.Coin.CoinsPerYuan = 10
	}
	if c.Coin.PayTimeoutSec <= 0 {
		c.Coin.PayTimeoutSec = 900
	}
	if c.Coin.GiftsFile == "" {
		c.Coin.GiftsFile = "gifts.json"
	}
}

func (c *Config) validate() error {
	if !c.IsTest() && c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required when env is not %s", EnvTest)
	}
	if c.IM.SystemToken == "" {
		return fmt.Errorf("im.system_token is required")
	}
	if c.Coin.CoinsPerYuan <= 0 {
		return fmt.Errorf("coin.coins_per_yuan must be > 0")
	}
	if c.Coin.PayTimeoutSec <= 0 {
		return fmt.Errorf("coin.pay_timeout_sec must be > 0")
	}
	return validateGifts(c.gifts)
}

func (c *Config) loadGifts(configPath string) error {
	path := c.Coin.GiftsFile
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(configPath), path)
	}
	bs, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read gifts %s: %w", path, err)
	}
	var gifts []Gift
	if err := json.Unmarshal(bs, &gifts); err != nil {
		return fmt.Errorf("parse gifts %s: %w", path, err)
	}
	c.gifts = gifts
	return nil
}

func validateGifts(gifts []Gift) error {
	if len(gifts) == 0 {
		return fmt.Errorf("coin.gifts requires at least one gift")
	}
	seen := make(map[int64]struct{}, len(gifts))
	for i, g := range gifts {
		if g.ID <= 0 {
			return fmt.Errorf("coin.gifts[%d] id must be > 0", i)
		}
		if _, ok := seen[g.ID]; ok {
			return fmt.Errorf("coin.gifts[%d] duplicate id %d", i, g.ID)
		}
		seen[g.ID] = struct{}{}
		if g.Name == "" {
			return fmt.Errorf("coin.gifts[%d] name is empty", i)
		}
		if utf8.RuneCountInString(g.Name) > 32 {
			return fmt.Errorf("coin.gifts[%d] name too long", i)
		}
		if len(g.Icon) > 512 {
			return fmt.Errorf("coin.gifts[%d] icon too long", i)
		}
		if len(g.Effect) > 64 {
			return fmt.Errorf("coin.gifts[%d] effect too long", i)
		}
		if g.Price <= 0 {
			return fmt.Errorf("coin.gifts[%d] price must be > 0", i)
		}
	}
	return nil
}
