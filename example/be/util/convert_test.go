package util

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertFromMapTo_NotFound(t *testing.T) {
	to, err := ConvertFromMapTo[string](map[string]any{}, "hello")
	assert.Nil(t, to)
	assert.NotNil(t, err)
}

func TestConvertFromMapTo_IncorrectType(t *testing.T) {
	to, err := ConvertFromMapTo[string](map[string]any{"hello": 1}, "hello")
	assert.Nil(t, to)
	assert.NotNil(t, err)
}

func TestConvertFromMapTo_OK(t *testing.T) {
	to, err := ConvertFromMapTo[string](map[string]any{"hello": "Hi!"}, "hello")
	require.NoError(t, err)
	require.NotNil(t, to)
	require.Equal(t, "Hi!", *to)
}
