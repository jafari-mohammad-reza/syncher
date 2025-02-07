package share

import (
	"github.com/fsnotify/fsnotify"
	"time"
)

type ResponseStatus string

const (
	Success ResponseStatus = "success"
	Failure ResponseStatus = "failure"
)

type ServerResponse struct {
	Status ResponseStatus
	Data   []byte
}

type ClientRequest struct {
	ClientId string
	Time     time.Time
}

type ChangeRequestChanges struct {
	FileName    string
	ChangeEvent fsnotify.Op
}
type ChangeRequest struct {
	ClientRequest
	Dir     string
	Changes []ChangeRequestChanges
}
type ChangeResponse map[string]int
