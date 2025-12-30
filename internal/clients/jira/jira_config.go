package jira

import (
	"buddy/internal/config"
	"fmt"
)

// JiraConfig holds environment-specific JIRA configuration
type JiraConfig struct {
	Domain   string
	Auth     JiraAuthInfo
	Project  string // Default project key (e.g., "TS")
	Timeout  int    // Timeout in seconds
	MaxItems int    // Maximum items per request
}

// GetJiraConfig returns the JIRA configuration for the specified environment
func GetJiraConfig(env string) JiraConfig {
	var cfg JiraConfig

	switch env {
	case "sg":
		cfg = JiraConfig{
			Domain: "https://gxsbank.atlassian.net",
			Auth: JiraAuthInfo{
				Username: config.Get("JIRA_USERNAME", ""),
				APIKey:   config.Get("JIRA_API_KEY", ""),
			},
			Project:  "TSE",
			Timeout:  30,
			MaxItems: 50,
		}
	case "my":
		cfg = JiraConfig{
			Domain: "https://gxbank.atlassian.net",
			Auth: JiraAuthInfo{
				Username: config.Get("JIRA_USERNAME", ""),
				APIKey:   config.Get("JIRA_API_KEY", ""),
			},
			Project:  "TS",
			Timeout:  30,
			MaxItems: 50,
		}
	default:
		panic(fmt.Sprintf("country %s is not supported", env))
	}

	return cfg
}
