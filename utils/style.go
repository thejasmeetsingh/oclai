package utils

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// colorTheme defines our custom color palette
type colorTheme struct {
	info    lipgloss.AdaptiveColor
	success lipgloss.AdaptiveColor
	err     lipgloss.AdaptiveColor

	primary   lipgloss.AdaptiveColor
	secondary lipgloss.AdaptiveColor

	border lipgloss.AdaptiveColor
	accent lipgloss.AdaptiveColor
}

var Theme = colorTheme{
	info: lipgloss.AdaptiveColor{
		Light: "#4338CA", // Deep indigo for light mode
		Dark:  "#818CF8", // Soft lavender for dark mode
	},

	success: lipgloss.AdaptiveColor{
		Light: "#059669", // Rich emerald for light mode
		Dark:  "#34D399", // Bright mint for dark mode
	},

	err: lipgloss.AdaptiveColor{
		Light: "#DC2626", // Deep coral-red for light mode
		Dark:  "#F87171", // Soft salmon-pink for dark mode
	},

	primary: lipgloss.AdaptiveColor{
		Light: "#D97706", // Rich amber for light mode
		Dark:  "#FBBF24", // Bright gold for dark mode
	},

	secondary: lipgloss.AdaptiveColor{
		Light: "#7C3AED", // Deep violet for light mode
		Dark:  "#A78BFA", // Light purple for dark mode
	},

	border: lipgloss.AdaptiveColor{
		Light: "#E5E7EB", // Light gray for light mode
		Dark:  "#374151", // Dark gray for dark mode
	},

	accent: lipgloss.AdaptiveColor{
		Light: "#0891B2", // Deep cyan for light mode
		Dark:  "#22D3EE", // Bright cyan for dark mode
	},
}

// Message Styles
var (
	infoStyle = lipgloss.NewStyle().
			Foreground(Theme.info).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Theme.info).
			BorderLeft(true)

	successStyle = lipgloss.NewStyle().
			Foreground(Theme.success).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Theme.success).
			BorderLeft(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(Theme.err).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Theme.err).
			BorderLeft(true)

	otherStyle = lipgloss.NewStyle().
			Foreground(Theme.accent).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Theme.accent).
			BorderLeft(true)
)

// Loader/Spinner Style
var LoaderStyle = lipgloss.NewStyle().
	Foreground(Theme.primary).
	Bold(true)

// Enhanced message styles with icons and background
var (
	infoBoxStyle = lipgloss.NewStyle().
			Foreground(Theme.info).
			Background(lipgloss.AdaptiveColor{Light: "#EEF2FF", Dark: "#1E1B4B"}).
			Bold(true).
			Padding(0, 1).
			Margin(1, 0).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Theme.info)

	successBoxStyle = lipgloss.NewStyle().
			Foreground(Theme.success).
			Background(lipgloss.AdaptiveColor{Light: "#ECFDF5", Dark: "#064E3B"}).
			Bold(true).
			Padding(0, 1).
			Margin(1, 0).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Theme.success)

	errorBoxStyle = lipgloss.NewStyle().
			Foreground(Theme.err).
			Background(lipgloss.AdaptiveColor{Light: "#FEF2F2", Dark: "#7F1D1D"}).
			Bold(true).
			Padding(0, 1).
			Margin(1, 0).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Theme.err)

	userMsgBoxStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Margin(1, 0).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Theme.primary).
			BorderLeft(true)

	aiMsgBoxStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Margin(1, 0).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Theme.secondary).
			BorderLeft(true)
)

// Message formatting functions
func InfoMessage(message string) string {
	return infoStyle.Render("ℹ " + message)
}

func SuccessMessage(message string) string {
	return successStyle.Render("✓ " + message)
}

func ErrorMessage(message string) string {
	return errorStyle.Render("✗ " + message)
}

func OtherMessage(message string) string {
	return otherStyle.Render(message)
}

// Box-style messages for more important notifications
func InfoBox(message string) string {
	return infoBoxStyle.Render(message)
}

func SuccessBox(message string) string {
	return successBoxStyle.Render(message)
}

func ErrorBox(message string) string {
	return errorBoxStyle.Render(message)
}

func UserMsgBox(timestamp, message string) string {
	return userMsgBoxStyle.Render(fmt.Sprintf("[%s] 👤:\n%s", timestamp, message))
}

func AiMsgBox(timestamp, message string) string {
	return aiMsgBoxStyle.Render(fmt.Sprintf("[%s] 🤖:\n%s", timestamp, message))
}
