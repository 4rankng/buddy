package config

import (
	"fmt"
	"os"
	"path/filepath"

	"buddy/internal/errors"
	"buddy/internal/logging"

	"gopkg.in/yaml.v3"
)

// WorkflowStates represents the workflow state configuration
type WorkflowStates struct {
	WorkflowStates map[string]map[int]string `yaml:"workflow_states"`
}

// FastAdapterStates represents the fast adapter state configuration
type FastAdapterStates struct {
	FastAdapterStates map[string]map[int]string `yaml:"fast_adapter_states"`
}

// ConfigLoader handles loading configuration from YAML files
type ConfigLoader struct {
	configDir string
	logger    *logging.Logger
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(configDir string) *ConfigLoader {
	return &ConfigLoader{
		configDir: configDir,
		logger:    logging.NewDefaultLogger("config"),
	}
}

// LoadWorkflowStates loads workflow state mappings from YAML
func (cl *ConfigLoader) LoadWorkflowStates() (map[string]map[int]string, error) {
	configPath := filepath.Join(cl.configDir, "workflow_states.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfiguration,
			fmt.Sprintf("failed to read workflow states config: %s", configPath))
	}

	var config WorkflowStates
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfiguration,
			"failed to parse workflow states YAML")
	}

	cl.logger.Info("Loaded workflow states for %d workflow types", len(config.WorkflowStates))
	return config.WorkflowStates, nil
}

// LoadFastAdapterStates loads fast adapter state mappings from YAML
func (cl *ConfigLoader) LoadFastAdapterStates() (map[string]map[int]string, error) {
	configPath := filepath.Join(cl.configDir, "fast_adapter_states.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Fast adapter states are optional, return empty map if file doesn't exist
		if os.IsNotExist(err) {
			cl.logger.Warn("Fast adapter states config not found, using empty mappings")
			return make(map[string]map[int]string), nil
		}
		return nil, errors.Wrap(err, errors.ErrorTypeConfiguration,
			fmt.Sprintf("failed to read fast adapter states config: %s", configPath))
	}

	var config FastAdapterStates
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfiguration,
			"failed to parse fast adapter states YAML")
	}

	cl.logger.Info("Loaded fast adapter states for %d adapter types", len(config.FastAdapterStates))
	return config.FastAdapterStates, nil
}

// ValidateConfig validates that required configuration files exist
func (cl *ConfigLoader) ValidateConfig() error {
	requiredFiles := []string{"workflow_states.yaml"}

	for _, file := range requiredFiles {
		configPath := filepath.Join(cl.configDir, file)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return errors.Configuration(fmt.Sprintf("required config file missing: %s", configPath))
		}
	}

	return nil
}

// GetConfigDir returns the configuration directory path
func (cl *ConfigLoader) GetConfigDir() string {
	return cl.configDir
}

// Global config loader instance
var defaultLoader *ConfigLoader

// InitializeConfigLoader initializes the global config loader
func InitializeConfigLoader(configDir string) error {
	defaultLoader = NewConfigLoader(configDir)
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
