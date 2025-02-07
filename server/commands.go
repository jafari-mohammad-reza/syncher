package server

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"math/rand/v2"
	"sync_server/share"
)

type MessageHandler struct {
	Cfg            *share.ServerConfig
	NatsConnection *share.NatsConn
}

func NewMessageHandler(cfg *share.ServerConfig) *MessageHandler {
	return &MessageHandler{
		Cfg:            cfg,
		NatsConnection: share.NewNatsConn(cfg.NatsUrl),
	}
}

func (m *MessageHandler) GetHandlerFunc(sbj string) (func(msg *nats.Msg) (*share.ServerResponse, error), error) {
	handlers := map[string]func(msg *nats.Msg) (*share.ServerResponse, error){
		"health": m.Health,
		"change": m.Change,
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
		Data:   []byte("healthy"),
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
		port := rand.IntN(1000) + 1000
		err := InitReceiver(port, fmt.Sprintf("%s/%s/%s", req.ClientId, req.Dir, change.FileName))
		if err != nil {
			return nil, err
		}
		res[change.FileName] = port
	}
	resBytes, err := json.Marshal(res)
	return &share.ServerResponse{
		Status: share.Success,
		Data:   resBytes,
	}, nil
}
