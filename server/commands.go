package server

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"math/rand/v2"
	"sync_server/share"
	"time"
)

type MessageHandler struct {
	Cfg             *share.ServerConfig
	NatsConnection  *share.NatsConn
	ReceiverService *ReceiverService
}

func NewMessageHandler(cfg *share.ServerConfig) *MessageHandler {
	return &MessageHandler{
		Cfg:             cfg,
		NatsConnection:  share.NewNatsConn(cfg.NatsUrl),
		ReceiverService: NewReceiverService(cfg),
	}
}

func (m *MessageHandler) GetHandlerFunc(sbj string) (func(msg *nats.Msg) (*share.ServerResponse, error), error) {
	handlers := map[string]func(msg *nats.Msg) (*share.ServerResponse, error){
		"health":        m.Health,
		"change":        m.Change,
		"server-change": m.ServerChange,
	}
	handler, ok := handlers[sbj]
	if !ok {
		return nil, fmt.Errorf("unknown subject %s", sbj)
	}
	return handler, nil
}

func (m *MessageHandler) Health(msg *nats.Msg) (*share.ServerResponse, error) {
	return &share.ServerResponse{
		Status: share.Success,
		Data:   "healthy",
	}, nil
}

func (m *MessageHandler) Change(msg *nats.Msg) (*share.ServerResponse, error) {
	var req share.ChangeRequest
	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		return nil, fmt.Errorf("error parsing change request %s", err.Error())
	}
	res := make(share.ChangeResponse, len(req.Changes))
	for _, change := range req.Changes {
		// if change was create or modify we create a receiver if its delete we remove the file
		if change.ChangeEvent == "delete" {
			// TODO: remove file
			continue
		}
		port := rand.IntN(2000) + 2000
		err := m.ReceiverService.InitReceiver(port, fmt.Sprintf("%s/%s/%s", req.ClientId, req.Dir, change.FileName))
		if err != nil {
			return nil, err
		}
		res[change.FileName] = port
	}
	resBytes, err := json.Marshal(res)
	err = m.recordServerChange(req)
	if err != nil {
		return nil, err
	}
	return &share.ServerResponse{
		Status: share.Success,
		Data:   string(resBytes),
	}, nil
}
func (m *MessageHandler) ServerChange(msg *nats.Msg) (*share.ServerResponse, error) {
	var log ChangeLog
	err := json.Unmarshal(msg.Data, &log)
	if err != nil {
		return nil, fmt.Errorf("error parsing change log %s", err.Error())
	}
	if log.ServerId != m.Cfg.ServerId {
		err := recordChange(log)
		if err != nil {
			return nil, err
		}
	}
	return &share.ServerResponse{
		Status: share.Success,
		Data:   "change applied",
	}, nil
}

func (m *MessageHandler) recordServerChange(req share.ChangeRequest) error {
	changes := []ChangeLogChanges{}
	for _, change := range req.Changes {
		changes = append(changes, ChangeLogChanges{
			FileName: change.FileName,
			Change:   change.ChangeEvent,
		})
	}
	changeLog := ChangeLog{
		ClientId:  req.ClientId,
		ServerId:  m.Cfg.ServerId,
		ChangeDir: req.Dir,
		Changes:   changes,
		Time:      time.Now(),
	}
	err := recordChange(changeLog)
	if err != nil {
		return err
	}
	log, _ := json.Marshal(changeLog)
	err = m.NatsConnection.PublishToSubject("server-change", log)
	if err != nil {
		return err
	}
	return nil
}
