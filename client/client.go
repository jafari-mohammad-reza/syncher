package client

import (
	"github.com/fsnotify/fsnotify"
	"log/slog"
	"path/filepath"
	"sync_server/share"
	"time"
)

type Client struct {
	Cfg          *share.ClientConfig
	HttpListener *HttpListener
	SyncService  *SyncService
	ErrChan      chan error
}

// server will run in background and will have a listening port on custome port to execute commands
func NewClient(cfg *share.ClientConfig) *Client {
	return &Client{
		Cfg:          cfg,
		HttpListener: NewHttpListener(cfg),
		SyncService:  NewSyncService(cfg),
	}
}
func (c *Client) Start() error {
	slog.Info("Client start")
	go func() {
		c.ErrChan <- c.HttpListener.Listen()
	}()
	go c.Sync()
	go c.handleErr()
	select {}
}

func (c *Client) handleErr() {
	for err := range c.ErrChan {
		slog.Error("Client error occurred", "err", err.Error())
	}
}

func (c *Client) Sync() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	for _, dir := range c.Cfg.SyncDirs {
		err := watcher.Add(dir)
		if err != nil {
			panic(err)
		}
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				parentDir := filepath.Dir(event.Name)
				fileName := filepath.Base(event.Name)

				c.SyncService.ChangeChan <- ChangeEvent{
					File: share.ChangeRequestChanges{
						FileName:    fileName,
						ChangeEvent: event.Op.String(),
					},
					Dir:  parentDir,
					Time: time.Now(),
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			c.ErrChan <- err
		}
	}
}
