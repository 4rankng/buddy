package jira

import (
	"buddy/internal/config"
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
	switch env {
	case "sg":
		return JiraConfig{
			Domain: config.Get("JIRA_DOMAIN", "https://gxsbank.atlassian.net"),
			Auth: JiraAuthInfo{
				Username: config.Get("JIRA_USERNAME", ""),
				APIKey:   config.Get("JIRA_API_KEY", ""),
			},
			Project:  "TSE",
			Timeout:  30,
			MaxItems: 50,
		}
	case "my":
		return JiraConfig{
			Domain: config.Get("JIRA_DOMAIN", "https://gxbank.atlassian.net"),
			Auth: JiraAuthInfo{
				Username: config.Get("JIRA_USERNAME", ""),
				APIKey:   config.Get("JIRA_API_KEY", ""),
			},
			Project:  "TS",
			Timeout:  30,
			MaxItems: 50,
		}
	default:
		// Default to Malaysia if environment not found
		return JiraConfig{
			Domain: config.Get("JIRA_DOMAIN", "https://gxbank.atlassian.net"),
			Auth: JiraAuthInfo{
				Username: config.Get("JIRA_USERNAME", ""),
				APIKey:   config.Get("JIRA_API_KEY", ""),
			},
			Project:  "TS",
			Timeout:  30,
			MaxItems: 50,
		}
	}
}
