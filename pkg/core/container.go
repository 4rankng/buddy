package core

import (
	"fmt"
	"sync"

	"oncall/pkg/config"
	"oncall/pkg/modules"
	"oncall/pkg/ports"
)

// Container manages dependencies and provides access to all modules
type Container struct {
	config   *config.Config
	doorman  ports.DoormanPort
	jira     ports.JiraPort
	datadog  ports.DatadogPort
	storage  ports.StoragePort

	mu sync.RWMutex
	initialized bool
}

var (
	instance     *Container
	instanceOnce sync.Once
)

// GetContainer returns the singleton container instance
func GetContainer() *Container {
	instanceOnce.Do(func() {
		instance = &Container{}
	})
	return instance
}

// Initialize initializes all modules with the provided configuration
func (c *Container) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	c.config = cfg

	// Initialize modules
	if err := c.initializeModules(); err != nil {
		return fmt.Errorf("failed to initialize modules: %w", err)
	}

	c.initialized = true
	return nil
}

// initializeModules creates and initializes all module instances
func (c *Container) initializeModules() error {
	var err error

	// Initialize Doorman module
	c.doorman, err = modules.NewDoormanModule(c.config.Doorman)
	if err != nil {
		return fmt.Errorf("failed to initialize doorman module: %w", err)
	}

	// Initialize Jira module
	c.jira, err = modules.NewJiraModule(c.config.Jira)
	if err != nil {
		return fmt.Errorf("failed to initialize jira module: %w", err)
	}

	// Initialize Datadog module
	c.datadog, err = modules.NewDatadogModule(c.config.Datadog)
	if err != nil {
		return fmt.Errorf("failed to initialize datadog module: %w", err)
	}

	// Initialize Storage module
	c.storage, err = modules.NewStorageModule(c.config.Storage)
	if err != nil {
		return fmt.Errorf("failed to initialize storage module: %w", err)
	}

	return nil
}

// Config returns the application configuration
func (c *Container) Config() *config.Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// Doorman returns the Doorman module
func (c *Container) Doorman() ports.DoormanPort {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.doorman
}

// Jira returns the Jira module
func (c *Container) Jira() ports.JiraPort {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.jira
}

// Datadog returns the Datadog module
func (c *Container) Datadog() ports.DatadogPort {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.datadog
}

// Storage returns the Storage module
func (c *Container) Storage() ports.StoragePort {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.storage
}

// HealthCheck performs health checks on all modules
func (c *Container) HealthCheck() map[string]error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make(map[string]error)

	if c.doorman != nil {
		if err := c.doorman.HealthCheck(); err != nil {
			results["doorman"] = err
		} else {
			results["doorman"] = nil
		}
	}

	if c.jira != nil {
		if err := c.jira.HealthCheck(); err != nil {
			results["jira"] = err
		} else {
			results["jira"] = nil
		}
	}

	if c.datadog != nil {
		if err := c.datadog.HealthCheck(); err != nil {
			results["datadog"] = err
		} else {
			results["datadog"] = nil
		}
	}

	if c.storage != nil {
		if err := c.storage.HealthCheck(); err != nil {
			results["storage"] = err
		} else {
			results["storage"] = nil
		}
	}

	return results
}

// IsInitialized returns whether the container has been initialized
func (c *Container) IsInitialized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized
}

// Reset resets the container (mainly for testing)
func (c *Container) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config = nil
	c.doorman = nil
	c.jira = nil
	c.datadog = nil
	c.storage = nil
	c.initialized = false
}

// NewContainer creates a new container and initializes it
func NewContainer() (*Container, error) {
	container := &Container{}
	if err := container.Initialize(); err != nil {
		return nil, err
	}
	return container, nil
}

// MustGetContainer returns the container and panics if initialization fails
func MustGetContainer() *Container {
	container := GetContainer()
	if err := container.Initialize(); err != nil {
		panic(fmt.Sprintf("Failed to initialize container: %v", err))
	}
	return container
}