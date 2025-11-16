package storage

import (
	"context"
	"sync"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/models"
)

// MemoryStorage сущность для хранения ссылок и их статусов. Так как мапа непотокобезопасна, ставлю sync.RWmutex.
type MemoryStorage struct {
	mu   sync.RWMutex
	data map[int]models.ResponseSentLinks
}

// NewMemoryStorage - создает новую структуру MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[int]models.ResponseSentLinks),
	}
}

// Save - сохраняет в базу данных ссылки пользователя с проверенным статусом доступности.
func (m *MemoryStorage) Save(ctx context.Context, resp models.ResponseSentLinks) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[resp.Num] = resp
	return nil
}

// Get - достает данные в пдф файл и выдает пользователю провернные ссылки по номерам запроса.
func (m *MemoryStorage) Get(ctx context.Context, nums []int) (map[int]models.ResponseSentLinks, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make(map[int]models.ResponseSentLinks, len(nums))
	for _, n := range nums {
		if v, ok := m.data[n]; ok {
			res[n] = v
		}
	}
	return res, nil
}
