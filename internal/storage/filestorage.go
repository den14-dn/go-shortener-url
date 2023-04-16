package storage

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type FileStorage struct {
	file       *os.File
	writer     *bufio.Writer
	memStorage *MemStorage
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	flag := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	file, err := os.OpenFile(filePath, flag, 0777)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		file:       file,
		writer:     bufio.NewWriter(file),
		memStorage: createMemStorage(filePath),
	}, nil
}

func (f *FileStorage) Add(idUser, shortURL, origURL string) error {
	v, _ := f.memStorage.Get(idUser, shortURL)
	if v == "" {
		err := f.memStorage.Add(idUser, shortURL, origURL)
		if err != nil {
			return err
		}
		data := fmt.Sprintf("%s=%s=%s\n", idUser, shortURL, origURL)
		if _, err := f.writer.Write([]byte(data)); err != nil {
			return err
		}
		return f.writer.Flush()
	}
	return nil
}

func (f *FileStorage) Get(idUser, shortURL string) (string, error) {
	return f.memStorage.Get(idUser, shortURL)
}

func (f *FileStorage) GetByUser(idUser string) (map[string]string, error) {
	return f.memStorage.GetByUser(idUser)
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}

func createMemStorage(filePath string) *MemStorage {
	storage := NewMemStorage()

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			data := scanner.Text()
			arr := strings.Split(data, "=")
			storage.Add(arr[0], arr[1], arr[2])
		}
	}

	return storage
}
