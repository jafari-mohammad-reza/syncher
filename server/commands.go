package server

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync_server/share"
	"time"

	"github.com/nats-io/nats.go"
)

type MessageHandler struct {
	Cfg             *share.ServerConfig
	NatsConnection  *share.NatsConn
	ReceiverService *ReceiverService
	ChangeStorage   Storage
	fileStorage     FileStorage
}

func NewMessageHandler(cfg *share.ServerConfig) *MessageHandler {
	return &MessageHandler{
		Cfg:             cfg,
		NatsConnection:  share.NewNatsConn(cfg.NatsUrl),
		ReceiverService: NewReceiverService(cfg),
		ChangeStorage:   NewChangeStorage(),
		fileStorage:     NewMinIoService(cfg),
	}
}

func (m *MessageHandler) GetHandlerFunc(sbj string) (func(msg *nats.Msg) (*share.ServerResponse, error), error) {
	handlers := map[string]func(msg *nats.Msg) (*share.ServerResponse, error){
		"health":        m.Health,
		"change":        m.Change,
		"server-change": m.ServerChange,
		"sync":          m.Sync,
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

const (
	minPort = 3000
	maxPort = 8000
	retries = 10
)

func (m *MessageHandler) Change(msg *nats.Msg) (*share.ServerResponse, error) {
	var req share.ChangeRequest
	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		return nil, fmt.Errorf("error parsing change request %s", err.Error())
	}
	res := make(share.ChangeResponse, len(req.Changes))
	for _, change := range req.Changes {
		if change.ChangeEvent == "REMOVE" {
			// TODO: remove file
			err := m.fileStorage.RemoveFile(fmt.Sprintf("%s%s/%s", req.ClientId, req.Dir, change.FileName))
			if err != nil {
				return nil, fmt.Errorf("error removing file %s", change.FileName)
			}
			continue
		}
		var port int
		for attempt := 0; attempt < retries; attempt++ {
			port = rand.IntN(maxPort-minPort) + minPort
			err := m.ReceiverService.InitReceiver(port, fmt.Sprintf("%s%s/%s", req.ClientId, req.Dir, change.FileName))

			if err == nil {
				break
			}
			if !strings.Contains(err.Error(), "address already in use") {
				return nil, err
			}

			if attempt == retries-1 {
				return nil, fmt.Errorf("failed to find an available port after %d attempts", retries)
			}
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
		err := recordChangeLog(log)
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
	err := recordChangeLog(changeLog)
	if err != nil {
		return err
	}
	log, _ := json.Marshal(changeLog)
	m.ChangeStorage.Set(req.ClientId, log)
	err = m.NatsConnection.PublishToSubject("server-change", log)
	if err != nil {
		return err
	}
	return nil
}

func (m *MessageHandler) Sync(msg *nats.Msg) (*share.ServerResponse, error) {
	var req share.ClientRequest
	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		return nil, fmt.Errorf("error parsing sync request %s", err.Error())
	}
	var res []share.SyncResponse
	clientChanges, err := m.ChangeStorage.Get(req.ClientId)
	if err != nil {
		return nil, fmt.Errorf("error fetching client changes %s", err.Error())
	}
	changemap := map[string][]share.ChangeRequestChanges{}
	for _, ch := range clientChanges.([]ChangeLog) {
		respChanges := []share.ChangeRequestChanges{}
		for _, a := range ch.Changes {
			respChanges = append(respChanges, share.ChangeRequestChanges{
				FileName:    a.FileName,
				ChangeEvent: a.Change,
			})
		}
		changemap[ch.ChangeDir] = append(changemap[ch.ChangeDir], respChanges...)
	}
	for dir, changes := range changemap {
		res = append(res, share.SyncResponse{Dir: dir, Changes: changes})
	}
	resBytes, _ := json.Marshal(res)
	return &share.ServerResponse{
		Status: share.Success,
		Data:   string(resBytes),
	}, nil
}
