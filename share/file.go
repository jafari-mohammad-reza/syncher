package share

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func GetFilesByte(files []string) (map[string][]byte, error) {
	result := make(map[string][]byte)

	for _, file := range files {
		data, err := getFileBytes(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}
		result[file] = data
	}

	return result, nil
}

func getFileBytes(file string) ([]byte, error) {
	info, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s does not exist", file)
		}
		return nil, err
	}

	if info.IsDir() {

		files, err := os.ReadDir(file)
		if err != nil {
			return nil, fmt.Errorf("error reading directory %s: %w", file, err)
		}

		var allData []byte
		for _, f := range files {
			fullPath := file + "/" + f.Name()
			data, err := getFileBytes(fullPath)
			if err != nil {
				return nil, fmt.Errorf("error reading file %s: %w", fullPath, err)
			}
			allData = append(allData, data...)
		}
		return allData, nil
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", file, err)
	}
	return data, nil
}

func CreateArchive(files map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzWriter)
	for filename, content := range files {
		hdr := &tar.Header{
			Name: filepath.Base(filename),
			Size: int64(len(content)),
			Mode: 0644,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write(content); err != nil {
			return nil, err
		}
	}
	tw.Close()
	gzWriter.Close()
	return buf.Bytes(), nil
}

func CheckFileSize(file string, fileMb int64) error {
	stat, err := os.Stat(file)
	if err != nil {
		return err
	}
	size := stat.Size()
	if size > 1024*1024*fileMb {
		return fmt.Errorf("file %s is too large, it is %d bytes", file, size)
	}
	return nil
}

func UploadFile(fileName string, file []byte) error {
	homeDir, _ := os.UserHomeDir()
	ph := path.Join(homeDir, ".syncher", "uploads", fileName)
	return os.WriteFile(ph, file, 0644)
}
