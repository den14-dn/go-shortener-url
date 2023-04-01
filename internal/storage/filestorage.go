package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type FileStorage struct {
	file *os.File
	rw   *bufio.ReadWriter
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	flag := os.O_RDWR | os.O_CREATE | os.O_APPEND | os.O_SYNC
	f, err := os.OpenFile(filePath, flag, 0777)
	if err != nil {
		return nil, err
	}
	return &FileStorage{
		file: f,
		rw:   bufio.NewReadWriter(bufio.NewReader(f), bufio.NewWriter(f)),
	}, nil
}

func (f *FileStorage) Add(id, value string) error {
	data := fmt.Sprintf("%s=%s\n", id, value)
	if _, err := f.rw.Write([]byte(data)); err != nil {
		return err
	}
	return f.rw.Writer.Flush()
}

func (f *FileStorage) Get(id string) (string, error) {
	for {
		data, err := f.rw.ReadBytes('\n')
		if err == io.EOF {
			return "", errors.New("URL not found")
		} else if err != nil {
			return "", err
		}

		arr := strings.Split(string(data), "=")
		if arr[0] == id {
			return arr[1], nil
		}
	}
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
