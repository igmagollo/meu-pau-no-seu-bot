package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Answers struct {
	Answers []string `yaml:"answers"`
}

func NewAnswers(path string) (*Answers, error) {
	var config Answers

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
