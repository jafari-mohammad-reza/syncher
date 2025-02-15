package main

import (
	"sync_server/client"
	"sync_server/share"

	"github.com/google/uuid"
)

func main() {
	share.InitClientConfig()
	cfg, err := share.GetClientConfig()
	if cfg.ClientId == "" {
		id, _ := uuid.NewUUID()
		cfg.ClientId = id.String()
		share.WriteClientConfig()
	}
	if err != nil {
		panic(err)
	}

	client.NewClient(cfg).Start()

}
