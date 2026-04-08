package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Domain   string         `yaml:"domain"`
	Proxy    ProxyConfig    `yaml:"proxy"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire string `yaml:"expire"`
}

type ProxyConfig struct {
	SizeLimit int64 `yaml:"size_limit"`
	JsDelivr  bool  `yaml:"jsdelivr"`
}

var C *Config

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return err
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "sqlite"
	}
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = "data.db"
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "change-me-to-a-random-string"
	}
	if cfg.JWT.Expire == "" {
		cfg.JWT.Expire = "24h"
	}
	if cfg.Domain == "" {
		cfg.Domain = "http://localhost:8080"
	}
	if cfg.Proxy.SizeLimit == 0 {
		cfg.Proxy.SizeLimit = 1073741824
	}
	C = cfg
	return nil
}

func (c *Config) GetJWTExpire() time.Duration {
	d, err := time.ParseDuration(c.JWT.Expire)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
