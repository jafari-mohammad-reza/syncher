package server

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log/slog"
	"sync_server/share"
)

type Server struct {
	cfg                  *share.Config
	nc                   *share.NatsConn
	ErrChan              chan *Error
	subscriptionSubjects []string
	subscriptions        chan *nats.Subscription
}

var server *Server

var subjects = []string{
	"test",
	// TODO: new subjects will be added
}

func InitServer(cfg *share.Config) {
	nc := share.NewNatsConn(cfg)
	info, err := share.ReadServerInfo()
	if err != nil {
		panic(err)
	}
	subscriptionSubjects := []string{}
	for _, sub := range subjects {
		subscriptionSubjects = append(subscriptionSubjects, fmt.Sprintf("%s-%s", info.ID, sub))
	}
	server = &Server{
		cfg,
		nc,
		make(chan *Error),
		subscriptionSubjects, // each server will subscribe to its own subject as we can connect many servers to same nats server
		make(chan *nats.Subscription, len(subscriptionSubjects)),
	}
	go server.Start()

	select {}
}
func (s *Server) Start() {
	slog.Info("Server is running")
	_, err := share.ReadServerInfo()
	if err != nil {
		s.ErrChan <- NewServerError("ReadServerInfo error: "+err.Error(), false)
	}
	go server.HandleErrors()
	go s.IniSubscriptions()
}
func (s *Server) IniSubscriptions() {
	slog.Info("Server IniSubscriptions")
	go s.SubHandler()
	for _, subject := range s.subscriptionSubjects {
		sub, err := s.nc.SubscribeToSubject(subject)
		if err != nil {
			s.ErrChan <- NewServerError("SubscribeToSubject error: "+err.Error(), false)
		}
		slog.Info("Subscribed to subject: " + subject)
		s.subscriptions <- sub
	}

}
func (s *Server) SubHandler() {
	slog.Info("Server SubHandler")
	for sub := range s.subscriptions {

		for msg, err := range sub.Msgs() {
			if err != nil {
				s.ErrChan <- NewServerError(fmt.Sprintf("SubHandler %s subject handle error: %s", sub.Subject, err.Error()), false)
			}
			cmd, err := parseCommand(msg)
			if err != nil {
				s.ErrChan <- NewServerError(fmt.Sprintf("parse %s subject command error: %s", sub.Subject, err.Error()), false)
			}
			resp, err := cmd.Execute()
			if err != nil {
				s.ErrChan <- NewServerError(fmt.Sprintf("execute %s subject %s command error: %s", sub.Subject, cmd.GetName(), err.Error()), false)
				msg.Respond([]byte(err.Error()))
			} else {
				msg.Respond([]byte(resp.(string)))
			}
		}
	}
}
func (s *Server) HandleErrors() {
	for err := range s.ErrChan { // Continuously listen for errors
		if err != nil {
			s.SaveErrorLog(err)
		}
	}
}
