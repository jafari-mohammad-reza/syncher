package main

import (
	"sync_server/client"
	"sync_server/share"

	"github.com/google/uuid"
)

func main() {
	id, _ := uuid.NewUUID()
	share.InitClientConfig()
	cfg, err := share.GetClientConfig()
	if err != nil {
		panic(err)
	}
	cfg.ClientId = id.String()
	client.NewClient(cfg).Start()

}
