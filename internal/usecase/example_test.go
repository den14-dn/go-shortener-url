package usecase_test

import (
	"context"
	"fmt"
	"go-shortener-url/internal/pkg/deleteurl"
	"go-shortener-url/internal/usecase"

	"go-shortener-url/internal/pkg/shortener"
	"go-shortener-url/internal/storage"
)

func ExampleManager_ExecDeleting() {
	userID := "Aa135798642"
	items := []string{"star.example.ru/questions/20467179"}
	fullURL := items[0]

	store := storage.NewMemStorage()
	baseURL := "http://localhost:8080"

	deleter := deleteurl.InitUrlDeleteService(store)
	go deleter.Run(1)
	defer deleter.Stop()

	manager := usecase.New(store, deleter, baseURL)

	id, err := shortener.ShortenURL(fullURL)
	if err != nil {
		return
	}
	shortURL := fmt.Sprintf("%s/%s", baseURL, id)

	err = store.Add(context.Background(), userID, shortURL, fullURL)
	if err != nil {
		return
	}

	manager.ExecDeleting(items, userID)
}
