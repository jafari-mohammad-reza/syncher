package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync_server/share"
	"time"
)

type Client struct {
	cfg       *share.Config
	info      *share.ClientInfo
	nc        *share.NatsConn
	ErrChan   chan error
	stdinChan chan string
}

var client *Client

func InitClient(cfg *share.Config) {
	info, _ := share.ReadClientInfo()
	nc := share.NewNatsConn(cfg)
	client = &Client{
		cfg,
		info,
		nc,
		make(chan error),
		make(chan string),
	}
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
	select {
	case line := <-c.stdinChan:
		fmt.Println("received", line)

		switch line {
		case "health":
			c.ensureServer()
			sbj := fmt.Sprintf("%s-test", c.info.Server.ID)
			fmt.Println("sbj", sbj)
			cmd := share.NewClientCommand("health_check", nil)
			req, _ := json.Marshal(cmd)
			msg, err := c.nc.RequestToSubject(sbj, req, time.Second)
			if err != nil {
				c.ErrChan <- errors.New(fmt.Sprintf("Error requesting subject %s: %s", sbj, err.Error()))
			} else {
				fmt.Println("msg.Reply", string(msg.Data))
			}
		default:
			c.ErrChan <- errors.New(fmt.Sprintf("Unknown command %s", line))
		}

	}
}
func (c *Client) ensureServer() {
	if c.info.Server == nil {
		c.ErrChan <- errors.New("Please register your server first")
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
	select {
	case err := <-c.ErrChan:
		if err != nil {
			slog.Error("Client HandleError", "err", err.Error())
		}
	}
}
