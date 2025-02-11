package client

import (
	"encoding/json"
	"fmt"
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
	fmt.Println("Syncing changes, items in channel:", len(s.ChangeChan))

	dirMap := make(map[string][]share.ChangeRequestChanges)
	reqs := []share.ChangeRequest{}

	// Collect changes per directory first
	for {
		select {
		case change := <-s.ChangeChan:
			fmt.Println("Received change:", change)
			dirMap[change.Dir] = append(dirMap[change.Dir], change.File)
		default:
			// Once the channel is empty, process the directories
			if len(dirMap) > 0 {
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
						fmt.Println("Error marshaling change request:", err)
						continue
					}

					msg, err := s.NatsConn.RequestToSubject("change", reqJson, time.Second*3)
					if err != nil {
						fmt.Println("Error sending message to NATS:", err)
						continue
					}
					var serverResp share.ServerResponse
					if err := json.Unmarshal(msg.Data, &serverResp); err != nil {
						fmt.Println("Error unmarshaling server response:", err)
						continue
					}

					if serverResp.Status != "success" {
						fmt.Println("failed from server:", serverResp.Data)
						continue
					}
					fmt.Println("Change request sent successfully:", string(msg.Data), serverResp.Data)
					var changeRes share.ChangeResponse
					err = json.Unmarshal([]byte(serverResp.Data), &changeRes)
					if err != nil {
						fmt.Println("Error unmarshaling change response:", err)
						continue
					}
					fmt.Println("changeResponses", changeRes)
					for fileName, port := range changeRes {
						fmt.Println("change-res", fileName, port)
						for _, ch := range req.Changes {
							if fileName == ch.FileName {
								// TODO: move them to upload channel
								fmt.Println("change file upload port", ch.FileName, port)
							}
						}
					}
				}

				// Reset reqs after sending
				reqs = nil
				dirMap = make(map[string][]share.ChangeRequestChanges)

			}
			return
		}
	}
}
