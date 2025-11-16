package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAppRoutes(t *testing.T) {
	mockStore := storage.NewMemoryStorage()
	logger := zap.NewNop()
	defer logger.Sync()

	app := NewApp(mockStore, logger.Sugar())

	t.Run("Create link and Get", func(t *testing.T) {
		reqBody := `{"links":["google.com"]}`
		req, err := http.NewRequest(http.MethodPost, "/links", strings.NewReader(reqBody))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		//запускаем тестовый запрос
		rec := httptest.NewRecorder()
		app.router.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusCreated, res.StatusCode, "Expected 201 status code")

		// Get
		reqBodyGet := `{"links_list":[1]}`
		reqGet, err := http.NewRequest(http.MethodGet, "/links_num", strings.NewReader(reqBodyGet))
		require.NoError(t, err)

		reqGet.Header.Set("Content-Type", "application/json")

		recGet := httptest.NewRecorder()
		app.router.ServeHTTP(recGet, reqGet)

		resGet := recGet.Result()
		defer resGet.Body.Close()

		assert.Equal(t, http.StatusOK, resGet.StatusCode, "Expected 200 status code")
		assert.Equal(t, "application/pdf", resGet.Header.Get("Content-Type"))
	})

}
