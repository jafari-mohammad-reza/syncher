package share

import (
	"context"
	"io"
	"log"
	"log/slog"

	"github.com/minio/minio-go"
)

type MinIOService struct {
	Cfg    *ServerConfig
	client *minio.Client
}

func NewMinIoService(cfg *ServerConfig) *MinIOService {
	server := &MinIOService{
		Cfg: cfg,
	}
	server.init()
	return server
}

func (m *MinIOService) init() {
	minioClient, err := minio.New(m.Cfg.Endpoint, m.Cfg.MinIO.AccessKeyID, m.Cfg.MinIO.SecretAccessKey, m.Cfg.MinIO.UseSSL)
	if err != nil {
		log.Fatalln(err)
	}
	m.client = minioClient
}

func (m *MinIOService) Upload(ctx context.Context, fileName string, reader io.Reader, size int64) error {
	slog.Info("Uploading file", "filename", fileName)
	_, err := m.client.PutObjectWithContext(ctx, "syncher", fileName, reader, size, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *MinIOService) UploadPath(ctx context.Context, fileName string, filePath string) error {
	slog.Info("Uploading file", "filename", fileName, "filePath", filePath)
	_, err := m.client.FPutObjectWithContext(ctx, "syncher", fileName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}
