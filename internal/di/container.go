package di

import (
	"fmt"
	"sync"

	"buddy/internal/clients/datadog"
	"buddy/internal/clients/doorman"
	"buddy/internal/clients/jira"
	"buddy/internal/config"
	"buddy/internal/txn/service"
)

// Container holds all application dependencies
type Container struct {
	doormanClient doorman.DoormanInterface
	jiraClient    jira.JiraInterface
	datadogClient datadog.DatadogInterface
	txnService    *service.TransactionQueryService
	mu            sync.RWMutex
}

// NewContainer creates a new dependency injection container
func NewContainer() *Container {
	return &Container{}
}

// InitializeForEnvironment initializes all services for the given environment
func (c *Container) InitializeForEnvironment(env string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize configuration loader first
	if err := config.InitializeConfigLoader("config"); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Initialize Doorman client
	doormanClient := doorman.NewDoormanClient(env)
	if doormanClient == nil {
		return fmt.Errorf("failed to initialize Doorman client for environment: %s", env)
	}
	c.doormanClient = doormanClient

	// Initialize Jira client
	jiraClient := jira.NewJiraClient(env)
	if jiraClient == nil {
		return fmt.Errorf("failed to initialize Jira client for environment: %s", env)
	}
	c.jiraClient = jiraClient
	jira.Jira = jiraClient // Set global instance for subcommands

	// Initialize Datadog client
	datadogClient := datadog.NewDatadogClient(env)
	if datadogClient == nil {
		return fmt.Errorf("failed to initialize Datadog client for environment: %s", env)
	}
	c.datadogClient = datadogClient

	// Initialize Transaction service
	txnService := service.NewTransactionQueryService(env)
	if txnService == nil {
		return fmt.Errorf("failed to initialize TransactionQueryService for environment: %s", env)
	}
	c.txnService = txnService

	return nil
}

// DoormanClient returns the Doorman client instance
func (c *Container) DoormanClient() doorman.DoormanInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.doormanClient
}

// JiraClient returns the Jira client instance
func (c *Container) JiraClient() jira.JiraInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.jiraClient
}

// DatadogClient returns the Datadog client instance
func (c *Container) DatadogClient() datadog.DatadogInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.datadogClient
}

// TransactionService returns the transaction service instance
func (c *Container) TransactionService() *service.TransactionQueryService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.txnService
}

// ClientSet contains all client dependencies for commands
type ClientSet struct {
	Doorman doorman.DoormanInterface
	Jira    jira.JiraInterface
	Datadog datadog.DatadogInterface
	TxnSvc  *service.TransactionQueryService
}

// GetClientSet returns all clients as a convenient struct
func (c *Container) GetClientSet() *ClientSet {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &ClientSet{
		Doorman: c.doormanClient,
		Jira:    c.jiraClient,
		Datadog: c.datadogClient,
		TxnSvc:  c.txnService,
	}
}

// GetDoorman returns the Doorman client
func (cs *ClientSet) GetDoorman() doorman.DoormanInterface {
	return cs.Doorman
}
