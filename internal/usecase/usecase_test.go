package usecase_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go-shortener-url/internal/pkg/deleteurl"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-shortener-url/internal/pkg/shortener"
	"go-shortener-url/internal/storage"
	"go-shortener-url/internal/usecase"
)

func TestExecDeleting(t *testing.T) {
	type (
		basic struct {
			userID   string
			shortURL string
			fullURL  string
		}

		want struct {
			err   error
			value string
		}

		test struct {
			want   want
			name   string
			userID string
			items  []string
		}
	)

	store := storage.NewMemStorage()
	baseURL := "http://localhost:8080"

	deleter := deleteurl.InitUrlDeleteService(store)
	deleter.Run(1)
	defer deleter.Stop()

	manager := usecase.New(store, deleter, baseURL)

	basics := []basic{
		{
			userID:   "1",
			shortURL: fmt.Sprintf("%s/%s", baseURL, "123"),
			fullURL:  "maps.yandex.ru/msk",
		},
	}

	tests := []test{
		{
			name:   "positive test",
			items:  []string{"123"},
			userID: "1",
			want:   want{value: "", err: usecase.ErrDeletedURL},
		},
	}

	for _, b := range basics {
		err := store.Add(context.Background(), b.userID, b.shortURL, b.fullURL)
		require.NoError(t, err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.ExecDeleting(tt.items, tt.userID)

			time.Sleep(1 * time.Millisecond)

			for _, i := range tt.items {
				shortURL := fmt.Sprintf("%s/%s", baseURL, i)
				v, err := store.Get(context.Background(), shortURL)
				assert.Equal(t, tt.want.value, v)
				assert.Equal(t, tt.want.err, err)
			}
		})
	}
}

func BenchmarkExecDeleting(b *testing.B) {
	type test struct {
		userID string
		items  []string
	}

	store := storage.NewMemStorage()
	baseURL := "http://localhost:8080"

	deleter := deleteurl.InitUrlDeleteService(store)
	deleter.Run(1)
	defer deleter.Stop()

	manager := usecase.New(store, deleter, baseURL)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		userID := generateSecureToken(7)
		fullURL := fmt.Sprintf("maps.yandex.ru/%s", generateSecureToken(10))
		shortURL, err := shortener.ShortenURL(fullURL)
		if err != nil {
			continue
		}
		_ = store.Add(context.Background(), userID, fmt.Sprintf("%s/%s", baseURL, shortURL), fullURL)

		tests := []test{
			{
				items:  []string{shortURL},
				userID: userID,
			},
		}

		b.StartTimer()
		for _, v := range tests {
			manager.ExecDeleting(v.items, v.userID)
		}
	}
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
