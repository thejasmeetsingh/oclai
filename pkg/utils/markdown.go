package utils

import (
	"github.com/charmbracelet/glamour"
)

// ToMarkDown converts the given content into Markdown format using the glamour library.
func ToMarkDown(content string) (string, error) {
	// Create a new TermRenderer with specific styling and formatting options.
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"), // Applies a dark-themed style to the rendered content.
		glamour.WithWordWrap(100),         // Wraps the text at 100 characters per line for better readability.
	)
	if err != nil {
		return "", err
	}

	// Render the content using the configured renderer and return the result.
	return renderer.Render(content)
}
