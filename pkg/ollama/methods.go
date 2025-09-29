package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/thejasmeetsingh/oclai/pkg/utils"
)

// CheckOllamaConnection checks if the Ollama service is running at the specified URL
func CheckOllamaConnection(url string) error {
	// Make a GET request to the 'tags' endpoint to check the connection
	resp, err := http.Get(url + "/api/tags")
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama service at %s: %v\nPlease ensure Ollama is running", url, err)
	}
	defer resp.Body.Close()

	// Check if the response status code is HTTP 200 OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

// SystemPromptMessage returns a system message for the assistant
func SystemPromptMessage() Message {
	return Message{
		Role:    SystemRole,
		Content: "You are a helpful assistant. You are designed to be helpful, honest, and safe.",
	}
}

// ListModels retrieves a list of available models from the Ollama service
func ListModels(url string) ([]ModelInfo, error) {
	// Make a GET request to the 'tags' endpoint to get the list of models
	resp, err := http.Get(url + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the response status code is HTTP 200 OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Decode the JSON response into a ModelsResponse struct
	var modelsResp ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, err
	}

	return modelsResp.Models, nil
}

// ShowModels displays the available models in a formatted way
func ShowModels(url string, models *[]ModelInfo) (string, error) {
	// If no models are provided, fetch them from the Ollama service
	if models == nil {
		newModels, err := ListModels(url)
		if err != nil {
			return "", err
		}
		models = &newModels
	}

	// If no models are found, return an error
	if len(*models) == 0 {
		return "", fmt.Errorf("no models found. Please install a model using: ollama pull <model-name>")
	}

	// Create a slice to store model details
	var modelDetails []string

	for idx, model := range *models {
		sizeGB := float64(model.Size) / (1024 * 1024 * 1024)
		modelDetail := fmt.Sprintf("%d. **%s** (%.1f GB) - Modified At: %s", idx+1, model.Name, sizeGB, model.ModifiedAt.Format("2006-01-02 15:04:05"))
		modelDetails = append(modelDetails, modelDetail)
	}

	// Format the content with a header and join the model messages
	content := fmt.Sprintf("# 📋 Available Models\n%s", strings.Join(modelDetails, "\n"))
	return utils.ToMarkDown(content)
}

// Chat sends a chat request to the Ollama service and returns the response
func Chat(url string, request ModelRequest) (*ModelResponse, error) {
	// Create a buffer to hold the request body
	body := &bytes.Buffer{}

	// Encode the request into JSON
	encoder := json.NewEncoder(body)
	encoder.Encode(request)

	// Send a POST request to the 'chat' endpoint
	response, err := http.Post(url+"/api/chat", "application/json", body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Check if the response status code is HTTP 200 OK
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", response.StatusCode)
	}

	// Parse response to ModelResponse struct
	var modelResponse ModelResponse
	if err = json.NewDecoder(response.Body).Decode(&modelResponse); err != nil {
		return nil, fmt.Errorf("error while parsing the model response: %s", err.Error())
	}

	if modelResponse.Done {
		return &modelResponse, nil
	}
	return nil, fmt.Errorf("no response is returned from ollama service")
}

// A util function to check if the given model exists or not.
func IsModelExists(url, model string, models *[]ModelInfo) (bool, error) {
	// If no models are provided, fetch them from the Ollama service
	if models == nil {
		newModels, err := ListModels(url)
		if err != nil {
			return false, err
		}
		models = &newModels
	}

	// If no models are found, return an error
	if len(*models) == 0 {
		return false, fmt.Errorf("no models found. Please install a model using: ollama pull <model-name>")
	}

	// Check whether the given model exists
	for _, _model := range *models {
		if strings.EqualFold(model, _model.Name) {
			return true, nil
		}
	}

	return false, nil
}
