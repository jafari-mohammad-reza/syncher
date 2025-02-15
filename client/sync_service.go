package client

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"runtime"
	"sync_server/share"
	"time"
)

type ChangeEvent struct {
	Dir  string
	File share.ChangeRequestChange
	Time time.Time
}

type SyncService struct {
	Cfg        *share.ClientConfig
	NatsConn   *share.NatsConn
	ChangeChan chan ChangeEvent
	done       chan bool
}

func NewSyncService(cfg *share.ClientConfig) *SyncService {
	service := &SyncService{
		Cfg:        cfg,
		NatsConn:   share.NewNatsConn(cfg.NatsUrl),
		ChangeChan: make(chan ChangeEvent, 100),
		done:       make(chan bool),
	}
	go service.Listen()
	return service
}

func (s *SyncService) Listen() {
	ticker := time.NewTicker(time.Duration(s.Cfg.SyncInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go s.syncChanges()
			go s.retrieveChanges()
		case <-s.done:
			fmt.Println("Shutting down SyncService...")
			return
		default:

			if len(s.ChangeChan) > 10 {
				go s.syncChanges()
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
func (s *SyncService) retrieveChanges() {
	slog.Info("Retrieve changes")
	runningSystem := runtime.GOOS
	req, err := json.Marshal(share.ClientRequest{ClientId: s.Cfg.ClientId, Time: time.Now(), Agent: runningSystem})
	if err != nil {
		slog.Error("Retrieve changes parsing request", "err", err.Error())
		return
	}
	msg, err := s.NatsConn.RequestToSubject("sync", req, time.Second)
	if err != nil {
		slog.Error("Retrieve changes request", "err", err.Error())
		return
	}
	var serverResp share.ServerResponse
	if err := json.Unmarshal(msg.Data, &serverResp); err != nil {
		slog.Error("Error unmarshalling server response:", "err", err)
		return
	}

	if serverResp.Status != "success" {
		slog.Error("Failure response from server:", "Response", serverResp.Data)
		return
	}

	var res []share.SyncResponse
	json.Unmarshal([]byte(serverResp.Data), &res)

	// TODO: apply other device changes here
	for _, changeRes := range res {
		for _, change := range changeRes.Changes {
			if change.Agent != runningSystem {
				s.applyChange(changeRes.Dir, change)
			}
		}
	}
}
func (s *SyncService) applyChange(dir string, change share.ChangeRequestChange) {

	switch change.ChangeEvent {
	case "CREATE":
		// TODO: download file from storage
		req, _ := json.Marshal(share.DownloadRequest{ClientRequest: share.ClientRequest{ClientId: s.Cfg.ClientId, Time: time.Now(), Agent: runtime.GOOS}, FilePath: fmt.Sprintf("%s%s/%s", s.Cfg.ClientId, dir, change.FileName)})
		msg, err := s.NatsConn.RequestToSubject("download-file", req, time.Second)
		if err != nil {
			slog.Error("Error downloading file", "err", err)
			return
		}
		var res share.ServerResponse
		err = json.Unmarshal(msg.Data, &res)
		if err != nil {
			slog.Error("Error unmarshaling download response", "err", err)
			return
		}
		var downloadRes share.DownloadResponse
		err = json.Unmarshal([]byte(res.Data), &downloadRes)
		if err != nil {
			slog.Error("Error unmarshaling download response", "err", err)
			return
		}
		fileBytes, err := s.downloadFile(downloadRes.Port)
		if err != nil {
			slog.Error("Error downloading file", "err", err)
			return
		}
		os.WriteFile(fmt.Sprintf("%s/%s", dir, change.FileName), fileBytes, 0644)
	case "REMOVE":
		os.Remove(fmt.Sprintf("%s/%s", dir, change.FileName))
	}
	// todo: after applying the change we should record it and don't apply it later

}
func (s *SyncService) syncChanges() {
	slog.Info("Syncing changes", "items in channel", len(s.ChangeChan))

	dirMap := make(map[string][]share.ChangeRequestChange)
	reqs := make([]share.ChangeRequest, 0)

	for {
		select {
		case change := <-s.ChangeChan:
			dirMap[change.Dir] = append(dirMap[change.Dir], change.File)
		default:
			if len(dirMap) == 0 {
				return
			}

			// Convert dirMap to requests
			for dir, changedFiles := range dirMap {
				reqs = append(reqs, share.ChangeRequest{
					ClientRequest: share.ClientRequest{
						ClientId: s.Cfg.ClientId,
						Time:     time.Now(),
						Agent:    runtime.GOOS,
					},
					Dir:     dir,
					Changes: changedFiles,
				})
			}

			for _, req := range reqs {
				reqJson, err := json.Marshal(req)
				if err != nil {
					slog.Error("Error marshaling change request:", "err", err)
					continue
				}

				msg, err := s.NatsConn.RequestToSubject("change", reqJson, time.Second*3)
				if err != nil {
					slog.Error("Error sending message to NATS:", "err", err)
					continue
				}

				var serverResp share.ServerResponse
				if err := json.Unmarshal(msg.Data, &serverResp); err != nil {
					slog.Error("Error unmarshalling server response:", "err", err)
					continue
				}

				if serverResp.Status != "success" {
					slog.Error("Failure response from server:", "Response", serverResp.Data)
					continue
				}

				var changeRes share.ChangeResponse
				err = json.Unmarshal([]byte(serverResp.Data), &changeRes)
				if err != nil {
					slog.Error("Error unmarshaling change response:", "err", err)
					continue
				}

				for fileName, port := range changeRes {
					for parentDir, changes := range dirMap {
						for _, change := range changes {
							if change.FileName == fileName {
								go s.uploadFile(fmt.Sprintf("%s/%s", parentDir, fileName), port)
							}
						}
					}
				}
			}

			// Clear the map and slice
			reqs = reqs[:0]
			dirMap = make(map[string][]share.ChangeRequestChange)
			return
		}
	}
}

func (s *SyncService) uploadFile(filePath string, port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("error dialing to %d: %s", port, err.Error())
		return
	}
	fileSize, _ := share.GetSize(filePath)
	err = binary.Write(conn, binary.BigEndian, fileSize)
	if err != nil {
		slog.Error("error sending file to %d: %s", port, err.Error())
		return
	}
	fileByte, _ := os.ReadFile(filePath)
	_, err = io.CopyN(conn, bytes.NewReader(fileByte), fileSize)
	if err != nil {
		slog.Error("error sending file to %d: %s", port, err.Error())
		return
	}
}

func (s *SyncService) downloadFile(port int) ([]byte, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("error dialing to %d: %s", port, err.Error())
		return nil, err
	}

	buf := new(bytes.Buffer)
	var size int64
	err = binary.Read(conn, binary.BigEndian, &size)
	if err != nil {
		slog.Error("Failed to read file size", "err", err)
		return nil, err
	}
	slog.Info("Size received", "size", size)
	_, err = io.CopyN(buf, conn, size)
	if err != nil {
		slog.Error("File reception error", "err", err)
		return nil, err
	}
	return buf.Bytes(), nil
}
