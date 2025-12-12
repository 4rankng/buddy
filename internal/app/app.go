package app

import (
	"buddy/config"
	"fmt"
)

type Context struct {
	Environment string
	BinaryName  string
}

func NewContext(binaryName string) (*Context, error) {
	env := "my" // default
	if binaryName == "sgbuddy" {
		env = "sg"
	}

	// Load environment-specific config file
	if err := config.LoadConfig(env); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

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
