package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Level represents log levels
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging
type Logger struct {
	level  Level
	output io.Writer
	prefix string
}

// NewLogger creates a new logger instance
func NewLogger(level Level, output io.Writer, prefix string) *Logger {
	return &Logger{
		level:  level,
		output: output,
		prefix: prefix,
	}
}

// NewDefaultLogger creates a logger with default settings
func NewDefaultLogger(prefix string) *Logger {
	return NewLogger(LevelInfo, os.Stdout, prefix)
}

// log writes a log message if the level is appropriate
func (l *Logger) log(level Level, format string, args ...any) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	logLine := fmt.Sprintf("[%s] %s %s%s\n",
		timestamp,
		level.String(),
		l.prefix,
		message)

	_, _ = l.output.Write([]byte(logLine))
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...any) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...any) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...any) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...any) {
	l.log(LevelError, format, args...)
}

// WithPrefix creates a new logger with an additional prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	newPrefix := l.prefix
	if newPrefix != "" {
		newPrefix += " "
	}
	newPrefix += prefix + ": "

	return &Logger{
		level:  l.level,
		output: l.output,
		prefix: newPrefix,
	}
}

// Global logger instance
var defaultLogger = NewDefaultLogger("")

// SetLevel sets the global log level
func SetLevel(level Level) {
	defaultLogger.level = level
}

// SetOutput sets the global log output
func SetOutput(output io.Writer) {
	defaultLogger.output = output
}

// Debug logs a debug message using the global logger
func Debug(format string, args ...any) {
	defaultLogger.Debug(format, args...)
}

// Info logs an info message using the global logger
func Info(format string, args ...any) {
	defaultLogger.Info(format, args...)
}

// Warn logs a warning message using the global logger
func Warn(format string, args ...any) {
	defaultLogger.Warn(format, args...)
}

// Error logs an error message using the global logger
func Error(format string, args ...any) {
	defaultLogger.Error(format, args...)
}

// Fatal logs an error message and exits
func Fatal(format string, args ...any) {
	defaultLogger.Error(format, args...)
	os.Exit(1)
}

// InitializeFromLegacy initializes logging from legacy log package
func InitializeFromLegacy() {
	log.SetOutput(defaultLogger.output)
	log.SetFlags(0) // We handle our own formatting
}
