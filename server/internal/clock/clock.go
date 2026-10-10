// Package clock gives the current time in the organization's time zone.
// Code asks a Clock for "now" instead of calling time.Now, so tests can pin it.
package clock

import (
	"fmt"
	"time"

	"the-slot-finder/server/internal/config"
)

// Clock reports the current time in the organization's time zone.
type Clock interface {
	Now() time.Time
	Location() *time.Location
}

type systemClock struct {
	loc *time.Location
}

// ProvideClock builds the system clock from the TIMEZONE setting.
func ProvideClock(cfg *config.Config) (Clock, error) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", cfg.Timezone, err)
	}
	return systemClock{loc: loc}, nil
}

func (c systemClock) Now() time.Time { return time.Now().In(c.loc) }

func (c systemClock) Location() *time.Location { return c.loc }

type fixedClock struct {
	now time.Time
}

// NewFixed returns a clock that always reports now, in now's location. Use it in tests.
func NewFixed(now time.Time) Clock {
	return fixedClock{now: now}
}

func (c fixedClock) Now() time.Time { return c.now }

func (c fixedClock) Location() *time.Location { return c.now.Location() }
