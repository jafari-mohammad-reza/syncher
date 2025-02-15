package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type Storage interface {
	Get(key string) (interface{}, error)
	Del(key string) error
	Set(key string, value interface{}) error
}

type ChangeStorage struct {
	clientChanges map[string][]ChangeLog
}

func NewChangeStorage() *ChangeStorage {
	logs, err := LoadLogs()
	if err != nil {
		slog.Error("Change storage load fail", "err", err.Error())
		logs = nil
	}
	return &ChangeStorage{
		clientChanges: logs,
	}
}

func (storage *ChangeStorage) Get(key string) (interface{}, error) {
	data, ok := storage.clientChanges[key]
	if !ok {
		return nil, fmt.Errorf("%s key not found", key)
	}
	return data, nil
}
func (storage *ChangeStorage) Del(key string) error {
	return nil
}
func (storage *ChangeStorage) Set(key string, value interface{}) error {
	return nil
}

func LoadLogs() (map[string][]ChangeLog, error) {
	file, err := os.ReadFile("logs/changes.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	var loadedLog []ChangeLog
	err = json.Unmarshal(file, &loadedLog)
	logs := make(map[string][]ChangeLog)

	for _, log := range loadedLog {
		logs[log.ClientId] = append(logs[log.ClientId], log)
	}
	return logs, nil
}
