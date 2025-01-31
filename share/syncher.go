package share

import (
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
	"os"
	"path"
)

type ServerInfo struct {
	ID uuid.UUID `json:"id"`
	IP string    `json:"ip"`
}
type ClientInfo struct {
}
type ChangeLog struct {
}

func InitServerSyncherDir() error {
	homeDir, _ := os.UserHomeDir()
	ph := path.Join(homeDir, ".syncher")
	_, err := os.ReadDir(ph)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(ph, 0755)
			ip, err := GetIPv4()
			if err != nil {
				return err
			}
			serverInfos, _ := json.Marshal(ServerInfo{
				ID: uuid.New(),
				IP: ip,
			})
			os.WriteFile(path.Join(ph, "server.json"), []byte(serverInfos), 0755)
		} else {
			slog.Error("InitSyncherDir", "err", err.Error())
		}
	}
	return nil
}

func ReadServerInfo() (*ServerInfo, error) {
	var info ServerInfo
	homeDir, _ := os.UserHomeDir()
	ph := path.Join(homeDir, ".syncher", "server.json")
	f, err := os.ReadFile(ph)
	if err != nil {
		slog.Error("ReadServerInfo", "err", err.Error())
	}
	err = json.Unmarshal(f, &info)
	return &info, nil
}
