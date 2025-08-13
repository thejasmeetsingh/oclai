package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var configFile string

type Config struct {
	BaseURL      string `json:"baseURL"`
	DefaultModel string `json:"defaultModel"`
	File         string `json:"file"`
}

func SetupConfig() error {
	viper.SetConfigName(".oclai-config")
	viper.SetConfigType("json")
	viper.AddConfigPath("$HOME")

	viper.SetDefault("baseURL", "http://localhost:11434")
	viper.SetDefault("defaultModel", "")
	viper.SetDefault("file", filepath.Join(os.Getenv("HOME"), ".oclai-config.json"))

	return viper.ReadInConfig()
}

func GetConfigFile() string {
	if configFile != "" {
		return configFile
	}

	return viper.GetString("file")
}
