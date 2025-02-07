package server

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"sync_server/share"
)

type ResponseStatus string

const (
	Success ResponseStatus = "success"
	Failure ResponseStatus = "failure"
)

type Response struct {
	Status ResponseStatus
	Data   []byte
}

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

func (m *MessageHandler) GetHandlerFunc(sbj string) (func(msg *nats.Msg) (*Response, error), error) {
	handlers := map[string]func(msg *nats.Msg) (*Response, error){
		"health": Health,
	}
	handler, ok := handlers[sbj]
	if !ok {
		return nil, fmt.Errorf("Unknown subject %s", sbj)
	}
	return handler, nil
}
func Health(msg *nats.Msg) (*Response, error) {
	return &Response{
		Status: Success,
		Data:   []byte("healthy"),
	}, nil
}
