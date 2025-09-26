package app

import (
	"encoding/json"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/thejasmeetsingh/oclai/utils"
)

const AppConfigFileName = "config"

type Config struct {
	BaseURL      string `json:"baseURL"`
	DefaultModel string `json:"defaultModel"`
	NumCtx       int    `json:"numCtx"`
	InitMCP      bool   `json:"initMCP"`
}

var OclaiConfig Config

func LoadConfig(rootPath string) error {
	filePath := filepath.Join(rootPath, AppConfigFileName)

	v := viper.New()
	v.SetConfigName(AppConfigFileName)
	v.SetConfigType("json")

	v.AddConfigPath(rootPath)
	v.SetDefault("baseURL", "http://localhost:11434")
	v.SetDefault("defaultModel", "")
	v.SetDefault("numCtx", 8000)
	v.SetDefault("initMCP", true)

	v.SafeWriteConfigAs(filePath)
	if err := v.ReadInConfig(); err != nil {
		return err
	}

	data, err := utils.ReadMcpConfig(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &OclaiConfig)
}

func UpdateConfig(rootPath string) error {
	filePath := filepath.Join(rootPath, AppConfigFileName)

	data, err := json.MarshalIndent(&OclaiConfig, "", "  ")
	if err != nil {
		return err
	}

	return utils.WriteFileContents(filePath, data)
}
