package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/thejasmeetsingh/oclai/utils"
)

// CheckOllamaConnection checks if the Ollama service is running
func CheckOllamaConnection(url string) error {
	resp, err := http.Get(url + "/api/tags")
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama service at %s: %v\nPlease ensure Ollama is running", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama service returned status %d", resp.StatusCode)
	}

	return nil
}

func SystemPromptMessage() Message {
	return Message{
		Role:    SystemRole,
		Content: "You are a helpful Assistant!",
	}
}

// ListModels retrieves a list of available models from the Ollama service
func ListModels(url string) ([]ModelInfo, error) {
	resp, err := http.Get(url + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama service returned status %d", resp.StatusCode)
	}

	var modelsResp ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, err
	}

	return modelsResp.Models, nil
}

// Fetch models from ollama and return them in an appropriate format
func ShowModels(url string, models *[]ModelInfo) (string, error) {
	if models == nil {
		newModels, err := ListModels(url)
		if err != nil {
			return "", err
		}
		models = &newModels
	}

	if len(*models) == 0 {
		return "", fmt.Errorf("no models found. Please install a model using: ollama pull <model-name>")
	}

	var modelMsgs []string

	for idx, model := range *models {
		sizeGB := float64(model.Size) / (1024 * 1024 * 1024)
		modelMsg := fmt.Sprintf("%d. **%s** (%.1f GB) - Modified At: %s", idx+1, model.Name, sizeGB, model.ModifiedAt.Format("2006-01-02 15:04:05"))
		modelMsgs = append(modelMsgs, modelMsg)
	}

	content := fmt.Sprintf("# 📋 Available Models\n%s", strings.Join(modelMsgs, "\n"))
	return utils.ToMarkDown(content)
}

// Chat sends a chat request to the model API and processes the response
func Chat(url string, request ModelRequest, showStats bool) (string, error) {
	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)
	encoder.Encode(request)

	response, err := http.Post(url+"/api/chat", "application/json", body)
	if err != nil {
		return "", nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d", response.StatusCode)
	}

	var modelResponse ModelResponse
	if err = json.NewDecoder(response.Body).Decode(&modelResponse); err != nil {
		return "", fmt.Errorf("error while parsing the model response: %s", err.Error())
	}

	if modelResponse.Done {
		content := modelResponse.Message.Content
		result, err := utils.ToMarkDown(content)
		if err != nil {
			return "", err
		}

		if showStats && modelResponse.TotalDuration > 0 {
			duration := time.Duration(modelResponse.TotalDuration)
			tokensPerSec := float64(modelResponse.EvalCount) / duration.Seconds()
			stat := color.New(color.FgGreen).Sprintf("✓ Generated %d tokens in %v (%.1f tokens/sec)\n\n",
				modelResponse.EvalCount, duration, tokensPerSec)

			result = fmt.Sprintf("%s\n%s", result, stat)
		}

		return result, nil
	}

	return "", nil
}
