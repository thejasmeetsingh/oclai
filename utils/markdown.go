// Provides functionality to render markdown content in the terminal with styling.

package utils

import (
	"github.com/charmbracelet/glamour"
)

// Takes a markdown content string and returns it as styled terminal output.
// It uses the glamour library to apply styling and word wrapping.
func ToMarkDown(content string) (string, error) {
	// Create a terminal renderer with auto-detected style
	renderer, err := glamour.NewTermRenderer(
		// Automatically pick dark or light style based on terminal
		glamour.WithStandardStyle("dark"),
		// Set word wrap to 80 characters
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return "", err
	}

	// Return the markdown as beautiful terminal output
	return renderer.Render(content)
}
