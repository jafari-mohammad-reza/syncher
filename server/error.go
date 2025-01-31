package server

import (
	"encoding/json"
	"log/slog"
	"os"
	"path"
	"time"
)

type Error struct {
	Msg           string // can be interface{} as well
	IsPublishable bool
	ClientId      *string // determine that we should return error to which client if its Publishable and client id not null
}
type ErrorLog struct {
	*Error
	Time time.Time
}

func NewServerError(msg string, IsPublishable bool) *Error {
	return &Error{
		Msg:           msg,
		IsPublishable: IsPublishable,
	}
}
func (s *Server) SaveErrorLog(err *Error) {
	slog.Error("Server error", "err", err.Msg)
	homeDir, _ := os.UserHomeDir()
	ph := path.Join(homeDir, ".syncher", "server.log")

	log, _ := json.Marshal(ErrorLog{
		Error: err,
		Time:  time.Now(),
	})
	writeErr := os.WriteFile(ph, log, 0755)
	if writeErr != nil {
		panic(err)
	}
}
