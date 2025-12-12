package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Environment string
}

var globalConfig *Config

func LoadConfig(env string) error {
	filename := ".env." + env

	if err := loadEnvFile(filename); err != nil {
		return fmt.Errorf("failed to load env file %s: %w", filename, err)
	}

	globalConfig = &Config{Environment: env}
	return nil
}

func GetEnvironment() string {
	if globalConfig != nil {
		return globalConfig.Environment
	}
	return "unknown"
}

func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Warning: failed to close env file %s: %v\n", filename, err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Remove surrounding quotes if present
			if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
				value = strings.Trim(value, `"`)
			}
			
			if err := os.Setenv(key, value); err != nil {
				fmt.Printf("Warning: failed to set env variable %s: %v\n", key, err)
			}
		}
	}
	return nil
}

func Get(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
