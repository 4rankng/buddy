package buildinfo

import (
	"fmt"
)

// These variables will be set at build time using ldflags
var (
	// JIRA configuration
	JiraDomain   string
	JiraUsername string
	JiraApiKey   string

	// Doorman configuration
	DoormanUsername string
	DoormanPassword string

	// Build information
	BuildEnvironment string
)

// ValidateConstants ensures all required constants are set
func ValidateConstants() error {
	if JiraDomain == "" {
		return fmt.Errorf("JIRA_DOMAIN not set at build time")
	}
	if JiraUsername == "" {
		return fmt.Errorf("JIRA_USERNAME not set at build time")
	}
	if JiraApiKey == "" {
		return fmt.Errorf("JIRA_API_KEY not set at build time")
	}
	if DoormanUsername == "" {
		return fmt.Errorf("DOORMAN_USERNAME not set at build time")
	}
	if DoormanPassword == "" {
		return fmt.Errorf("DOORMAN_PASSWORD not set at build time")
	}
	if BuildEnvironment == "" {
		return fmt.Errorf("BUILD_ENVIRONMENT not set at build time")
	}
	return nil
}
