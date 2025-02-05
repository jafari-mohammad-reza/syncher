package share

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
)

type ServerInfo struct {
	ID uuid.UUID `json:"id"`
	IP string    `json:"ip"`
}

type ClientInfo struct {
	ID        uuid.UUID   `json:"id"`
	Server    *ServerInfo `json:"server"`
	SharePath []string    `json:"share_paths"`
}
type ChangeEvent string

const (
	Modify ChangeEvent = "modify"
	Create ChangeEvent = "create"
	Delete ChangeEvent = "delete"
)

type ChangeLog struct {
	FileName string
	ClientId uuid.UUID
	Date     time.Time
	Event    ChangeEvent
}

func InitServerSyncherDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	ph := path.Join(homeDir, ".syncher")

	if err := os.MkdirAll(ph, 0755); err != nil {
		return err
	}

	serverFile := path.Join(ph, "server.json")
	if _, err := os.Stat(serverFile); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	ip, err := GetIPv4()
	if err != nil {
		return err
	}
	serverInfo := ServerInfo{
		ID: uuid.New(),
		IP: ip,
	}

	serverData, err := json.MarshalIndent(serverInfo, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(serverFile, serverData, 0644); err != nil {
		return err
	}

	changeLogFile := path.Join(ph, "server-changes.json")
	if _, err := os.Stat(changeLogFile); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(changeLogFile, []byte("{}"), 0644); err != nil {
			return err
		}
	}

	UploadDir := path.Join(ph, "uploads")
	if _, err := os.Stat(UploadDir); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(UploadDir, 0644); err != nil {
			return err
		}
	}

	return nil
}

func InitClientSyncherDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	ph := path.Join(homeDir, ".syncher")

	if err := os.MkdirAll(ph, 0755); err != nil {
		return err
	}

	clientFile := path.Join(ph, "client.json")
	if _, err := os.Stat(clientFile); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	clientInfo := ClientInfo{
		ID:        uuid.New(),
		Server:    nil,
		SharePath: []string{ph},
	}

	clientData, err := json.MarshalIndent(clientInfo, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(clientFile, clientData, 0644); err != nil {
		return err
	}

	changeLogFile := path.Join(ph, "changes.json")
	if _, err := os.Stat(changeLogFile); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(changeLogFile, []byte("{}"), 0644); err != nil {
			return err
		}
	}

	return nil
}

func ReadServerInfo() (*ServerInfo, error) {
	if err := InitServerSyncherDir(); err != nil {
		return nil, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	serverFile := path.Join(homeDir, ".syncher", "server.json")

	data, err := os.ReadFile(serverFile)
	if err != nil {
		return nil, err
	}

	var info ServerInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func ReadClientInfo() (*ClientInfo, error) {
	if err := InitClientSyncherDir(); err != nil {
		return nil, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	clientFile := path.Join(homeDir, ".syncher", "client.json")

	data, err := os.ReadFile(clientFile)
	if err != nil {
		return nil, err
	}

	var info ClientInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

type ClientCommand struct {
	ID       uuid.UUID         `json:"id"`
	ClientId uuid.UUID         `json:"client_id"`
	Args     map[string][]byte `json:"args"`
}

func NewClientCommand(clientId uuid.UUID, args map[string][]byte) *ClientCommand {
	return &ClientCommand{
		ID:       uuid.New(),
		ClientId: clientId,
		Args:     args,
	}
}

func ParseClientCommand(command []byte) (*ClientCommand, error) {
	var cmd ClientCommand
	err := json.Unmarshal(command, &cmd)
	if err != nil {
		slog.Error("ParseClientCommand", "err", err.Error())
		return nil, err
	}
	return &cmd, nil
}

type ServerReply struct {
	Msg      string
	ClientId uuid.UUID
}
