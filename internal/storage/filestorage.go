package storage

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// FileStorage manages the storage of data in a file on disk.
type FileStorage struct {
	file       *os.File
	writer     *bufio.Writer
	memStorage *MemStorage
}

// NewFileStorage is a constructor for the FileStorage structure.
func NewFileStorage(ctx context.Context, filePath string) *FileStorage {
	flag := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	file, _ := os.OpenFile(filePath, flag, 0777)

	return &FileStorage{
		file:       file,
		writer:     bufio.NewWriter(file),
		memStorage: createMemStorage(ctx, filePath),
	}
}

// Add writes the original and its shortened URL by user id.
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

// Get retrieves the original URL by its shortened value. In-memory storage is used for acceleration.
func (f *FileStorage) Get(ctx context.Context, shortURL string) (string, error) {
	return f.memStorage.Get(ctx, shortURL)
}

// GetByUser gets a map of URLs by user ID. In-memory storage is used for acceleration.
func (f *FileStorage) GetByUser(ctx context.Context, userID string) (map[string]string, error) {
	return f.memStorage.GetByUser(ctx, userID)
}

// CheckStorage checks for the presence of a file.
func (f *FileStorage) CheckStorage(_ context.Context) error {
	_, err := f.file.Stat()
	return err
}

// Delete marks the URL as deleted in the file.
func (f *FileStorage) Delete(ctx context.Context, shortURL string) error {
	origURL, err := f.memStorage.Get(ctx, shortURL)
	if err != nil {
		return err
	}

	f.memStorage.Delete(ctx, shortURL)

	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return errors.New("UserID not found")
	}

	data := fmt.Sprintf("%s=%s=%s=%s\n", userID, shortURL, origURL, "true")
	if _, err := f.writer.Write([]byte(data)); err != nil {
		return err
	}

	return f.writer.Flush()
}

// Close closes the file after writing, reading.
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
			if len(arr) > 3 && arr[3] == "true" {
				storage.Delete(ctx, arr[1])
			}
		}
	}

	return storage
}
