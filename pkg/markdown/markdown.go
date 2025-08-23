// Package markdown provides functionality to render markdown content in the terminal with styling.
package markdown

import (
	"github.com/charmbracelet/glamour"
)

// Render takes a markdown content string and returns it as styled terminal output.
// It uses the glamour library to apply styling and word wrapping.
func Render(content string) (string, error) {
	// Create a terminal renderer with auto-detected style
	renderer, err := glamour.NewTermRenderer(
		// Automatically pick dark or light style based on terminal
		glamour.WithAutoStyle(),
		// Set word wrap to 80 characters
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return "", err
	}

	// Return the markdown as beautiful terminal output
	return renderer.Render(content)
}
