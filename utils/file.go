package utils

import (
	"os"
	"path/filepath"
)

const (
	rootDirName   = ".oclai"
	dirWritePerm  = 0755
	fileWritePerm = 0644
)

func createAppDir(appRootPath string) error {
	if _, err := os.Stat(appRootPath); os.IsNotExist(err) {
		return os.MkdirAll(appRootPath, os.FileMode(dirWritePerm))
	}
	return nil
}

func GetAppRootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	appRootPath := filepath.Join(home, rootDirName)
	if err = createAppDir(appRootPath); err != nil {
		return "", err
	}

	return appRootPath, nil
}

func ReadFileContents(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func WriteFileContents(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, os.FileMode(fileWritePerm))
}
