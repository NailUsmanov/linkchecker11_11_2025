package service

import (
	"context"
	"testing"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckLink_Empty(t *testing.T) {
	ctx := context.Background()

	ok, err := CheckLink(ctx, "")
	assert.False(t, ok)
	require.Error(t, err)
}

func TestCheckLink_InvalidURL(t *testing.T) {
	ctx := context.Background()

	ok, err := CheckLink(ctx, "://bad")
	assert.False(t, ok)
	require.Error(t, err)
}

func TestCheckLink_UnsupportedScheme(t *testing.T) {
	ctx := context.Background()

	ok, err := CheckLink(ctx, "ftp://example.com")
	assert.False(t, ok)
	require.Error(t, err)
}

func TestCreatePDF(t *testing.T) {
	data := map[int]models.ResponseSentLinks{
		1: {
			Num: 1,
			Links: map[string]string{
				"google.com": "available",
				"ya.ru":      "not available",
			},
		},
	}

	pdf, err := CreatePDF(data)

	require.NoError(t, err)
	require.NotNil(t, pdf)
	assert.Greater(t, len(pdf), 0)

}
