package main

import (
	"log/slog"
	"sync_server/client"
	"sync_server/server"
	"sync_server/share"

	"github.com/labstack/gommon/log"
)

func main() {
	cfg, err := share.InitConfig(".env")
	if err != nil {
		slog.Error("Local config", "err", err.Error())
	}
	switch cfg.AppType {
	case "SERVER":
		server.InitServer(cfg)
	case "CLIENT":
		client.InitClient(cfg)
	default:
		log.Error("Invalid app type")
	}
}
