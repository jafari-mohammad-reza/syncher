package client

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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
	sbj := fmt.Sprintf("%s-health", ch.client.info.Server.ID)
	cmd := share.NewClientCommand(ch.client.info.ID, nil)
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
	fileName := args[0]
	for _, path := range args[1:] {
		err := share.CheckFileSize(path, 1000)
		if err != nil {
			ch.client.ErrChan <- errors.New(fmt.Sprintf("error checking file size %s: %s", path, err.Error()))
			return
		}
		paths = append(paths, path)
	}
	files, err := share.GetFilesByte(paths)
	if err != nil {
		ch.client.ErrChan <- errors.New(fmt.Sprintf("error getting files bytes: %s", err.Error()))
		return
	}
	arch, err := share.CreateArchive(files)
	if err != nil {
		ch.client.ErrChan <- errors.New(fmt.Sprintf("error CreateArchive: %s", err.Error()))
		return
	}
	cmd := share.NewClientCommand(ch.client.info.ID, map[string]string{"fileName": fmt.Sprintf("%s.tar.gz", fileName)})
	req, _ := json.Marshal(cmd)
	msg, err := ch.client.nc.RequestToSubject(sbj, req, time.Second)
	var repl share.ServerReply
	json.Unmarshal(msg.Data, &repl)
	port := repl.Msg
	conn, err := net.Dial("tcp", ":"+port)
	if err != nil {
		ch.client.ErrChan <- fmt.Errorf("connection failed to port %s: %s", port, err.Error())
	}
	err = binary.Write(conn, binary.BigEndian, int64(len(arch)))
	if err != nil {
		ch.client.ErrChan <- fmt.Errorf("failed to send file size: %s", err.Error())
	}
	n, err := io.CopyN(conn, bytes.NewReader(arch), int64(len(arch)))
	if err != nil {
		ch.client.ErrChan <- errors.New(fmt.Sprintf("error sending file to %d: %s", port, err.Error()))
	}
	fmt.Println("coppied file", n)
	if err != nil {
		ch.client.ErrChan <- errors.New(fmt.Sprintf("error requesting subject %s: %s", sbj, err.Error()))
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
