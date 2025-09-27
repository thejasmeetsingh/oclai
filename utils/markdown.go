package utils

import (
	"github.com/charmbracelet/glamour"
)

func ToMarkDown(content string) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return "", err
	}

	return renderer.Render(content)
}
