package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// Constants for directory and file permissions
const (
	// rootDirName is the name of the root directory used for application data
	rootDirName = ".oclai"

	// dirWritePerm is the permission mode for creating directories
	dirWritePerm = 0755

	// fileWritePerm is the permission mode for creating files
	fileWritePerm = 0644
)

// createAppDir creates the application root directory if it doesn't exist
func createAppDir(appRootPath string) error {
	// Check if the directory exists
	if _, err := os.Stat(appRootPath); os.IsNotExist(err) {
		// If it doesn't exist, create it with the specified permissions
		return os.MkdirAll(appRootPath, os.FileMode(dirWritePerm))
	}
	return nil
}

// GetAppRootDir returns the application root directory path
func GetAppRootDir() (string, error) {
	// Get the user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Construct the application root directory path
	appRootPath := filepath.Join(home, rootDirName)

	// Create the application root directory if it doesn't exist
	if err = createAppDir(appRootPath); err != nil {
		return "", err
	}

	return appRootPath, nil
}

// ReadConfig reads the contents of a configuration file
func ReadConfig(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// WriteFileContents writes the given data to a file
func WriteFileContents(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, os.FileMode(fileWritePerm))
}

// isValidFilePath checks if a file path is valid
func isValidFilePath(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// readFromReader reads content from a file reader
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

// ReadFileContent reads the contents of a file
func ReadFileContent(filePath string) ([]string, error) {
	var result []string

	if !isValidFilePath(filePath) {
		return result, fmt.Errorf("'%s' is not a valid file path", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return result, fmt.Errorf("failed to open file: %s", err)
	}
	defer file.Close()

	return readFromReader(file)
}

// ReadPipedInput reads input from standard input (stdin)
func ReadPipedInput() ([]string, error) {
	var result []string

	stat, err := os.Stdin.Stat()
	if err != nil {
		return result, fmt.Errorf("failed to check stdin status: %s", err)
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return readFromReader(os.Stdin)
	}

	return result, nil
}
