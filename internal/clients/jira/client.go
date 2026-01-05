package jira

import (
	"net/http"
	"time"

	"buddy/internal/logging"
)

// JiraClient implements the JiraInterface
type JiraClient struct {
	config     JiraConfig
	httpClient *http.Client
	logger     *logging.Logger
}

// NewJiraClient creates a new JIRA client instance
func NewJiraClient(env string) *JiraClient {
	logger := logging.NewDefaultLogger("jira")

	cfg := GetJiraConfig(env)
	return &JiraClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		logger: logger,
	}
}
