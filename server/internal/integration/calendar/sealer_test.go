package calendar

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testLink = "https://calendar.google.com/calendar/ical/a%40b/private-xyz/basic.ics"

func testKey(fill byte) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Repeat(string(fill), keyBytes)))
}

func TestSealer(t *testing.T) {
	sealer, err := NewSealer(testKey('k'))
	require.NoError(t, err)

	sealed, err := sealer.Seal(testLink)
	require.NoError(t, err)
	assert.NotContains(t, string(sealed), "private-xyz")

	again, err := sealer.Seal(testLink)
	require.NoError(t, err)
	assert.NotEqual(t, sealed, again, "each seal uses a fresh nonce")

	opened, err := sealer.Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, testLink, opened)

	other, err := NewSealer(testKey('o'))
	require.NoError(t, err)
	_, err = other.Open(sealed)
	assert.Error(t, err, "another key cannot open it")
}

func TestNewSealer_BadKey(t *testing.T) {
	for _, key := range []string{"", "not base64!", base64.StdEncoding.EncodeToString([]byte("short"))} {
		_, err := NewSealer(key)
		assert.ErrorIs(t, err, ErrBadKey, key)
	}
}
