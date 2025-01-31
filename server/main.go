package server

import (
	"errors"
	"fmt"
	"log/slog"
	"sync_server/share"
)

type Server struct {
	cfg     *share.Config
	nc      *share.NatsConn
	ErrChan chan error
}

var server *Server

func InitServer(cfg *share.Config) {
	nc := share.NewNatsConn(cfg)
	share.InitServerSyncherDir()

	server = &Server{
		cfg,
		nc,
		make(chan error),
	}
	go server.Start()
	fmt.Println("Server is running")

	select {
	// TODO: run a error handler for server later that saves errors in log file
	case err := <-server.ErrChan:
		if err != nil {
			slog.Error("Server error", "err", err.Error())
		}
	}
}
func (s *Server) Start() {
	info, err := share.ReadServerInfo()
	if err != nil {
		s.ErrChan <- errors.New("ReadServerInfo error:" + err.Error())
	}
	fmt.Println(info)
}
