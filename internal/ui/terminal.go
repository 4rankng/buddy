package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
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

// OpenURL opens the specified URL in the default browser
func OpenURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}
