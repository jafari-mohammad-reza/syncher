package server

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

// ICommand interface with GetName method for better logging
type ICommand interface {
	GetName() string
	Execute() (interface{}, error)
}

// Base struct for common command properties
type Command struct {
	args []interface{}
}

// HealthCheckCommand struct embedding Command
type HealthCheckCommand struct {
	Command
}

// GetName provides structured logs for each command
func (c *HealthCheckCommand) GetName() string {
	return "HealthCheck"
}

// Execute runs the command logic
func (c *HealthCheckCommand) Execute() (interface{}, error) {
	return "healthy", nil
}

// ParseCommand dynamically creates commands from messages
func parseCommand(msg *nats.Msg) (ICommand, error) {
	// Simulating command parsing (replace with actual logic)
	cmdName := string(msg.Data) // Assume msg.Data contains command name

	switch cmdName {
	case "health_check":
		return &HealthCheckCommand{}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", cmdName)
	}
}
