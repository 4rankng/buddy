package jira

import (
	"errors"
)

var (
	jiraEnv string
)

// Jira returns the global JIRA client instance
// This follows the singleton pattern used by Doorman
var Jira JiraInterface

// GetJiraClient returns the initialized JIRA client instance
// Deprecated: Use clients.Jira directly after initialization
func GetJiraClient() (JiraInterface, error) {
	if Jira == nil {
		return nil, errors.New("jira client not initialized. Call NewJiraClient first")
	}
	return Jira, nil
}

// GetJiraEnvironment returns the environment the client was initialized with
func GetJiraEnvironment() string {
	return jiraEnv
}
