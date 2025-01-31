package client

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync_server/share"
	"time"
)

type Client struct {
	cfg            *share.Config
	info           *share.ClientInfo
	nc             *share.NatsConn
	ErrChan        chan error
	stdinChan      chan string
	commandHandler *CommandHandler
}

var client *Client

func InitClient(cfg *share.Config) {
	info, _ := share.ReadClientInfo()
	nc := share.NewNatsConn(cfg)
	ErrChan := make(chan error)
	client = &Client{
		cfg,
		info,
		nc,
		ErrChan,
		make(chan string),
		nil,
	}
	handler := NewCommandHandler(client)
	client.commandHandler = handler
	go client.ReadInput()
	go client.Sync()
	go client.HandleError()
	select {}
}
func (c *Client) ReadInput() {
	scanner := bufio.NewScanner(os.Stdin)
	go client.HandleInput()
	for scanner.Scan() {
		line := scanner.Text()
		c.stdinChan <- line
	}
	if err := scanner.Err(); err != nil {
		c.ErrChan <- errors.New(fmt.Sprintf("Error reading stdin %s", err.Error()))
	}
}

func (c *Client) HandleInput() {
	for line := range c.stdinChan {
		if strings.Trim(line, " ") != "" {
			c.commandHandler.ParseCommand(line)
		}
	}
}
func (c *Client) ensureServer() {
	if c.info.Server == nil {
		c.ErrChan <- errors.New("please register your server first")
		time.Sleep(time.Second)
		defer os.Exit(1)
	}
}

// Sync this method will watch to client shared dirs and client settings
func (c *Client) Sync() error {
	share.InitServerSyncherDir()
	slog.Info("InitClient")
	return nil
}
func (c *Client) HandleError() {
	for err := range c.ErrChan {
		if err != nil {
			slog.Error("Client HandleError", "err", err.Error())
		}
	}
}
