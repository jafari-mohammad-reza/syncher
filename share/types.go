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
	Agent    string
}

type ChangeRequestChange struct {
	FileName    string
	ChangeEvent string
	Agent       string
}
type ChangeRequest struct {
	ClientRequest
	Dir     string
	Changes []ChangeRequestChange
}
type ChangeResponse map[string]int

type SyncResponse struct {
	Dir     string
	Changes []ChangeRequestChange
}

type DownloadRequest struct {
	ClientRequest
	FilePath string
}

type DownloadResponse struct {
	Port int
}
