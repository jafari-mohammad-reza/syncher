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

	"github.com/google/uuid"
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
	sbj := fmt.Sprintf("%s-upload", ch.client.info.Server.ID)
	fileName := args[0]
	err := uploadFile(sbj, fileName, ch.client.info.ID, args[1:], ch.client.nc)
	if err != nil {
		ch.client.ErrChan <- err
	}
}

func uploadFile(sbj, fileName string, clientId uuid.UUID, filePaths []string, nc *share.NatsConn) error {

	var paths []string

	for _, path := range filePaths {
		err := share.CheckFileSize(path, 1000)
		if err != nil {
			return errors.New(fmt.Sprintf("error checking file size %s: %s", path, err.Error()))
		}
		paths = append(paths, path)
	}
	files, err := share.GetFilesByte(paths)
	if err != nil {
		return errors.New(fmt.Sprintf("error getting files bytes: %s", err.Error()))

	}
	arch, err := share.CreateArchive(files)
	if err != nil {
		return errors.New(fmt.Sprintf("error CreateArchive: %s", err.Error()))
	}
	cmd := share.NewClientCommand(clientId, map[string][]byte{"fileName": []byte(fmt.Sprintf("%s.tar.gz", fileName))})
	req, _ := json.Marshal(cmd)
	msg, err := nc.RequestToSubject(sbj, req, time.Second)
	var repl share.ServerReply
	json.Unmarshal(msg.Data, &repl)
	port := repl.Msg
	retryCount := 0
	var conn net.Conn
	for retryCount <= 5 {
		conn, err = net.Dial("tcp", ":"+port)
		if err != nil {
			fmt.Errorf("connection failed to port %s: %s", port, err.Error())
		}
		if conn != nil {
			break
		}
		retryCount++
		continue
	}
	err = binary.Write(conn, binary.BigEndian, int64(len(arch)))
	if err != nil {
		return fmt.Errorf("failed to send file size: %s", err.Error())
	}
	n, err := io.CopyN(conn, bytes.NewReader(arch), int64(len(arch)))
	if err != nil {
		return errors.New(fmt.Sprintf("error sending file to %d: %s", port, err.Error()))
	}
	fmt.Println("coppied file", n)
	if err != nil {
		return errors.New(fmt.Sprintf("error requesting subject %s: %s", sbj, err.Error()))
	}
	return nil
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
