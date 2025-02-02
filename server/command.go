package server

import (
	"fmt"
	"sync_server/share"
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
	if len(c.cmd.Args) > 0 {
		for fileName, fileContent := range c.cmd.Args {
			share.UploadFile(fmt.Sprintf("%s_%s", c.cmd.ClientId, fileName), fileContent)
		}
	}
	return "file uploaded", nil
}

// ParseCommand dynamically creates commands from messages
func parseCommand(msg string, cmd *share.ClientCommand) (ICommand, error) {
	switch msg {
	case "health":
		return &HealthCheckCommand{}, nil
	case "upload":
		return &UploadCommand{
			Command: Command{
				cmd: cmd,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", msg)
	}
}
