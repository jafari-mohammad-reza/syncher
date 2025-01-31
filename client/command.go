package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync_server/share"
	"time"
)

type CommandHandler struct {
	client *Client
	cmap   map[string]func()
}

func NewCommandHandler(client *Client) *CommandHandler {
	ch := &CommandHandler{
		client: client,
		cmap:   make(map[string]func()),
	}

	ch.cmap["health"] = ch.HealthCheck

	return ch
}

func (ch *CommandHandler) HealthCheck() {
	sbj := fmt.Sprintf("%s-test", ch.client.info.Server.ID)
	cmd := share.NewClientCommand("health_check", nil)
	req, _ := json.Marshal(cmd)
	msg, err := ch.client.nc.RequestToSubject(sbj, req, time.Second)
	if err != nil {
		ch.client.ErrChan <- errors.New(fmt.Sprintf("Error requesting subject %s: %s", sbj, err.Error()))
	} else {
		fmt.Println(string(msg.Data))
	}
}

func (ch *CommandHandler) ParseCommand(line string) {
	if handler, exists := ch.cmap[line]; exists {
		handler()
	} else {
		err := fmt.Errorf("unknown command '%s'", line)
		ch.client.ErrChan <- err
	}
}
