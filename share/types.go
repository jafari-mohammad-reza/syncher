package share

import (
	"time"
)

type ResponseStatus string

const (
	Success ResponseStatus = "success"
	Failure ResponseStatus = "failure"
)

type ServerResponse struct {
	Status ResponseStatus
	Data   string `json:"data"`
}

type ClientRequest struct {
	ClientId string
	Time     time.Time
}

type ChangeRequestChanges struct {
	FileName    string
	ChangeEvent string
}
type ChangeRequest struct {
	ClientRequest
	Dir     string
	Changes []ChangeRequestChanges
}
type ChangeResponse map[string]int
type SyncResponseItem struct {
	Dir     string
	Changes []ChangeRequestChanges
}
type SyncResponse struct {
	Changes []SyncResponseItem
}
