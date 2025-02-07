package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
)

var ActiveTransfers = struct {
	sync.Mutex
	Transfers map[int]string
}{Transfers: make(map[int]string)}

func InitReceiver(port int, filePath string) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	ActiveTransfers.Lock()
	ActiveTransfers.Transfers[port] = filePath
	ActiveTransfers.Unlock()

	go func() {
		defer ln.Close()

		slog.Info("Receiver started", "port", port, "path", filePath)

		for {
			conn, err := ln.Accept()
			if err != nil {
				slog.Error("Failed to accept connection", "err", err)
				break
			}
			go handleConnection(conn, filePath, port)
		}

		// Remove completed transfer
		ActiveTransfers.Lock()
		delete(ActiveTransfers.Transfers, port)
		ActiveTransfers.Unlock()
	}()

	return nil
}

func handleConnection(conn net.Conn, filePath string, port int) {
	defer conn.Close()
	if err := handleUpload(conn, filePath); err != nil {
		slog.Error("Upload failed", "err", err)
	}
	slog.Info("Transfer completed", "port", port, "path", filePath)
}

func handleUpload(conn net.Conn, fileName string) error {
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
	slog.Info("writing file", "size", size, fileName, buf.Bytes())
	err = os.WriteFile(fileName, buf.Bytes(), 0655)
	if err != nil {
		slog.Error("Failed to write file", "err", err)
		return err
	}
	slog.Info("File saved successfully", "path", fileName)
	return nil
}
