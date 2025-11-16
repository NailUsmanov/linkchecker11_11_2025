package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockStorage struct {
	Data map[int]models.ResponseSentLinks
}

func (m *MockStorage) Save(ctx context.Context, links models.ResponseSentLinks) error {
	m.Data[links.Num] = links
	return nil
}

func (m *MockStorage) Get(ctx context.Context, num []int) (map[int]models.ResponseSentLinks, error) {
	result := make(map[int]models.ResponseSentLinks)
	for _, n := range num {
		if v, ok := m.Data[n]; ok {
			result[n] = v
		}
	}
	return result, nil
}

func TestNewCreateLinks(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantStatus  int
	}{
		{
			name:        "correct",
			requestBody: `{"links":["ya.ru","google.com"]}`,
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "invalid JSON",
			requestBody: `{"links":["ya.ru","google.com"]`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty body",
			requestBody: `{"links":[]}`,
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//Сброс атомик счетчика для будущих тестов
			reqNum.Store(0)

			storage := &MockStorage{Data: make(map[int]models.ResponseSentLinks)}
			logger := zap.NewNop()
			defer logger.Sync()
			sugar := logger.Sugar()

			handler := NewCreateLinks(storage, sugar)

			req := httptest.NewRequest(http.MethodPost, "/links", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantStatus == http.StatusCreated {
				resp := models.ResponseSentLinks{}
				err := json.NewDecoder(res.Body).Decode(&resp)
				require.NoError(t, err)

				assert.Equal(t, 1, resp.Num)
				assert.Len(t, resp.Links, 2)

				saved, ok := storage.Data[resp.Num]
				require.True(t, ok, "response not saved in storage")
				assert.Equal(t, resp, saved)
			}
		})

	}
}

func TestNewGetLinks(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantStatus  int
	}{
		{
			name:        "correct",
			requestBody: `{"links_list":[1]}`,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "invalid JSON",
			requestBody: `{"links_list":[1,2`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "not found number",
			requestBody: `{"links_list":[99]}`,
			wantStatus:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MockStorage{Data: make(map[int]models.ResponseSentLinks)}
			logger := zap.NewNop()
			defer logger.Sync()
			sugar := logger.Sugar()

			// Тестовые данные
			links := models.ResponseSentLinks{
				Links: map[string]string{
					"google.com": "available",
					"ya.ru":      "available",
					"test.com":   "not available"},
				Num: 1,
			}
			ctx := context.Background()
			err := storage.Save(ctx, links)
			require.NoError(t, err)

			handler := NewGetLinks(storage, sugar)

			req := httptest.NewRequest(http.MethodGet, "/links_num", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantStatus == http.StatusOK {

			}
		})
	}

	t.Run("wrong content type", func(t *testing.T) {
		s := &MockStorage{Data: make(map[int]models.ResponseSentLinks)}
		l := zap.NewNop()
		sugar := l.Sugar()

		h := NewGetLinks(s, sugar)
		req := httptest.NewRequest(http.MethodGet, "/links_list", strings.NewReader(`{"links_list":[1]}`))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		h(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}
