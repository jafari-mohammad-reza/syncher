package main

import (
	"github.com/google/uuid"
	"sync_server/server"
	"sync_server/share"
)

func main() {
	id, _ := uuid.NewUUID()
	cfg, err := share.GetServerConfig()
	if err != nil {
		panic(err)
	}
	cfg.ServerId = id.String()
	server.NewServer(cfg).Start()
}
