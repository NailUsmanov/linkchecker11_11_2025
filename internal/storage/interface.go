package storage

import (
	"context"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/models"
)

// Storage сохраняет ссылки, переданные пользователем и выдает их при запросе
type Storage interface {
	Save(ctx context.Context, links models.ResponseSentLinks) error
	Get(ctx context.Context, num []int) (map[int]models.ResponseSentLinks, error)
}
