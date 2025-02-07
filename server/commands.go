package server

import (
	"github.com/nats-io/nats.go"
)

type ResponseStatus string

const (
	Success ResponseStatus = "success"
	Failure ResponseStatus = "failure"
)

type Response struct {
	Status ResponseStatus
	Data   []byte
}

func Health(msg *nats.Msg) (*Response, error) {
	return &Response{
		Status: Success,
		Data:   []byte("healthy"),
	}, nil
}
