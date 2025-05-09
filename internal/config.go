package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AnswerRate float64             `yaml:"answer_rate"`
	Suffixes   map[string][]string `yaml:"suffixes"`
}

func NewConfig(path string) (*Config, error) {
	var config Config

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return &config, nil
}
