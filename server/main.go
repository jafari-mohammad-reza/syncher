package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync_server/share"

	"github.com/nats-io/nats.go"
)

type Server struct {
	cfg                  *share.Config
	info                 *share.ServerInfo
	nc                   *share.NatsConn
	ErrChan              chan *Error
	subscriptionSubjects []string
	commandHandler       *CommandHandler
}

var server *Server

var subjects = []string{
	"upload",
	"pull",
	"sync",
	"health",
	// TODO: new subjects will be added
}

func InitServer(cfg *share.Config) {
	nc := share.NewNatsConn(cfg)
	info, err := share.ReadServerInfo()
	if err != nil {
		slog.Error("InitServer", "ReadServerInfo err", err)
		os.Exit(1)
	}
	subscriptionSubjects := []string{}
	for _, sub := range subjects {
		subscriptionSubjects = append(subscriptionSubjects, fmt.Sprintf("%s-%s", info.ID, sub))
	}
	server = &Server{
		cfg,
		info,
		nc,
		make(chan *Error),
		subscriptionSubjects, // each server will subscribe to its own subject as we can connect many servers to same nats server
		nil,
	}
	commandHandler := NewCommandHandler(server)
	server.commandHandler = commandHandler

	go server.Start()

	select {}
}
func (s *Server) Start() {
	slog.Info("Server is running")
	go server.HandleErrors()
	go s.IniSubscriptions()
}
func (s *Server) IniSubscriptions() {
	slog.Info("Server IniSubscriptions")
	for _, subject := range s.subscriptionSubjects {
		sub, err := s.nc.SubscribeToSubject(subject)
		if err != nil {
			s.ErrChan <- NewServerError("SubscribeToSubject error: "+err.Error(), false, nil)
			continue
		}
		slog.Info("Subscribed to subject: " + subject)
		go s.HandleSubscription(sub)
	}
}
func (s *Server) HandleSubscription(sub *nats.Subscription) {
	for msg, err := range sub.Msgs() {
		if err != nil {
			s.ErrChan <- NewServerError(fmt.Sprintf("SubHandler %s subject handle error: %s", sub.Subject, err.Error()), true, msg)
			continue
		}
		go s.HandleMsg(msg)
	}
}

func (s *Server) HandleMsg(msg *nats.Msg) {
	resp, err := s.commandHandler.parseCommand(msg)
	if err != nil {
		s.ErrChan <- NewServerError(fmt.Sprintf("parse %s subject command error: %s", msg.Subject, err.Error()), true, msg)
		return
	}
	r, _ := json.Marshal(resp)
	msg.Respond(r)
}

func (s *Server) HandleErrors() {
	for err := range s.ErrChan { // Continuously listen for errors
		if err != nil {
			s.SaveErrorLog(err)
			if err.IsPublishable {
				err.NcMsg.Respond([]byte(err.Msg))
			}
		}
	}
}
