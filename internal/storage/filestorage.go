package storage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

type FileStorage struct {
	file       *os.File
	writer     *bufio.Writer
	memStorage *MemStorage
}

func NewFileStorage(ctx context.Context, filePath string) *FileStorage {
	flag := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	file, _ := os.OpenFile(filePath, flag, 0777)

	return &FileStorage{
		file:       file,
		writer:     bufio.NewWriter(file),
		memStorage: createMemStorage(ctx, filePath),
	}
}

func (f *FileStorage) Add(ctx context.Context, userID, shortURL, origURL string) error {
	_, errUser := f.memStorage.GetByUser(ctx, userID)
	_, errURLs := f.memStorage.Get(ctx, shortURL)
	if errUser != nil || errURLs != nil {
		err := f.memStorage.Add(ctx, userID, shortURL, origURL)
		if err != nil {
			return err
		}
		data := fmt.Sprintf("%s=%s=%s\n", userID, shortURL, origURL)
		if _, err := f.writer.Write([]byte(data)); err != nil {
			return err
		}
		return f.writer.Flush()
	}
	return nil
}

func (f *FileStorage) Get(ctx context.Context, shortURL string) (string, error) {
	return f.memStorage.Get(ctx, shortURL)
}

func (f *FileStorage) GetByUser(ctx context.Context, userID string) (map[string]string, error) {
	return f.memStorage.GetByUser(ctx, userID)
}

func (f *FileStorage) CheckStorage(ctx context.Context) error {
	_, err := f.file.Stat()
	return err
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}

func createMemStorage(ctx context.Context, filePath string) *MemStorage {
	storage := NewMemStorage()

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			data := scanner.Text()
			arr := strings.Split(data, "=")
			storage.Add(ctx, arr[0], arr[1], arr[2])
		}
	}

	return storage
}
