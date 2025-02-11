package client

import (
	"fmt"
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
	fmt.Println("Listening on port:", h.Cfg.HttpPort)
	if err := http.ListenAndServe(":"+h.Cfg.HttpPort, nil); err != nil {
		fmt.Println("Server error:", err)
	}
	return nil
}
