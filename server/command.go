package server

import (
	"fmt"
	"sync_server/share"

	"github.com/nats-io/nats.go"
)

type ICommand interface {
	GetName() string
	Execute() (interface{}, error)
}

type Command struct {
	args []interface{}
}

type HealthCheckCommand struct {
	Command
}

func (c *HealthCheckCommand) GetName() string {
	return "HealthCheck"
}

func (c *HealthCheckCommand) Execute() (interface{}, error) {
	return "healthy", nil
}

// ParseCommand dynamically creates commands from messages
func parseCommand(msg *nats.Msg) (ICommand, error) {
	cmd, err := share.ParseClientCommand(msg.Data)
	if err != nil {
		return nil, err
	}
	switch cmd.Name {
	case "health_check":
		return &HealthCheckCommand{}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", cmd.Name)
	}
}
