package server

import (
	"fmt"
	"sync_server/share"
)

type Server struct {
	cfg     *share.Config
	nc      *share.NatsConn
	ErrChan chan *Error
}

var server *Server

func InitServer(cfg *share.Config) {
	nc := share.NewNatsConn(cfg)
	share.InitServerSyncherDir()

	server = &Server{
		cfg,
		nc,
		make(chan *Error),
	}
	go server.Start()
	go server.HandleErrors()

	select {}
}
func (s *Server) Start() {
	fmt.Println("Server is running")
	_, err := share.ReadServerInfo()
	if err != nil {
		s.ErrChan <- NewServerError("ReadServerInfo error: "+err.Error(), false)
	}
	s.ErrChan <- NewServerError("Test error", false)
}
func (s *Server) HandleErrors() {
	for err := range s.ErrChan { // Continuously listen for errors
		if err != nil {
			s.SaveErrorLog(err)
		}
	}
}
