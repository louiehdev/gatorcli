package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileName string = ".gatorconfig.json"

type Config struct {
	Url      string `json:"db_url"`
	Username string `json:"current_user_name"`
}

func (c *Config) SetUser(user string) error {
	c.Username = user

	configData, err := json.Marshal(c)
	if err != nil {
		return err
	}

	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	if err := os.WriteFile(configFilePath, configData, 0644); err != nil {
		return err
	}

	return nil
}

func Read() (Config, error) {
	var config Config
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return config, err
	}
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configFilePath := fmt.Sprintf("%v/Documents/workspace/github/gatorcli/%v", homeDir, configFileName)
	return configFilePath, nil
}
