package server

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

type LogEvent string

const (
	Err  LogEvent = "error"
	Warn LogEvent = "warn"
	Info LogEvent = "info"
)

type Log struct {
	Message string
	Event   LogEvent
	Time    time.Time
}

func (s *Server) log(sbj, msg string) {
	slog.Info("Server Info", sbj, msg)
	s.recordLog(Log{
		Message: msg,
		Event:   Info,
		Time:    time.Now(),
	})
}
func (s *Server) error(err Error) {
	slog.Error("Server Error", "Err", err.ErrorMsg)
	s.recordLog(Log{
		Message: err.ErrorMsg,
		Event:   Err,
		Time:    time.Now(),
	})
}

func (s *Server) recordLog(log Log) error {
	var logPath string
	switch log.Event {
	case Err:
		logPath = "logs/error.log"
	case Warn, Info:
		logPath = "logs/server.log"
	default:
		return fmt.Errorf("unknown log event type")
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	logMessage := fmt.Sprintf("[%s] %s: %s\n", log.Time.Format(time.RFC3339), log.Event, log.Message)
	if _, err := file.WriteString(logMessage); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

type ChangeLogChanges struct {
	FileName string
	Change   string
}
type ChangeLog struct {
	ClientId  string
	ServerId  string
	ChangeDir string
	Changes   []ChangeLogChanges
	Time      time.Time
}

func recordChange(log ChangeLog) error {
	logPath := "logs/change.log"

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	logMessage := fmt.Sprintf("[%s] clientId: %s - dir: %s - files: %+v - serverId: %s\n", log.Time.Format(time.RFC3339), log.ClientId, log.ChangeDir, log.Changes, log.ServerId)
	if _, err := file.WriteString(logMessage); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}
