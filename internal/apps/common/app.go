package common

import (
	"buddy/internal/config"
	"fmt"
)

type Context struct {
	Environment string
	BinaryName  string
}

func NewContext(binaryName string) (*Context, error) {
	// Load config which will set the environment from build-time constants
	if err := config.LoadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Environment is now determined at build time
	env := config.GetEnvironment()

	return &Context{
		Environment: env,
		BinaryName:  binaryName,
	}, nil
}

func (c *Context) GetPrefix() string {
	if c.Environment == "sg" {
		return "[SG] "
	}
	return "[MY] "
}

func (c *Context) IsSG() bool {
	return c.Environment == "sg"
}

func (c *Context) IsMY() bool {
	return c.Environment == "my"
}
