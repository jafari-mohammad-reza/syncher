package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync_server/share"
	"time"
)

type CommandHandler struct {
	client *Client
	cmap   map[string]func(args []string)
}

func NewCommandHandler(client *Client) *CommandHandler {
	ch := &CommandHandler{
		client: client,
		cmap:   make(map[string]func(args []string)),
	}

	ch.cmap["health"] = ch.HealthCheck
	ch.cmap["upload"] = ch.Upload

	return ch
}

func (ch *CommandHandler) HealthCheck(args []string) {
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

func (ch *CommandHandler) Upload(args []string) {
	// TODO: apply chunk sending later
	sbj := fmt.Sprintf("%s-upload", ch.client.info.Server.ID)
	var paths []string
	for _, path := range args {
		err := share.CheckFileSize(path, 5)
		if err != nil {
			ch.client.ErrChan <- errors.New(fmt.Sprintf("Error checking file size %s: %s", path, err.Error()))
			return
		}
		paths = append(paths, path)
	}
	files, err := share.GetFilesByte(paths)
	if err != nil {
		ch.client.ErrChan <- errors.New(fmt.Sprintf("Error getting files bytes: %s", err.Error()))
		return
	}
	_, err = share.CreateArchive(files)
	if err != nil {
		ch.client.ErrChan <- errors.New(fmt.Sprintf("Error CreateArchive: %s", err.Error()))
		return
	}
	cmd := share.NewClientCommand("upload", nil)
	req, _ := json.Marshal(cmd)
	fmt.Println(string(req))
	err = ch.client.nc.PublishToSubject(sbj, req)
	if err != nil {
		ch.client.ErrChan <- errors.New(fmt.Sprintf("Error requesting subject %s: %s", sbj, err.Error()))
	}
}

func (ch *CommandHandler) ParseCommand(line string) {
	lines := strings.Split(line, " ")
	cmd := lines[0]
	argsArray := lines[1:]
	var args []string
	for _, arg := range argsArray {
		trimmedArg := strings.TrimSpace(arg)
		if trimmedArg != "" {
			args = append(args, trimmedArg)
		}
	}
	if handler, exists := ch.cmap[cmd]; exists {
		handler(args)
	} else {
		err := fmt.Errorf("unknown command '%s'", cmd)
		ch.client.ErrChan <- err
	}
}
