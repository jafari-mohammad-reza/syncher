package server

import (
	"context"
	"io"
	"log/slog"
	"sync_server/share"

	"github.com/minio/minio-go"
)

type FileStorage interface {
	Init() error
	Upload(ctx context.Context, fileName string, reader io.Reader, size int64) error
	UploadPath(ctx context.Context, fileName string, filePath string) error
	RemoveFile(fileName string) error
	Download(ctx context.Context, fileName string) (io.ReadCloser, error)
}

type MiniOStorage struct {
	Cfg    *share.ServerConfig
	client *minio.Client
}

func NewMinIoService(cfg *share.ServerConfig) *MiniOStorage {
	server := &MiniOStorage{
		Cfg: cfg,
	}
	server.Init()
	return server
}

func (m *MiniOStorage) Init() error {
	minioClient, err := minio.New(m.Cfg.Endpoint, m.Cfg.MinIO.AccessKeyID, m.Cfg.MinIO.SecretAccessKey, m.Cfg.MinIO.UseSSL)
	if err != nil {
		return err
	}
	m.client = minioClient
	return nil
}

func (m *MiniOStorage) Upload(ctx context.Context, fileName string, reader io.Reader, size int64) error {
	slog.Info("Uploading file", "filename", fileName)
	_, err := m.client.PutObjectWithContext(ctx, "syncher", fileName, reader, size, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *MiniOStorage) UploadPath(ctx context.Context, fileName string, filePath string) error {
	slog.Info("Uploading file", "filename", fileName, "filePath", filePath)
	_, err := m.client.FPutObjectWithContext(ctx, "syncher", fileName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}
func (m *MiniOStorage) RemoveFile(fileName string) error {
	slog.Info("Removing file", "filename", fileName)
	err := m.client.RemoveObject("syncher", fileName)
	if err != nil {
		return err
	}
	return nil
}

func (m *MiniOStorage) Download(ctx context.Context, fileName string) (io.ReadCloser, error) {
	return m.client.GetObjectWithContext(ctx, "syncher", fileName, minio.GetObjectOptions{})
}
