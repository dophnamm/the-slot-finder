package monitoring

import (
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
)

func TestScrubEvent(t *testing.T) {
	const link = "https://calendar.google.com/calendar/ical/x%40y/private-abc123/basic.ics"
	event := &sentry.Event{
		Message:   "fetch failed for " + link,
		Exception: []sentry.Exception{{Value: "get " + link + " timeout"}},
		Tags:      map[string]string{"url": link, "env": "prod"},
		Request: &sentry.Request{
			Data:        `{"url":"` + link + `"}`,
			QueryString: "token=abc",
			Cookies:     "session=1",
			Headers: map[string]string{"Authorization": "Bearer x", "X-CSRF-Token": "t", "X-Api-Key": "k", "X-Client-Secret": "s",
				"Cookie": "c", "Content-Type": "application/json"},
		},
	}

	scrubbed := scrubEvent(event, []string{"calendar.google.com", "outlook.office365.com"})

	assert.Empty(t, scrubbed.Request.Data)
	assert.Empty(t, scrubbed.Request.QueryString)
	assert.Empty(t, scrubbed.Request.Cookies)
	assert.Equal(t, map[string]string{"Content-Type": "application/json"}, scrubbed.Request.Headers)
	assert.Equal(t, "fetch failed for [calendar link]", scrubbed.Message)
	assert.Equal(t, "get [calendar link] timeout", scrubbed.Exception[0].Value)
	assert.Equal(t, map[string]string{"url": "[calendar link]", "env": "prod"}, scrubbed.Tags)
}

func TestScrubEvent_WithoutRequestOrHosts(t *testing.T) {
	event := &sentry.Event{Message: "plain"}
	assert.Equal(t, "plain", scrubEvent(event, nil).Message)
}
