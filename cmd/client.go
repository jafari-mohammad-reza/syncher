package main

import (
	"github.com/google/uuid"
	"sync_server/client"
	"sync_server/share"
)

func main() {
	id, _ := uuid.NewUUID()
	cfg, err := share.GetClientConfig()
	if err != nil {
		panic(err)
	}
	cfg.ClientId = id.String()
	client.NewClient(cfg).Start()

}
