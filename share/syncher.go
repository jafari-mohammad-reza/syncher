package share

import (
	"github.com/google/uuid"
)

type ServerInfo struct {
	ID uuid.UUID `json:"id"`
	IP string    `json:"ip"`
}

type ChangeEvent string

const (
	Modify ChangeEvent = "modify"
	Create ChangeEvent = "create"
	Delete ChangeEvent = "delete"
)

type ServerReply struct {
	Msg      string
	ClientId uuid.UUID
}
