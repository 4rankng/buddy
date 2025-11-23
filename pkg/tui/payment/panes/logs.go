package panes

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarn    LogLevel = "WARN"
	LogLevelError   LogLevel = "ERROR"
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelSuccess LogLevel = "SUCCESS"
)

// LogEntry represents a log message in the system
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
}

// LogsPane manages the logs pane
type LogsPane struct {
	Logs     []LogEntry
	MaxLines int
	Scroll   int
}

// NewLogsPane creates a new logs pane
func NewLogsPane(maxLines int) *LogsPane {
	return &LogsPane{
		Logs:     generateInitialLogs(),
		MaxLines: maxLines,
		Scroll:   0,
	}
}

// generateInitialLogs creates initial log entries
func generateInitialLogs() []LogEntry {
	// Fixed timestamps to match the mockup
	now := time.Now()

	return []LogEntry{
		{
			Timestamp: now,
			Level:     LogLevelInfo,
			Message:   "Jira module loaded 3 tickets.",
		},
		{
			Timestamp: now.Add(1 * time.Second),
			Level:     LogLevelInfo,
			Message:   "Datadog client connected. Waiting for input...",
		},
		{
			Timestamp: now.Add(2 * time.Second),
			Level:     LogLevelInfo,
			Message:   "Payment dashboard initialized successfully.",
		},
	}
}

// AddLog adds a new log entry
func (p *LogsPane) AddLog(level LogLevel, message string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}

	p.Logs = append(p.Logs, entry)

	// Keep only the most recent logs
	if len(p.Logs) > 100 {
		p.Logs = p.Logs[len(p.Logs)-100:]
	}

	// Auto-scroll to bottom
	p.Scroll = len(p.Logs) - p.MaxLines
	if p.Scroll < 0 {
		p.Scroll = 0
	}
}

// ScrollUp scrolls up in the log view
func (p *LogsPane) ScrollUp() {
	if p.Scroll > 0 {
		p.Scroll--
	}
}

// ScrollDown scrolls down in the log view
func (p *LogsPane) ScrollDown() {
	maxScroll := len(p.Logs) - p.MaxLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	if p.Scroll < maxScroll {
		p.Scroll++
	}
}

// ScrollToBottom scrolls to the bottom of the logs
func (p *LogsPane) ScrollToBottom() {
	maxScroll := len(p.Logs) - p.MaxLines
	if maxScroll < 0 {
		p.Scroll = 0
	} else {
		p.Scroll = maxScroll
	}
}

// GetVisibleLogs returns the currently visible log entries
func (p *LogsPane) GetVisibleLogs() []LogEntry {
	if len(p.Logs) == 0 {
		return []LogEntry{}
	}

	start := p.Scroll
	end := start + p.MaxLines

	if end > len(p.Logs) {
		end = len(p.Logs)
	}

	if start < 0 {
		start = 0
	}

	if start >= len(p.Logs) {
		return []LogEntry{}
	}

	return p.Logs[start:end]
}

// formatLogEntry formats a single log entry
func (p *LogsPane) formatLogEntry(entry LogEntry) string {
	timestampStr := entry.Timestamp.Format("15:04:05")
	levelStyle := p.getLevelStyle(entry.Level)
	levelStr := levelStyle.Render(string(entry.Level))
	messageStr := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#F0F0F0"}).Render(entry.Message)

	return fmt.Sprintf("[%s] %s: %s",
		lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#B0B0B0"}).Italic(true).Render(timestampStr),
		levelStr,
		messageStr,
	)
}

// getLevelStyle returns the appropriate style for the log level
func (p *LogsPane) getLevelStyle(level LogLevel) lipgloss.Style {
	switch level {
	case LogLevelInfo:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0088CC", Dark: "#00BFFF"})
	case LogLevelWarn:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#CC8800", Dark: "#FFA500"})
	case LogLevelError:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#CC0000", Dark: "#FF6B6B"}).Bold(true)
	case LogLevelDebug:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#808080"})
	case LogLevelSuccess:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#008800", Dark: "#00FF7F"})
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0088CC", Dark: "#00BFFF"})
	}
}

// Clear clears all logs
func (p *LogsPane) Clear() {
	p.Logs = []LogEntry{}
	p.Scroll = 0
}

// GetLogStats returns statistics about the logs
func (p *LogsPane) GetLogStats() map[string]int {
	stats := map[string]int{
		"total":   len(p.Logs),
		"info":    0,
		"warn":    0,
		"error":   0,
		"debug":   0,
		"success": 0,
	}

	for _, log := range p.Logs {
		switch log.Level {
		case LogLevelInfo:
			stats["info"]++
		case LogLevelWarn:
			stats["warn"]++
		case LogLevelError:
			stats["error"]++
		case LogLevelDebug:
			stats["debug"]++
		case LogLevelSuccess:
			stats["success"]++
		}
	}

	return stats
}
