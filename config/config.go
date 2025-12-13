package config

import (
	"buddy/internal/compiletime"
	"fmt"
)

type Config struct {
	Environment string
}

var globalConfig *Config

func LoadConfig() error {
	// Validate that all required constants were set at build time
	if err := compiletime.ValidateConstants(); err != nil {
		return fmt.Errorf("build-time validation failed: %w", err)
	}

	globalConfig = &Config{Environment: compiletime.BuildEnvironment}
	return nil
}

func GetEnvironment() string {
	if globalConfig != nil {
		return globalConfig.Environment
	}
	return "unknown"
}

func Get(key, defaultValue string) string {
	// Return compile-time constants instead of os.Getenv
	switch key {
	case "JIRA_DOMAIN":
		if compiletime.JiraDomain != "" {
			return compiletime.JiraDomain
		}
	case "JIRA_USERNAME":
		if compiletime.JiraUsername != "" {
			return compiletime.JiraUsername
		}
	case "JIRA_API_KEY":
		if compiletime.JiraApiKey != "" {
			return compiletime.JiraApiKey
		}
	case "DOORMAN_USERNAME":
		if compiletime.DoormanUsername != "" {
			return compiletime.DoormanUsername
		}
	case "DOORMAN_PASSWORD":
		if compiletime.DoormanPassword != "" {
			return compiletime.DoormanPassword
		}
	}
	return defaultValue
}
