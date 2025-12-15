package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GetDescriptionLength returns the configured description truncate length
// Defaults to 100 characters if not set
func GetDescriptionLength() int {
	lengthStr := os.Getenv("JIRA_DESC_LENGTH")
	if lengthStr == "" {
		return 100
	}

	if length, err := strconv.Atoi(lengthStr); err == nil && length > 0 {
		return length
	}

	return 100
}

// TruncateText truncates text to the specified length, adding "..." if truncated
func TruncateText(text string, maxLen int) string {
	if text == "" {
		return ""
	}

	// Clean up newlines and extra spaces
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.Join(strings.Fields(text), " ")

	if len(text) <= maxLen {
		return text
	}

	if maxLen <= 3 {
		return text[:maxLen]
	}

	return text[:maxLen-3] + "..."
}

// WordWrap wraps text to the specified width, preserving word boundaries
func WordWrap(text string, width int) []string {
	if text == "" || width <= 0 {
		return []string{}
	}

	var lines []string
	words := strings.Fields(text)

	currentLine := ""
	currentLength := 0

	for _, word := range words {
		wordLength := len(word)

		if currentLength == 0 {
			currentLine = word
			currentLength = wordLength
		} else if currentLength+1+wordLength <= width {
			currentLine += " " + word
			currentLength += 1 + wordLength
		} else {
			lines = append(lines, currentLine)
			currentLine = word
			currentLength = wordLength
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// FormatLink wraps text with OSC-8 hyperlink if enabled, otherwise returns text
func FormatLink(text, url string, enabled bool) string {
	if !enabled {
		return fmt.Sprintf("%s (%s)", text, url)
	}

	// ANSI hyperlink escape sequence: \x1b]8;;URL\x1b\URL Text\x1b]8;;\x1b\
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}
