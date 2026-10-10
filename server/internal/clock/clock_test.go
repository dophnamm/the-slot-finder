package clock

import (
	"testing"
	"time"

	"the-slot-finder/server/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideClock(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{name: "utc", timezone: "UTC"},
		{name: "named zone", timezone: "Asia/Ho_Chi_Minh"},
		{name: "unknown zone", timezone: "Mars/Olympus", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ProvideClock(&config.Config{Timezone: tt.timezone})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.timezone, c.Location().String())
			assert.Equal(t, tt.timezone, c.Now().Location().String())
		})
	}
}

func TestNewFixed(t *testing.T) {
	now := time.Date(2026, time.October, 12, 9, 0, 0, 0, time.UTC)
	c := NewFixed(now)
	assert.Equal(t, now, c.Now())
	assert.Equal(t, time.UTC, c.Location())
}
