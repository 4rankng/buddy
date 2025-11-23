package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	loaded = false
)

// Load loads environment variables from .env file (if it exists)
func Load() error {
	if !loaded {
		if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error loading .env file: %w", err)
		}
		loaded = true
	}
	return nil
}

// Get retrieves an environment variable with a default value
func Get(key, defaultValue string) string {
	_ = Load() // Ensure environment is loaded
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetBool retrieves an environment variable as a boolean
func GetBool(key string, defaultValue bool) bool {
	_ = Load()
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetInt retrieves an environment variable as an integer
func GetInt(key string, defaultValue int) int {
	_ = Load()
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetInt64 retrieves an environment variable as a 64-bit integer
func GetInt64(key string, defaultValue int64) int64 {
	_ = Load()
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetDuration retrieves an environment variable as a time.Duration
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	_ = Load()
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetStringSlice retrieves an environment variable as a string slice (comma-separated)
func GetStringSlice(key string, defaultValue []string) []string {
	_ = Load()
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// Require checks that required environment variables are set
func Require(keys ...string) error {
	_ = Load()
	var missing []string

	for _, key := range keys {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required environment variables not set: %s", strings.Join(missing, ", "))
	}

	return nil
}

// Config represents the application configuration
type Config struct {
	Doorman DoormanConfig `json:"doorman"`
	Jira    JiraConfig    `json:"jira"`
	Datadog DatadogConfig `json:"datadog"`
	Storage StorageConfig `json:"storage"`
	App     AppConfig     `json:"app"`
}

// DoormanConfig contains Doorman-specific configuration
type DoormanConfig struct {
	BaseURL       string        `json:"base_url"`
	User          string        `json:"user"`
	Password      string        `json:"password"`
	Timeout       time.Duration `json:"timeout"`
	RetryAttempts int           `json:"retry_attempts"`
	RetryDelay    time.Duration `json:"retry_delay"`
}

// JiraConfig contains Jira-specific configuration
type JiraConfig struct {
	BaseURL       string        `json:"base_url"`
	Email         string        `json:"email"`
	Token         string        `json:"token"`
	Timeout       time.Duration `json:"timeout"`
	RetryAttempts int           `json:"retry_attempts"`
	RetryDelay    time.Duration `json:"retry_delay"`
}

// DatadogConfig contains Datadog-specific configuration
type DatadogConfig struct {
	BaseURL       string        `json:"base_url"`
	APIKey        string        `json:"api_key"`
	AppKey        string        `json:"app_key"`
	Timeout       time.Duration `json:"timeout"`
	RetryAttempts int           `json:"retry_attempts"`
	RetryDelay    time.Duration `json:"retry_delay"`
}

// StorageConfig contains storage-specific configuration
type StorageConfig struct {
	BasePath        string        `json:"base_path"`
	MaxSize         int64         `json:"max_size"`
	MaxFileSize     int64         `json:"max_file_size"`
	Compression     bool          `json:"compression"`
	Encryption      bool          `json:"encryption"`
	DefaultTTL      time.Duration `json:"default_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// AppConfig contains general application configuration
type AppConfig struct {
	Name     string        `json:"name"`
	Version  string        `json:"version"`
	Debug    bool          `json:"debug"`
	LogLevel string        `json:"log_level"`
	Timeout  time.Duration `json:"timeout"`
}

// LoadConfig loads the complete application configuration
func LoadConfig() (*Config, error) {
	_ = Load()

	config := &Config{
		Doorman: DoormanConfig{
			BaseURL:       Get("DOORMAN_BASE_URL", "https://doorman.sgbank.pr"),
			User:          Get("DOORMAN_USERNAME", ""),
			Password:      Get("DOORMAN_PASSWORD", ""),
			Timeout:       GetDuration("DOORMAN_TIMEOUT", 30*time.Second),
			RetryAttempts: GetInt("DOORMAN_RETRY_ATTEMPTS", 3),
			RetryDelay:    GetDuration("DOORMAN_RETRY_DELAY", 1*time.Second),
		},
		Jira: JiraConfig{
			BaseURL:       Get("JIRA_BASE_URL", "https://gxsbank.atlassian.net"),
			Email:         Get("JIRA_EMAIL", ""),
			Token:         Get("JIRA_TOKEN", ""),
			Timeout:       GetDuration("JIRA_TIMEOUT", 30*time.Second),
			RetryAttempts: GetInt("JIRA_RETRY_ATTEMPTS", 3),
			RetryDelay:    GetDuration("JIRA_RETRY_DELAY", 1*time.Second),
		},
		Datadog: DatadogConfig{
			BaseURL:       Get("DATADOG_BASE_URL", "https://api.datadoghq.com"),
			APIKey:        Get("DATADOG_API_KEY", ""),
			AppKey:        Get("DATADOG_APP_KEY", ""),
			Timeout:       GetDuration("DATADOG_TIMEOUT", 30*time.Second),
			RetryAttempts: GetInt("DATADOG_RETRY_ATTEMPTS", 3),
			RetryDelay:    GetDuration("DATADOG_RETRY_DELAY", 1*time.Second),
		},
		Storage: StorageConfig{
			BasePath:        Get("STORAGE_BASE_PATH", "./data"),
			MaxSize:         GetInt64("STORAGE_MAX_SIZE", 100*1024*1024),  // 100MB
			MaxFileSize:     GetInt64("STORAGE_MAX_FILE_SIZE", 1024*1024), // 1MB
			Compression:     GetBool("STORAGE_COMPRESSION", true),
			Encryption:      GetBool("STORAGE_ENCRYPTION", false),
			DefaultTTL:      GetDuration("STORAGE_DEFAULT_TTL", 24*time.Hour),
			CleanupInterval: GetDuration("STORAGE_CLEANUP_INTERVAL", 1*time.Hour),
		},
		App: AppConfig{
			Name:     Get("APP_NAME", "oncall"),
			Version:  Get("APP_VERSION", "1.0.0"),
			Debug:    GetBool("APP_DEBUG", false),
			LogLevel: Get("APP_LOG_LEVEL", "info"),
			Timeout:  GetDuration("APP_TIMEOUT", 30*time.Second),
		},
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errors []string

	// Validate Doorman config
	if c.Doorman.BaseURL == "" {
		errors = append(errors, "DOORMAN_BASE_URL is required")
	}
	if c.Doorman.User == "" {
		errors = append(errors, "DOORMAN_USERNAME is required")
	}
	if c.Doorman.Password == "" {
		errors = append(errors, "DOORMAN_PASSWORD is required")
	}

	// Validate Jira config
	if c.Jira.BaseURL == "" {
		errors = append(errors, "JIRA_BASE_URL is required")
	}
	if c.Jira.Email == "" {
		errors = append(errors, "JIRA_EMAIL is required")
	}
	if c.Jira.Token == "" {
		errors = append(errors, "JIRA_TOKEN is required")
	}

	// Validate Datadog config
	if c.Datadog.APIKey == "" {
		errors = append(errors, "DATADOG_API_KEY is required")
	}
	if c.Datadog.AppKey == "" {
		errors = append(errors, "DATADOG_APP_KEY is required")
	}

	// Validate storage config
	if c.Storage.BasePath == "" {
		errors = append(errors, "STORAGE_BASE_PATH is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
