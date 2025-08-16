// Package markdown provides functionality to render markdown content in the terminal with styling.
package markdown

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

// Render takes a markdown content string and renders it as styled terminal output.
// It uses the glamour library to apply styling and word wrapping.
func Render(content string) error {
	// Create a terminal renderer with auto-detected style
	renderer, err := glamour.NewTermRenderer(
		// Automatically pick dark or light style based on terminal
		glamour.WithAutoStyle(),
		// Set word wrap to 80 characters
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return err
	}

	// Render the markdown as beautiful terminal output
	output, err := renderer.Render(content)
	if err != nil {
		return err
	}

	// Print the formatted output
	fmt.Println(output)
	return nil
}
