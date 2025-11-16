package storage

import (
	"context"
	"testing"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryStorage(t *testing.T) {
	st := NewMemoryStorage()
	require.NotNil(t, st)
	require.NotNil(t, st.data)
}

func TestMemoryStorage_SaveAndGet(t *testing.T) {
	s := NewMemoryStorage()
	ctx := context.Background()

	resp := models.ResponseSentLinks{
		Num: 1,
		Links: map[string]string{
			"google.com": "available",
			"ya.ru":      "available",
		},
	}

	err := s.Save(ctx, resp)
	require.NoError(t, err)

	out, err := s.Get(ctx, []int{1})
	require.NoError(t, err)

	require.Len(t, out, 1)

	got, ok := out[1]
	require.True(t, ok, "result must contain key 1")

	assert.Equal(t, resp.Num, got.Num)
	assert.Equal(t, resp.Links, got.Links)
}

func TestMemoryStorage_GetNotFound(t *testing.T) {
	s := NewMemoryStorage()
	ctx := context.Background()

	err := s.Save(ctx, models.ResponseSentLinks{
		Num: 1,
		Links: map[string]string{
			"google.com": "available",
		},
	})
	require.NoError(t, err)

	out, err := s.Get(ctx, []int{99})
	require.NoError(t, err)
	assert.Len(t, out, 0)
}

func TestMemoryStorage_GetEmptyNums(t *testing.T) {
	s := NewMemoryStorage()
	ctx := context.Background()

	out, err := s.Get(ctx, []int{})
	require.NoError(t, err)
	assert.Len(t, out, 0)
}
