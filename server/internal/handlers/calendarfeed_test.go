package handlers

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const calendarLink = "https://calendar.google.com/calendar/ical/a%40b/private-secret/basic.ics"

// fakeFeeds records the link it is given and answers with a connected calendar.
type fakeFeeds struct {
	services.CalendarFeedService
	connected *services.ConnectFeedInput
}

func (f *fakeFeeds) Connect(_ context.Context, in services.ConnectFeedInput) (*models.CalendarFeed, error) {
	f.connected = &in
	return &models.CalendarFeed{Host: "calendar.google.com", URLCiphertext: []byte(in.URL), EventsFound: 2}, nil
}

func TestCalendarFeedHandler_Connect(t *testing.T) {
	route := constants.PathPersonCalendarFeed
	target := strings.Replace(route, ":id", testPersonID, 1)
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "connects a trimmed link", body: `{"url":"  ` + calendarLink + `  "}`, wantStatus: http.StatusOK},
		{name: "not a link", body: `{"url":"calendar"}`, wantStatus: http.StatusBadRequest},
		{name: "empty", body: `{}`, wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feeds := &fakeFeeds{}
			h := ProvideCalendarFeedHandler(feeds, newTestValidator(t))
			rec, _ := serve(t, route, http.MethodPut, target, tt.body, h.Connect)

			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantStatus != http.StatusOK {
				assert.Nil(t, feeds.connected)
				return
			}
			require.NotNil(t, feeds.connected)
			assert.Equal(t, calendarLink, feeds.connected.URL)
			assert.Equal(t, testPersonID, feeds.connected.PersonID)
			assert.NotContains(t, rec.Body.String(), "private-secret", "the response never holds the link")
		})
	}
}
