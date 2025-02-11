package client

import (
	"log/slog"
	"net/http"
	"sync_server/share"
)

type HttpListener struct {
	Cfg *share.ClientConfig
}

func NewHttpListener(cfg *share.ClientConfig) *HttpListener {
	return &HttpListener{Cfg: cfg}
}

func (h *HttpListener) Listen() error {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`client`))
	})
	slog.Info("Client http listening on", "port", h.Cfg.HttpPort)
	if err := http.ListenAndServe(":"+h.Cfg.HttpPort, nil); err != nil {
		slog.Error("Client http listening", "Err", err)
	}
	return nil
}
