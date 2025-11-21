package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/models"
)

// FileStorage хранит данные в памяти и периодически сбрасывает их в JSON-файл.
// При старте сервера пытается прочитать данные из файла.
type FileStorage struct {
	mu   sync.RWMutex
	path string
	data map[int]models.ResponseSentLinks
}

// NewFileStorage создаёт файловое хранилище.
// Если файл существует — читаем данные, если нет — начинаем с пустой мапы.
func NewFileStorage(path string) (*FileStorage, error) {
	fs := &FileStorage{
		path: path,
		data: make(map[int]models.ResponseSentLinks),
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// файла нет — ок, стартуем с пустой мапы
			return fs, nil
		}
		return nil, err
	}

	if info.IsDir() {
		return nil, errors.New("file storage path is a directory")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return fs, nil
	}

	if err := json.Unmarshal(b, &fs.data); err != nil {
		return nil, err
	}

	return fs, nil
}

func (f *FileStorage) Save(ctx context.Context, resp models.ResponseSentLinks) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.data[resp.Num] = resp
	return f.flush()
}

func (f *FileStorage) Get(ctx context.Context, nums []int) (map[int]models.ResponseSentLinks, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	res := make(map[int]models.ResponseSentLinks, len(nums))
	for _, n := range nums {
		if v, ok := f.data[n]; ok {
			res[n] = v
		}
	}
	return res, nil
}

// flush сбрасывает всю мапу в JSON-файл через временный файл.
func (f *FileStorage) flush() error {
	b, err := json.MarshalIndent(f.data, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(f.path)
	tmpFile, err := os.CreateTemp(dir, "linkchecker-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	_, err = tmpFile.Write(b)
	closeErr := tmpFile.Close()
	if err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return closeErr
	}

	if err := os.Rename(tmpPath, f.path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

var _ Storage = (*FileStorage)(nil)
