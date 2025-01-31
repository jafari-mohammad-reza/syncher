package server

import (
	"fmt"
	"strings"
	"sync_server/share"

	"github.com/nats-io/nats.go"
)

type ICommand interface {
	GetName() string
	Execute() (interface{}, error)
}

type Command struct {
	cmd *share.ClientCommand
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

type UploadCommand struct {
	Command
}

func (c *UploadCommand) GetName() string {
	return "Upload"
}

func (c *UploadCommand) Execute() (interface{}, error) {
	fmt.Println("UPLOAD")
	return c.cmd.Args, nil
}

// ParseCommand dynamically creates commands from messages
func parseCommand(msg *nats.Msg) (ICommand, error) {
	cmd, err := share.ParseClientCommand(msg.Data)
	names := strings.Split(msg.Subject, "-")
	if err != nil {
		return nil, err
	}
	name := names[len(names)-1]
	switch name {
	case "health_check":
		return &HealthCheckCommand{}, nil
	case "upload":
		return &UploadCommand{
			Command: Command{
				cmd: cmd,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", name)
	}
}
