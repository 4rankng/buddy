package config

import (
	"buddy/internal/buildinfo"
	"fmt"
)

type Config struct {
	Environment string
}

var globalConfig *Config

func LoadConfig() error {
	// Validate that all required constants were set at build time
	if err := buildinfo.ValidateConstants(); err != nil {
		return fmt.Errorf("build-time validation failed: %w", err)
	}

	globalConfig = &Config{Environment: buildinfo.BuildEnvironment}
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
		if buildinfo.JiraDomain != "" {
			return buildinfo.JiraDomain
		}
	case "JIRA_USERNAME":
		if buildinfo.JiraUsername != "" {
			return buildinfo.JiraUsername
		}
	case "JIRA_API_KEY":
		if buildinfo.JiraApiKey != "" {
			return buildinfo.JiraApiKey
		}
	case "DOORMAN_USERNAME":
		if buildinfo.DoormanUsername != "" {
			return buildinfo.DoormanUsername
		}
	case "DOORMAN_PASSWORD":
		if buildinfo.DoormanPassword != "" {
			return buildinfo.DoormanPassword
		}
	}
	return defaultValue
}
