package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type Storage interface {
	Get(key string) (*[]byte, error)
	Del(key string) error
	Set(key string, value []byte) error
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

func (storage *ChangeStorage) Get(key string) (*[]byte, error) {
	return nil, nil
}
func (storage *ChangeStorage) Del(key string) error {
	return nil
}
func (storage *ChangeStorage) Set(key string, value []byte) error {
	return nil
}

func LoadLogs() (map[string][]ChangeLog, error) {
	file, err := os.Open("logs/changes.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	logs := make(map[string][]ChangeLog)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var log ChangeLog
		err := json.Unmarshal(scanner.Bytes(), &log)
		if err != nil {
			continue
		}
		logs[log.ClientId] = append(logs[log.ClientId], log)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return logs, nil
}
