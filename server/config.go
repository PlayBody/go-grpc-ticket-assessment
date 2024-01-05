package server

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

var (
	SConfig = Config{}
)

type Config struct {
	Train     TrainConfig `yaml:"train"`
	Auth      AuthConfig  `yaml:"auth"`
	RoleUsers []RoleUser  `yaml:"roles"`
}

type AuthConfig struct {
	SecretKey string `yaml:"secret_key"`
	Expire    int64  `yaml:"expire"`
}

type RoleUser struct {
	Email        string   `yaml:"email"`
	Capabilities []string `yaml:"caps"`
}

type TrainConfig struct {
	Sections  []string `yaml:"sections,omitempty"`
	SeatCount int      `yaml:"seat_count,omitempty"`
	Routes    []struct {
		From  string `yaml:"from,omitempty"`
		To    string `yaml:"to,omitempty"`
		Price int32  `yaml:"price,omitempty"`
	} `yaml:"routes,omitempty"`
}

func (s *Config) InitConfig(path string) error {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Error reading YAML file: %v\n", err)
	}

	err = yaml.Unmarshal(yamlFile, s)
	if err != nil {
		return fmt.Errorf("Error unmarshalling YAML content: %v\n", err)
	}
	return nil
}
