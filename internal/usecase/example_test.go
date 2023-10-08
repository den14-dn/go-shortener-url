package usecase_test

import (
	"context"
	"fmt"

	"go-shortener-url/internal/pkg/shortener"
	"go-shortener-url/internal/storage"
	"go-shortener-url/internal/usecase"
)

func ExampleManager_ExecDeleting() {
	userID := "Aa135798642"
	items := []string{"star.example.ru/questions/20467179"}
	fullURL := items[0]

	store := storage.NewMemStorage()
	baseURL := "http://localhost:8080"
	manager := usecase.New(store, baseURL)

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
