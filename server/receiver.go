package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync_server/share"
)

type ReceiverService struct {
	Cfg             *share.ServerConfig
	ActiveTransfers struct {
		sync.Mutex
		Transfers map[int]string
	}
	MinIOService *share.MinIOService
}

func NewReceiverService(Cfg *share.ServerConfig) *ReceiverService {
	var ActiveTransfers = struct {
		sync.Mutex
		Transfers map[int]string
	}{Transfers: make(map[int]string)}
	MinIOService := share.NewMinIoService(Cfg)
	return &ReceiverService{
		Cfg,
		ActiveTransfers,
		MinIOService,
	}
}

func (r *ReceiverService) InitReceiver(port int, filePath string) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	r.ActiveTransfers.Lock()
	r.ActiveTransfers.Transfers[port] = filePath
	r.ActiveTransfers.Unlock()

	go func() {
		defer ln.Close()

		slog.Info("Receiver started", "port", port, "path", filePath)

		for {
			conn, err := ln.Accept()
			if err != nil {
				slog.Error("Failed to accept connection", "err", err)
				break
			}
			go r.handleConnection(conn, filePath, port)
		}

		// Remove completed transfer
		r.ActiveTransfers.Lock()
		delete(r.ActiveTransfers.Transfers, port)
		r.ActiveTransfers.Unlock()
	}()

	return nil
}

func (r *ReceiverService) handleConnection(conn net.Conn, filePath string, port int) {
	defer conn.Close()
	if err := r.handleUpload(conn, filePath); err != nil {
		slog.Error("Upload failed", "err", err)
	}
	slog.Info("Transfer completed", "port", port, "path", filePath)
}

func (r *ReceiverService) handleUpload(conn net.Conn, fileName string) error {
	buf := new(bytes.Buffer)
	var size int64
	err := binary.Read(conn, binary.BigEndian, &size)
	if err != nil {
		slog.Error("Failed to read file size", "err", err)
		return err
	}
	slog.Info("Size received", "size", size)
	_, err = io.CopyN(buf, conn, size)
	if err != nil {
		slog.Error("File reception error", "err", err)
		return err
	}
	err = r.MinIOService.Upload(context.Background(), fileName, buf, size)
	if err != nil {
		slog.Error("Failed to save file", "err", err)
		return err
	}
	slog.Info("File saved successfully", "path", fileName)
	return nil
}
