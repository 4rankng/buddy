package ui

import (
	"fmt"
	"os"
	"strings"
)

// CreateHyperlink creates a clickable hyperlink using ANSI escape sequences
// Falls back to plain text if hyperlinks are disabled
func CreateHyperlink(url, text string) string {
	// Check if hyperlinks are disabled via environment variable
	if os.Getenv("JIRA_NO_HYPERLINKS") == "1" {
		return fmt.Sprintf("%s (%s)", text, url)
	}

	// ANSI hyperlink escape sequence: \x1b]8;;URL\x1b\URL Text\x1b]8;;\x1b\
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

// IsInteractive returns true if running in an interactive terminal
func IsInteractive() bool {
	// Check if stdout is a terminal
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// PromptForInput prompts the user for input and returns the response
func PromptForInput(prompt string) string {
	if !IsInteractive() {
		return ""
	}

	fmt.Print(prompt)
	var input string
	_, _ = fmt.Scanln(&input)
	return strings.TrimSpace(input)
}
