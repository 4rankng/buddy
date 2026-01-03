package jira

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

// setAuthHeaders sets authentication headers for JIRA API requests
func (c *JiraClient) setAuthHeaders(req *http.Request) {
	// Validate credentials before setting headers
	if c.config.Auth.Username == "" || c.config.Auth.APIKey == "" {
		c.logger.Warn("Missing JIRA credentials")
		return
	}

	// Create Basic Auth header
	auth := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", c.config.Auth.Username, c.config.Auth.APIKey)),
	)
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))

	// Set standard headers
	req.Header.Set("Accept", "application/json")

	// Only set Content-Type for requests with a body
	if req.Method != "GET" && req.Method != "HEAD" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add user agent for better debugging
	req.Header.Set("User-Agent", "buddy-jira-client/1.0")
}
