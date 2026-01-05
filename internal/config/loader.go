package config

import (
	"embed"

	"buddy/internal/errors"
	"buddy/internal/logging"

	"gopkg.in/yaml.v3"
)

//go:embed workflow_states.yaml fast_adapter_states.yaml
var configFS embed.FS

// WorkflowStates represents the workflow state configuration
type WorkflowStates struct {
	WorkflowStates map[string]map[int]string `yaml:"workflow_states"`
}

// FastAdapterStates represents the fast adapter state configuration
type FastAdapterStates struct {
	FastAdapterStates map[string]map[int]string `yaml:"fast_adapter_states"`
}

// ConfigLoader handles loading configuration from embedded YAML files
type ConfigLoader struct {
	logger *logging.Logger
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		logger: logging.NewDefaultLogger("config"),
	}
}

// LoadWorkflowStates loads workflow state mappings from embedded YAML
func (cl *ConfigLoader) LoadWorkflowStates() (map[string]map[int]string, error) {
	data, err := configFS.ReadFile("workflow_states.yaml")
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfiguration,
			"failed to read embedded workflow states config")
	}

	var config WorkflowStates
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfiguration,
			"failed to parse workflow states YAML")
	}

	cl.logger.Info("Loaded workflow states for %d workflow types", len(config.WorkflowStates))
	return config.WorkflowStates, nil
}

// LoadFastAdapterStates loads fast adapter state mappings from embedded YAML
func (cl *ConfigLoader) LoadFastAdapterStates() (map[string]map[int]string, error) {
	data, err := configFS.ReadFile("fast_adapter_states.yaml")
	if err != nil {
		// Fast adapter states are optional, return empty map if file doesn't exist
		cl.logger.Warn("Fast adapter states config not found, using empty mappings")
		return make(map[string]map[int]string), nil
	}

	var config FastAdapterStates
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfiguration,
			"failed to parse fast adapter states YAML")
	}

	cl.logger.Info("Loaded fast adapter states for %d adapter types", len(config.FastAdapterStates))
	return config.FastAdapterStates, nil
}

// ValidateConfig validates that required configuration files are embedded
func (cl *ConfigLoader) ValidateConfig() error {
	// Config files are embedded at build time, no runtime validation needed
	return nil
}

// Global config loader instance
var defaultLoader *ConfigLoader

// InitializeConfigLoader initializes the global config loader
func InitializeConfigLoader() error {
	defaultLoader = NewConfigLoader()
	return defaultLoader.ValidateConfig()
}

// GetWorkflowStates returns workflow states using the global loader
func GetWorkflowStates() (map[string]map[int]string, error) {
	if defaultLoader == nil {
		return nil, errors.Configuration("config loader not initialized")
	}
	return defaultLoader.LoadWorkflowStates()
}

// GetFastAdapterStates returns fast adapter states using the global loader
func GetFastAdapterStates() (map[string]map[int]string, error) {
	if defaultLoader == nil {
		return nil, errors.Configuration("config loader not initialized")
	}
	return defaultLoader.LoadFastAdapterStates()
}
