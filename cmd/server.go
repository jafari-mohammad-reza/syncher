package main

import (
	"sync_server/server"
	"sync_server/share"
)

func main() {
	cfg, err := share.GetServerConfig()
	if err != nil {
		panic(err)
	}
	server.NewServer(cfg).Start()
}
