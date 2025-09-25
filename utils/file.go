package utils

import (
	"bufio"
	"fmt"
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

func ReadMcpConfig(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func WriteFileContents(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, os.FileMode(fileWritePerm))
}

func isValidFilePath(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func readFromReader(reader *os.File) ([]string, error) {
	fileContents := make([]string, 0)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fileContents = append(fileContents, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fileContents, err
	}

	return fileContents, nil
}

func ReadFileContent(filePath string) ([]string, error) {
	var result []string

	if !isValidFilePath(filePath) {
		return result, fmt.Errorf("'%s' is not a valid file path", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return result, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return readFromReader(file)
}

func ReadPipedInput() ([]string, error) {
	var result []string

	stat, err := os.Stdin.Stat()
	if err != nil {
		return result, fmt.Errorf("failed to check stdin status: %w", err)
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return readFromReader(os.Stdin)
	}

	return result, nil
}
