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
	"sync_server/share"
	"time"
)

type ChangeEvent struct {
	Dir  string
	File share.ChangeRequestChanges
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
func (s *SyncService) syncChanges() {
	slog.Info("Syncing changes", "items in channel", len(s.ChangeChan))

	dirMap := make(map[string][]share.ChangeRequestChanges)
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

				// Process change responses
				for fileName, port := range changeRes {
					fmt.Println("change-res", fileName, port)
					for _, ch := range req.Changes {
						if fileName == ch.FileName {
							fmt.Println("change file upload port", ch.FileName, port)

							// Directly fetch the parent directory from dirMap
							if parentDir, exists := dirMap[ch.FileName]; exists {
								go s.uploadFile(fmt.Sprintf("%s/%s", parentDir, fileName), port)
							}
						}
					}
				}
			}

			// Clear the map and slice
			reqs = reqs[:0]
			dirMap = make(map[string][]share.ChangeRequestChanges)
			return
		}
	}
}

func (s *SyncService) uploadFile(filePath string, port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("error dialing to %d: %s", port, err.Error())
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
	}
}
