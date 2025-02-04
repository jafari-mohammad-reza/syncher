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

	"github.com/fsnotify/fsnotify"
)

type Client struct {
	cfg            *share.Config
	info           *share.ClientInfo
	nc             *share.NatsConn
	ErrChan        chan error
	stdinChan      chan string
	commandHandler *CommandHandler
	syncher        *Syncher
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
		NewSyncher(cfg, info),
	}
	handler := NewCommandHandler(client)
	client.commandHandler = handler
	go client.ReadInput()
	go client.Sync()
	go client.HandleError()
	go client.syncher.Start()
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
func (c *Client) Sync() {
	err := share.InitClientSyncherDir()
	if err != nil {
		c.ErrChan <- err
		return
	}
	info, _ := share.ReadClientInfo()
	fmt.Println("Start syncing")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	for _, sh := range info.SharePath {
		err := watcher.Add(sh)
		if err != nil {
			panic(err)
		}
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			fmt.Println("event", event, ok)
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				c.syncher.ChangeChan <- ChangeEvent{
					FileName: event.Name,
					Date:     time.Now(),
					Event:    event.Op,
				}
			}
		case err, ok := <-watcher.Errors:
			fmt.Println("err", err, ok)
			if !ok {
				return
			}
			c.ErrChan <- err
		}
	}

}
func (c *Client) HandleError() {
	for err := range c.ErrChan {
		if err != nil {
			slog.Error("Client HandleError", "err", err.Error())
		}
	}
}
