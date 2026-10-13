package dtos

import (
	"strings"
	"time"

	"the-slot-finder/server/internal/models"
)

// ConnectCalendarFeedRequest is the body of PUT /people/{id}/calendar-feed: the secret iCal
// address of the person's Google or Outlook calendar.
type ConnectCalendarFeedRequest struct {
	URL string `json:"url" validate:"required,url,max=2048" example:"https://calendar.google.com/calendar/ical/name%40example.com/private-abc/basic.ics"`
}

// Normalize trims the link.
func (r *ConnectCalendarFeedRequest) Normalize() {
	r.URL = strings.TrimSpace(r.URL)
}

// CalendarFeedResponse is a calendar's status. It never holds the link, which is a secret.
type CalendarFeedResponse struct {
	Host         string     `json:"host" example:"calendar.google.com"`
	LastSyncedAt *time.Time `json:"last_synced_at"`
	LastError    *string    `json:"last_error" example:"CALENDAR_FEED_UNREACHABLE"`
	EventsFound  int        `json:"events_found" example:"2"`
}

// NewCalendarFeedResponse maps a calendar feed without its link.
func NewCalendarFeedResponse(f models.CalendarFeed) CalendarFeedResponse {
	return CalendarFeedResponse{Host: f.Host, LastSyncedAt: f.LastSyncedAt, LastError: f.LastError, EventsFound: f.EventsFound}
}
