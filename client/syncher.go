package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
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
	clientInfo *share.ClientInfo
}

func NewSyncher(cfg *share.ClientConfig, clientInfo *share.ClientInfo) *Syncher {
	return &Syncher{
		make(chan ChangeEvent),
		make(chan error),
		share.NewNatsConn(cfg.NatsUrl),
		clientInfo,
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
		s.saveChange(&change)
		// TODO: get upload link and upload
	}
}
func (s *Syncher) saveChange(change *ChangeEvent) {
	var event share.ChangeEvent
	switch {
	case change.Event.Has(fsnotify.Create):
		event = share.Create
	case change.Event.Has(fsnotify.Remove):
		event = share.Delete
	case change.Event.Has(fsnotify.Rename), change.Event.Has(fsnotify.Write), change.Event.Has(fsnotify.Chmod):
		event = share.Modify
	}

	changeLog := share.ChangeLog{
		ClientId: s.clientInfo.ID,
		FileName: change.FileName,
		Date:     change.Date,
		Event:    event,
	}
	fmt.Println(changeLog)
	s.saveChangeLog(changeLog)
	s.syncChange(changeLog)
}

func (s *Syncher) syncChange(cl share.ChangeLog) {
	paths := strings.Split(cl.FileName, "/")
	fileName := paths[len(paths)-1]
	sbj := fmt.Sprintf("%s-sync", s.clientInfo.Server.ID)
	err := uploadFile(sbj, fileName, s.clientInfo.ID, []string{cl.FileName}, s.nc)
	if err != nil {
		s.errChan <- err
	}
}

func (s *Syncher) saveChangeLog(changeLog share.ChangeLog) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		s.errChan <- err
		return
	}

	ph := path.Join(homeDir, ".syncher", "changes.json")

	var logs []share.ChangeLog
	data, err := os.ReadFile(ph)
	if err == nil && len(data) > 0 {
		err = json.Unmarshal(data, &logs)
		if err != nil {
			fmt.Println("Warning: Could not parse existing JSON file. Creating a new one.")
			logs = []share.ChangeLog{}
		}
	}
	logs = append(logs, changeLog)

	newData, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		s.errChan <- err
		return
	}
	err = os.WriteFile(ph, newData, 0644)
	if err != nil {
		s.errChan <- err
		return
	}

	fmt.Println("Saved change:", changeLog)
}

func (s *Syncher) handleErr() {
	for err := range s.errChan {
		if err != nil {
			slog.Error("Client syncher error", "err", err.Error())
		}
	}
}
