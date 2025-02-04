package client

import (
	"log/slog"
	"sync_server/share"
	"time"

	"github.com/fsnotify/fsnotify"
)

type ChangeEvent struct {
	FileName string
	Event    fsnotify.Op
	Date     time.Time
}
type Syncher struct {
	ChangeChan chan ChangeEvent
	errChan    chan error
	nc         *share.NatsConn
}

func NewSyncher(cfg *share.Config) *Syncher {
	return &Syncher{
		make(chan ChangeEvent),
		make(chan error),
		share.NewNatsConn(cfg),
	}
}

func (s *Syncher) Start() {
	go s.handleErr()
	go s.handleChange()
	select {}
}

func (s *Syncher) handleChange() {
	for change := range s.ChangeChan {
		slog.Info("Change", "filename", change.FileName, "event", change.Event)
	}
}
func (s *Syncher) handleErr() {
	for err := range s.errChan {
		if err != nil {
			slog.Error("Client syncher error", "err", err.Error())
		}
	}
}
