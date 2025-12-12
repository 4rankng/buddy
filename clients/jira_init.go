package clients

import (
	"errors"
	"sync"
)

var (
	jiraOnce sync.Once
	jiraEnv  string
)

// Jira returns the global JIRA client instance
// This follows the singleton pattern used by Doorman
var Jira JiraInterface

// NewJiraClientSingleton initializes the global JIRA client with the specified environment
func NewJiraClientSingleton(env string) error {
	// Only initialize once
	var initErr error
	jiraOnce.Do(func() {
		if Jira != nil {
			return // Already initialized
		}

		// Validate environment
		if env != "my" && env != "sg" {
			initErr = errors.New("invalid environment: must be 'my' or 'sg'")
			return
		}

		// Create new client instance
		Jira = NewJiraClient(env)
		jiraEnv = env
	})

	return initErr
}

// GetJiraClient returns the initialized JIRA client instance
// Deprecated: Use clients.Jira directly after initialization
func GetJiraClient() (JiraInterface, error) {
	if Jira == nil {
		return nil, errors.New("jira client not initialized. Call NewJiraClientSingleton first")
	}
	return Jira, nil
}

// GetJiraEnvironment returns the environment the client was initialized with
func GetJiraEnvironment() string {
	return jiraEnv
}

// ResetJiraSingleton resets the JIRA client singleton
// This should only be used in tests
func ResetJiraSingleton() {
	Jira = nil
	jiraOnce = sync.Once{}
	jiraEnv = ""
}
