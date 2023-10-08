package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-shortener-url/internal/storage"
)

func TestExecDeleting(t *testing.T) {
	type (
		basic struct {
			userID   string
			shortURL string
			fullURL  string
		}

		want struct {
			value string
			err   error
		}

		test struct {
			name   string
			items  []string
			userID string
			want   want
		}
	)

	manager := &Manager{
		store:   storage.NewMemStorage(),
		baseURL: "http://localhost:8080",
	}

	basics := []basic{
		{
			userID:   "1",
			shortURL: fmt.Sprintf("%s/%s", manager.baseURL, "123"),
			fullURL:  "maps.yandex.ru/msk",
		},
	}

	tests := []test{
		{
			name:   "positive test",
			items:  []string{"123"},
			userID: "1",
			want:   want{value: "", err: ErrDeletedURL},
		},
	}

	for _, b := range basics {
		err := manager.store.Add(context.Background(), b.userID, b.shortURL, b.fullURL)
		require.NoError(t, err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.ExecDeleting(tt.items, tt.userID)

			for _, i := range tt.items {
				shortURL := fmt.Sprintf("%s/%s", manager.baseURL, i)
				v, err := manager.store.Get(context.Background(), shortURL)
				assert.Equal(t, tt.want.value, v)
				assert.Equal(t, tt.want.err, err)
			}
		})
	}
}

func BenchmarkExecDeleting(b *testing.B) {
	type test struct {
		items  []string
		userID string
	}

	manager := &Manager{
		store:   storage.NewMemStorage(),
		baseURL: "http://localhost:8080",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		userID := generateSecureToken(7)
		fullURL := fmt.Sprintf("maps.yandex.ru/%s", generateSecureToken(10))
		shortURL, err := shortenURL(fullURL)
		if err != nil {
			continue
		}
		_ = manager.store.Add(context.Background(), userID, fmt.Sprintf("%s/%s", manager.baseURL, shortURL), fullURL)

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
