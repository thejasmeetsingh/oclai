package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	BaseURL      string `json:"baseURL"`
	DefaultModel string `json:"defaultModel"`
	File         string `json:"file"`
}

var OclaiConfig Config

func setupConfig() error {
	filePath := filepath.Join(os.Getenv("HOME"), ".oclai-config.json")

	viper.SetConfigName(".oclai-config")
	viper.SetConfigType("json")
	viper.AddConfigPath("$HOME")

	viper.SetDefault("baseURL", "http://localhost:11434")
	viper.SetDefault("defaultModel", "")
	viper.SetDefault("file", filePath)

	viper.SafeWriteConfigAs(filePath)
	return viper.ReadInConfig()
}

func LoadConfig() error {
	err := setupConfig()
	if err != nil {
		return err
	}

	file := viper.GetString("file")
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &OclaiConfig)
}

func UpdateConfig() error {
	data, err := json.MarshalIndent(&OclaiConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(OclaiConfig.File, data, 0644)
}

func DefaultModelCheck() error {
	if OclaiConfig.DefaultModel == "" {
		return fmt.Errorf("please select a default model 🤖")
	}
	return nil
}
