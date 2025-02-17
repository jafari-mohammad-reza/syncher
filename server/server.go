package server

import (
	"encoding/json"
	"fmt"
	"sync_server/share"

	"github.com/nats-io/nats.go"
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
	Handler        *MessageHandler
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
			"download-file",
		},
		share.NewNatsConn(Cfg.NatsUrl),
		NewMessageHandler(Cfg),
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
	handler, err := s.Handler.GetHandlerFunc(sbj)
	if err != nil {
		s.ErrChan <- Error{
			ErrorMsg:  err.Error(),
			IsPublish: true,
			Receiver:  msg,
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
	if msg.Reply == "" {
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
			resp := share.ServerResponse{
				Status: share.Failure,
				Data:   err.ErrorMsg,
			}
			respJson, parseErr := json.Marshal(resp)
			err.Receiver.Respond(respJson)
			if parseErr != nil {
				s.ErrChan <- Error{
					ErrorMsg:  parseErr.Error(),
					IsPublish: false,
					Receiver:  nil,
				}
			}
		}
	}
}
