package config

import (
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
)

var ConfigCache = &Config{}

func init() {
	// 加载配置文件
	err := LoadYamlConfig("config.yaml", ConfigCache)
	if err != nil {
		slog.Error("加载配置文件异常", "error", err)
		return
	}
}

type Token struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
}

type Account struct {
	Name   string `yaml:"name"`
	Addr   string `yaml:"addr"`
	IsAlet bool   `yaml:"isAlet"`
}

type Config struct {
	Mysql struct {
		User            string `yaml:"user"`
		Password        string `yaml:"pwd"`
		Host            string `yaml:"host"`
		DBName          string `yaml:"db"`
		MaxOpenConns    int    `yaml:"maxOpenConns"`
		MaxIdleConns    int    `yaml:"maxIdleConns"`
		ConnMaxLifetime int    `yaml:"connMaxLifetime"`
	} `yaml:"mysql"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"pwd"`
		Db       int    `yaml:"db"`
	} `yaml:"redis"`
}

func LoadYamlConfig(file string, cfg interface{}) error {
	// 打开文件
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// 解析 YAML
	decoder := yaml.NewDecoder(f)
	if err = decoder.Decode(cfg); err != nil {
		return err
	}

	return nil
}
