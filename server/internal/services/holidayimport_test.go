package services

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/integration/calendar"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const holidayICS = `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
DTSTART;VALUE=DATE:20260902
DTEND;VALUE=DATE:20260903
UID:national@google.com
SUMMARY:National Day
END:VEVENT
BEGIN:VEVENT
DTSTART;VALUE=DATE:20270206
DTEND;VALUE=DATE:20270211
UID:tet@google.com
SUMMARY:Lunar New Year Holidays
END:VEVENT
END:VCALENDAR
`

// fetchedCalendar is a calendar fetcher that returns these bytes and remembers the link.
type fetchedCalendar struct {
	data    []byte
	fetched []string
}

func (f *fetchedCalendar) Fetch(_ context.Context, rawURL string) ([]byte, error) {
	f.fetched = append(f.fetched, rawURL)
	return f.data, nil
}

func newHolidayImportService(repo *MockAbsenceRepository, fetcher calendar.Fetcher, events *recordedEvents) HolidayImportService {
	cfg := heatmapConfig()
	cfg.CalendarAllowedHosts, cfg.CalendarMaxBytes = []string{"calendar.google.com"}, 4096
	return ProvideHolidayImportService(repo, ProvideAuthzService(leadPeople{}), passthroughTx{}, events, fetcher,
		clock.NewFixed(time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC)), cfg)
}

func TestHolidayImportService_Import(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	from, to := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, time.December, 31, 0, 0, 0, 0, time.UTC)
	known := "national@google.com"

	t.Run("a file adds each new holiday by name, and skips one imported before", func(t *testing.T) {
		repo, events := new(MockAbsenceRepository), &recordedEvents{}
		repo.On("ListHolidays", mock.Anything).Return([]models.Absence{{ExternalUID: &known}}, nil)
		repo.On("Create", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
			return a.Name == "Lunar New Year Holidays" && a.PersonID == nil && a.Type == constants.AbsenceTypeHoliday &&
				a.Status == constants.AbsenceStatusApproved && a.Source == constants.AbsenceSourceCalendar && a.HoursPerDay == 8 &&
				*a.ExternalUID == "tet@google.com" && a.ToDate.Equal(time.Date(2027, time.February, 10, 0, 0, 0, 0, time.UTC))
		})).Return(nil).Once()

		result, err := newHolidayImportService(repo, &fetchedCalendar{}, events).Import(planner, ImportHolidaysInput{
			File: strings.NewReader(holidayICS), From: from, To: to, HoursPerDay: 8})

		require.NoError(t, err)
		assert.Equal(t, HolidayImportResult{Added: 1, Skipped: 1}, *result)
		require.Len(t, events.events, 1)
		assert.Equal(t, constants.EventAbsenceAdded, events.events[0].Type)
		repo.AssertExpectations(t)
	})

	t.Run("a link from an allowed host is fetched", func(t *testing.T) {
		repo, fetcher := new(MockAbsenceRepository), &fetchedCalendar{data: []byte(holidayICS)}
		repo.On("ListHolidays", mock.Anything).Return([]models.Absence{}, nil)
		repo.On("Create", mock.Anything, mock.Anything).Return(nil).Twice()
		link := "https://calendar.google.com/calendar/ical/en.vietnamese%23holiday%40group.v.calendar.google.com/public/basic.ics"

		result, err := newHolidayImportService(repo, fetcher, &recordedEvents{}).Import(planner, ImportHolidaysInput{
			URL: link, From: from, To: to, HoursPerDay: 8})

		require.NoError(t, err)
		assert.Equal(t, 2, result.Added)
		assert.Equal(t, []string{link}, fetcher.fetched)
	})

	t.Run("refusals", func(t *testing.T) {
		big := strings.Repeat("X", 5000)
		tests := []struct {
			name     string
			in       ImportHolidaysInput
			wantCode string
		}{
			{name: "neither a file nor a link", in: ImportHolidaysInput{From: from, To: to, HoursPerDay: 8}, wantCode: constants.HolidaySourceInvalid},
			{name: "both a file and a link", in: ImportHolidaysInput{File: strings.NewReader(holidayICS), URL: "https://calendar.google.com/x.ics",
				From: from, To: to, HoursPerDay: 8}, wantCode: constants.HolidaySourceInvalid},
			{name: "a host that is not allowed", in: ImportHolidaysInput{URL: "https://example.com/x.ics", From: from, To: to, HoursPerDay: 8},
				wantCode: constants.CalendarFeedHostNotAllowed},
			{name: "not a calendar", in: ImportHolidaysInput{File: strings.NewReader("date,name\n"), From: from, To: to, HoursPerDay: 8},
				wantCode: constants.CalendarFeedNotICS},
			{name: "a file over the size limit", in: ImportHolidaysInput{File: strings.NewReader(big), From: from, To: to, HoursPerDay: 8},
				wantCode: constants.CalendarFeedTooLarge},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := newHolidayImportService(new(MockAbsenceRepository), &fetchedCalendar{}, &recordedEvents{}).Import(planner, tt.in)
				requireAppError(t, err, tt.wantCode, http.StatusBadRequest)
			})
		}
	})

	t.Run("only planners import", func(t *testing.T) {
		_, err := newHolidayImportService(new(MockAbsenceRepository), &fetchedCalendar{}, &recordedEvents{}).Import(
			asRole("kc-manager", constants.RoleManager), ImportHolidaysInput{File: strings.NewReader(holidayICS), From: from, To: to, HoursPerDay: 8})
		requireStatus(t, err, http.StatusForbidden)
	})
}
