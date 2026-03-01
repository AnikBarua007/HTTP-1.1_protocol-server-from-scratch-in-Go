package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")

		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 23, n)
		assert.False(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host:\t   localhost:42069   \t\r\n\r\n")

		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 30, n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["X-Existing"] = "yes"
		data := []byte("Host: localhost:42069\r\nUser-Agent: curl/7.81.0\r\n\r\n")

		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "yes", headers["X-Existing"])
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Empty(t, headers["User-Agent"])
		assert.Equal(t, 23, n)
		assert.False(t, done)
	})

	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("\r\n")

		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.True(t, done)
		assert.Len(t, headers, 0)
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("       Host : localhost:42069       \r\n\r\n")

		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
		assert.Len(t, headers, 0)
	})

	t.Run("Invalid value spacing around colon", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost : 42069\r\n\r\n")

		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
		assert.Len(t, headers, 0)
	})

	t.Run("Invalid key character", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Ho(st: localhost:42069\r\n\r\n")

		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
		assert.Len(t, headers, 0)
	})
}
