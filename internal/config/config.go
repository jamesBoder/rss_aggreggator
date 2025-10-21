package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, configFileName), nil
}

// Create a write function that writes the Config struct to the JSON config file
func write(config *Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return err
	}
	return nil
}

// Create a Read function that reads the JSON config file and unmarshals it into a Config struct
func Read() (*Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Export a SetUser method on the Config struct that writes the config struct to the JSON file after setting the current_user_name field.
func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName
	return write(c)
}
