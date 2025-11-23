package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MigrateConfig 迁移配置
type MigrateConfig struct {
	ZenTao struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"zentao"`
	GoProject struct {
		Type string `yaml:"type"`
		DSN  string `yaml:"dsn"`
	} `yaml:"goproject"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*MigrateConfig, error) {
	if configPath == "" {
		configPath = "migrate-config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config MigrateConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

