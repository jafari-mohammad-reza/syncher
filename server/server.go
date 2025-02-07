package server

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"sync_server/share"
)

type Error struct {
	ErrorMsg  string
	IsPublish bool      `json:"is_publish,omitempty"`
	Receiver  *nats.Msg `json:"receiver,omitempty"`
}
type Server struct {
	Cfg            *share.ServerConfig
	ErrChan        chan Error
	Subjects       []string
	NatsConnection *share.NatsConn
}

func NewServer(Cfg *share.ServerConfig) *Server {
	return &Server{
		Cfg,
		make(chan Error),
		[]string{
			"change",
			"sync",
			"health",
			"server-change",
		},
		share.NewNatsConn(Cfg.NatsUrl),
	}
}
func (s *Server) Start() {
	go s.subscribe()
	go s.handleError()
	s.log("Start", "server started successfully.")
	select {}
}

func (s *Server) subscribe() {
	s.log("Subscribe", fmt.Sprintf("subscribing to %+v", s.Subjects))
	for _, sbj := range s.Subjects {
		sub, err := s.NatsConnection.SubscribeToSubject(sbj)
		if err != nil {
			s.ErrChan <- Error{
				ErrorMsg:  fmt.Sprintf("Failed to subscribe to subject %s", sbj),
				IsPublish: true,
				Receiver:  nil,
			}
			return
		}
		go s.handleSubscription(sub)
	}
}
func (s *Server) handleSubscription(sub *nats.Subscription) {
	for msg, err := range sub.Msgs() {
		if err != nil {
			s.ErrChan <- Error{
				ErrorMsg:  fmt.Sprintf("Failed to handling subject %s messages %s", msg.Subject, err.Error()),
				IsPublish: true,
				Receiver:  nil,
			}
			return
		}
		s.log("Server Message", string(msg.Data))
		go s.handleMessage(msg)
	}
}

func (s *Server) handleMessage(msg *nats.Msg) {
	sbj := msg.Subject
	handlers := map[string]func(msg *nats.Msg) (*Response, error){
		"health": Health,
	}
	handler, ok := handlers[sbj]
	if !ok {
		s.ErrChan <- Error{
			ErrorMsg:  fmt.Sprintf("Unknown subject %s", sbj),
			IsPublish: true,
			Receiver:  nil,
		}
		return
	}
	response, err := handler(msg)
	if err != nil {
		s.ErrChan <- Error{
			ErrorMsg:  fmt.Sprintf("Failed to handle message %s. error: %s", sbj, err.Error()),
			IsPublish: true,
			Receiver:  msg,
		}
		return
	}
	resp, err := json.Marshal(response)
	if err != nil {
		s.ErrChan <- Error{
			ErrorMsg:  fmt.Sprintf("Failed to parse message %s. error: %s", sbj, err.Error()),
			IsPublish: true,
			Receiver:  msg,
		}
		return
	}
	err = msg.Respond(resp)
	if err != nil {
		s.ErrChan <- Error{
			ErrorMsg:  fmt.Sprintf("Failed to responde message %s. error: %s", sbj, err.Error()),
			IsPublish: true,
			Receiver:  msg,
		}
		return
	}
}

func (s *Server) handleError() {
	for err := range s.ErrChan {
		s.error(err)
		if err.IsPublish {
			err := err.Receiver.Respond([]byte(err.ErrorMsg))
			if err != nil {
				s.ErrChan <- Error{
					ErrorMsg:  err.Error(),
					IsPublish: false,
					Receiver:  nil,
				}
			}
		}
	}
}
