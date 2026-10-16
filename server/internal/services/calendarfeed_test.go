package services

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/integration/calendar"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	googleLink  = "https://calendar.google.com/calendar/ical/a%40b/private-secret/basic.ics"
	feedRowID   = "0192f3a4-0000-7000-8000-0000000000f1"
	allDayOOO   = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:ooo\r\nSUMMARY:OOO\r\nDTSTART;VALUE=DATE:20261019\r\nDTEND;VALUE=DATE:20261021\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	halfDayOOO  = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:half\r\nSUMMARY:Out of office\r\nDTSTART:20261019T130000Z\r\nDTEND:20261019T163000Z\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	testFeedKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=" // 32 bytes, for tests only
)

// syncNow is a Wednesday, 13 Oct 2026 09:00 UTC.
var syncNow = time.Date(2026, time.October, 13, 9, 0, 0, 0, time.UTC)

// fakeFeeds keeps calendar feeds in memory.
type fakeFeeds struct {
	repositories.CalendarFeedRepository
	byPerson map[string]*models.CalendarFeed
	claimed  bool
	// locked holds the ids of feeds whose sync another server is running.
	locked map[string]bool
}

func (f *fakeFeeds) FindByPerson(_ context.Context, personID string) (*models.CalendarFeed, error) {
	if feed, ok := f.byPerson[personID]; ok {
		copied := *feed
		return &copied, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeFeeds) Save(_ context.Context, feed *models.CalendarFeed) error {
	copied := *feed
	f.byPerson[feed.PersonID] = &copied
	return nil
}

func (f *fakeFeeds) RecordSync(_ context.Context, id string, at time.Time, eventsFound int, errorCode *string) error {
	for _, feed := range f.byPerson {
		if feed.ID == id {
			feed.LastError = errorCode
			if errorCode == nil {
				feed.LastSyncedAt, feed.EventsFound = &at, eventsFound
			}
		}
	}
	return nil
}

func (f *fakeFeeds) ClaimDue(_ context.Context, now, next time.Time, _ int) ([]models.CalendarFeed, error) {
	f.claimed = true
	var due []models.CalendarFeed
	for _, feed := range f.byPerson {
		if !feed.NextSyncAt.After(now) {
			feed.NextSyncAt = next
			due = append(due, *feed)
		}
	}
	return due, nil
}

func (f *fakeFeeds) TryLockSync(_ context.Context, feedID string) (bool, error) {
	if f.locked[feedID] {
		return false, nil
	}
	return true, nil
}

func (f *fakeFeeds) DeleteByPerson(_ context.Context, personID string) error {
	delete(f.byPerson, personID)
	return nil
}

// fakeFetcher returns body for any link and remembers the links it was asked for.
type fakeFetcher struct {
	body  string
	err   error
	links []string
}

func (f *fakeFetcher) Fetch(_ context.Context, link string) ([]byte, error) {
	f.links = append(f.links, link)
	return []byte(f.body), f.err
}

// fakeLeave records the leave a sync applies, and the people whose calendar leave was detached.
type fakeLeave struct {
	AvailabilityService
	applied  []ApplyCalendarLeaveInput
	detached []string
}

func (f *fakeLeave) DetachCalendarLeave(_ context.Context, personID string) error {
	f.detached = append(f.detached, personID)
	return nil
}

func (f *fakeLeave) ApplyCalendarLeave(_ context.Context, in ApplyCalendarLeaveInput) (CalendarLeaveResult, error) {
	f.applied = append(f.applied, in)
	return CalendarLeaveResult{Created: len(in.Leave)}, nil
}

// hoursPeople returns a person with weeklyHours for any id and links the specialist to selfPersonID.
type hoursPeople struct {
	stubPeople
	weeklyHours float64
}

func (p hoursPeople) Get(_ context.Context, id string) (*models.Person, error) {
	return &models.Person{BaseModel: models.BaseModel{ID: id}, Name: "Alex", WeeklyHours: p.weeklyHours}, nil
}

// fullTime is the weekly hours of the people in these tests, unless a test says otherwise.
const fullTime = 40

func feedConfig(key string) *config.Config {
	return &config.Config{CalendarFeedKey: key, CalendarAllowedHosts: []string{"calendar.google.com"},
		CalendarSyncInterval: 15 * time.Minute, CalendarSyncBatchSize: 20, CalendarSyncWeeks: 26,
		CalendarLeaveKeywords: []string{"OOO", "Out of office"},
		WorkingWeekdays:       []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}}
}

func newFeedService(t *testing.T, key string, feeds *fakeFeeds, fetcher *fakeFetcher, leave *fakeLeave) CalendarFeedService {
	t.Helper()
	return newFeedServiceFor(t, key, feeds, fetcher, leave, fullTime)
}

func newFeedServiceFor(t *testing.T, key string, feeds *fakeFeeds, fetcher *fakeFetcher, leave *fakeLeave, weeklyHours float64) CalendarFeedService {
	t.Helper()
	people := hoursPeople{stubPeople: stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: selfPersonID}}}, weeklyHours: weeklyHours}
	s, err := ProvideCalendarFeedService(feeds, people, leave, ProvideAuthzService(people), passthroughTx{},
		clock.NewFixed(syncNow), feedConfig(key), fetcher)
	require.NoError(t, err)
	return s
}

func TestProvideCalendarFeedService_BadKey(t *testing.T) {
	_, err := ProvideCalendarFeedService(nil, nil, nil, nil, nil, clock.NewFixed(syncNow),
		feedConfig(base64.StdEncoding.EncodeToString([]byte("short"))), nil)
	require.ErrorIs(t, err, calendar.ErrBadKey)
}

func TestCalendarFeedService_Connect(t *testing.T) {
	specialist := asRole("kc-specialist", constants.RoleSpecialist)
	planner := asRole("kc-planner", constants.RolePlanner)
	tests := []struct {
		name      string
		key       string
		ctx       context.Context
		personID  string
		link      string
		body      string
		fetchErr  error
		wantCode  string
		wantHTTP  int
		wantHours float64
	}{
		{name: "a person connects their own calendar: all-day OOO counts full days", key: testFeedKey, ctx: specialist,
			personID: selfPersonID, link: googleLink, body: allDayOOO, wantHours: 8},
		{name: "a timed event counts its hours", key: testFeedKey, ctx: planner, personID: otherPersonID, link: googleLink,
			body: halfDayOOO, wantHours: 3.5},
		{name: "off without a key", ctx: planner, personID: otherPersonID, link: googleLink,
			wantCode: constants.CalendarSyncDisabled, wantHTTP: http.StatusServiceUnavailable},
		{name: "someone else's calendar", key: testFeedKey, ctx: specialist, personID: otherPersonID, link: googleLink,
			wantHTTP: http.StatusForbidden},
		{name: "a host outside the list is refused before any fetch", key: testFeedKey, ctx: planner, personID: otherPersonID,
			link: "https://evil.example/basic.ics", wantCode: constants.CalendarFeedHostNotAllowed, wantHTTP: http.StatusBadRequest},
		{name: "http is refused", key: testFeedKey, ctx: planner, personID: otherPersonID,
			link: strings.Replace(googleLink, "https", "http", 1), wantCode: constants.CalendarFeedHostNotAllowed, wantHTTP: http.StatusBadRequest},
		{name: "a link that is not a calendar", key: testFeedKey, ctx: planner, personID: otherPersonID, link: googleLink,
			body: "<html>Sign in</html>", wantCode: constants.CalendarFeedNotICS, wantHTTP: http.StatusBadRequest},
		{name: "a calendar that cannot be fetched", key: testFeedKey, ctx: planner, personID: otherPersonID, link: googleLink,
			fetchErr: calendar.ErrUnreachable, wantCode: constants.CalendarFeedUnreachable, wantHTTP: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
			fetcher := &fakeFetcher{body: tt.body, err: tt.fetchErr}
			leave := &fakeLeave{}
			feed, err := newFeedService(t, tt.key, feeds, fetcher, leave).Connect(tt.ctx, ConnectFeedInput{PersonID: tt.personID, URL: tt.link})

			if tt.wantHTTP != 0 {
				if tt.wantCode != "" {
					requireAppError(t, err, tt.wantCode, tt.wantHTTP)
				} else {
					requireStatus(t, err, tt.wantHTTP)
				}
				assert.NotContains(t, err.Error(), "private-secret", "an error must never hold the link")
				assert.Empty(t, feeds.byPerson, "nothing is saved")
				assert.Empty(t, leave.applied)
				if tt.wantCode == constants.CalendarFeedHostNotAllowed {
					assert.Empty(t, fetcher.links, "a refused link is never fetched")
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "calendar.google.com", feed.Host)
			assert.NotContains(t, string(feed.URLCiphertext), "private-secret", "the link is stored encrypted")
			assert.Equal(t, syncNow.Add(15*time.Minute), feed.NextSyncAt)
			require.NotNil(t, feed.LastSyncedAt)
			assert.Equal(t, 1, feed.EventsFound)
			require.Len(t, leave.applied, 1)
			require.Len(t, leave.applied[0].Leave, 1)
			assert.InDelta(t, tt.wantHours, leave.applied[0].Leave[0].HoursPerDay, 1e-9)
			assert.Equal(t, time.Date(2026, time.October, 13, 0, 0, 0, 0, time.UTC), leave.applied[0].From)
			assert.Equal(t, time.Date(2027, time.April, 13, 0, 0, 0, 0, time.UTC), leave.applied[0].To)
		})
	}
}

func TestCalendarFeedService_SyncNowRefusesWhileAnotherSyncRuns(t *testing.T) {
	feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
	fetcher := &fakeFetcher{body: allDayOOO}
	service := newFeedService(t, testFeedKey, feeds, fetcher, &fakeLeave{})
	planner := asRole("kc-planner", constants.RolePlanner)
	_, err := service.Connect(planner, ConnectFeedInput{PersonID: otherPersonID, URL: googleLink})
	require.NoError(t, err)
	fetched := len(fetcher.links)
	feeds.locked = map[string]bool{feeds.byPerson[otherPersonID].ID: true}

	_, err = service.SyncNow(planner, otherPersonID)

	requireAppError(t, err, constants.CalendarSyncRunning, http.StatusConflict)
	assert.Len(t, fetcher.links, fetched, "a locked calendar is not fetched")
}

func TestCalendarFeedService_SyncKeepsAbsencesWhenTheFetchFails(t *testing.T) {
	feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
	fetcher := &fakeFetcher{body: allDayOOO}
	leave := &fakeLeave{}
	service := newFeedService(t, testFeedKey, feeds, fetcher, leave)
	planner := asRole("kc-planner", constants.RolePlanner)
	_, err := service.Connect(planner, ConnectFeedInput{PersonID: otherPersonID, URL: googleLink})
	require.NoError(t, err)

	fetcher.err = calendar.ErrUnreachable
	feed, err := service.SyncNow(planner, otherPersonID)

	require.NoError(t, err)
	require.NotNil(t, feed.LastError)
	assert.Equal(t, constants.CalendarFeedUnreachable, *feed.LastError)
	assert.Equal(t, 1, feed.EventsFound, "the last good sync stays")
	assert.Len(t, leave.applied, 1, "a failed fetch applies nothing")
	assert.Equal(t, []string{googleLink, googleLink}, fetcher.links, "the stored link opens back to the same link")
}

func TestCalendarFeedService_RunOnce(t *testing.T) {
	t.Run("does nothing while off", func(t *testing.T) {
		feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
		n, err := newFeedService(t, "", feeds, &fakeFetcher{}, &fakeLeave{}).RunOnce(context.Background())
		require.NoError(t, err)
		assert.Zero(t, n)
		assert.False(t, feeds.claimed)
	})

	t.Run("syncs the calendars that are due, for the system", func(t *testing.T) {
		feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
		fetcher := &fakeFetcher{body: allDayOOO}
		leave := &fakeLeave{}
		service := newFeedService(t, testFeedKey, feeds, fetcher, leave)
		_, err := service.Connect(asRole("kc-planner", constants.RolePlanner), ConnectFeedInput{PersonID: otherPersonID, URL: googleLink})
		require.NoError(t, err)
		feeds.byPerson[otherPersonID].ID = feedRowID
		feeds.byPerson[otherPersonID].NextSyncAt = syncNow

		n, err := service.RunOnce(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Len(t, leave.applied, 2, "connect, then the runner")
		assert.Equal(t, syncNow.Add(15*time.Minute), feeds.byPerson[otherPersonID].NextSyncAt)
	})
}

func TestCalendarFeedService_GetAndDisconnect(t *testing.T) {
	feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
	leave := &fakeLeave{}
	service := newFeedService(t, testFeedKey, feeds, &fakeFetcher{body: allDayOOO}, leave)
	specialist := asRole("kc-specialist", constants.RoleSpecialist)

	_, err := service.Get(specialist, selfPersonID)
	requireAppError(t, err, constants.CalendarFeedNotFound, http.StatusNotFound)

	_, err = service.Connect(specialist, ConnectFeedInput{PersonID: selfPersonID, URL: googleLink})
	require.NoError(t, err)
	_, err = service.Get(asRole("kc-lead", constants.RoleLead), selfPersonID)
	requireStatus(t, err, http.StatusForbidden)

	require.NoError(t, service.Disconnect(specialist, selfPersonID))
	assert.Empty(t, feeds.byPerson)
	assert.Equal(t, []string{selfPersonID}, leave.detached, "the leave it added stays, as ordinary absences")
}

// TestCalendarFeedService_SeesLeaveShortenedBeforeToday: leave running since last week whose
// event now ends before today must still reach the availability service, so the absence is
// shortened rather than deleted. It is not counted as an event found in the sync window.
func TestCalendarFeedService_SeesLeaveShortenedBeforeToday(t *testing.T) {
	const trimmed = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:trimmed\r\nSUMMARY:OOO\r\nDTSTART;VALUE=DATE:20261009\r\n" +
		"DTEND;VALUE=DATE:20261013\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
	leave := &fakeLeave{}
	service := newFeedService(t, testFeedKey, feeds, &fakeFetcher{body: trimmed}, leave)

	feed, err := service.Connect(asRole("kc-planner", constants.RolePlanner), ConnectFeedInput{PersonID: selfPersonID, URL: googleLink})

	require.NoError(t, err)
	require.Len(t, leave.applied, 1)
	require.Len(t, leave.applied[0].Leave, 1)
	got := leave.applied[0].Leave[0]
	assert.Equal(t, "trimmed", got.UID)
	assert.Equal(t, time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC), got.ToDate)
	assert.Equal(t, 0, feed.EventsFound)
}

func TestCalendarFeedService_LeaveHours(t *testing.T) {
	const oddLength = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:odd\r\nSUMMARY:OOO dentist\r\nDTSTART:20261019T090000Z\r\nDTEND:20261019T113500Z\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	tests := []struct {
		name        string
		weeklyHours float64
		body        string
		wantHours   []float64
	}{
		{name: "an odd length is rounded to two decimals, as the database stores it", weeklyHours: fullTime, body: oddLength,
			wantHours: []float64{2.58}},
		{name: "a person with no weekly hours gets no zero-hour leave", weeklyHours: 0, body: allDayOOO, wantHours: []float64{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leave := &fakeLeave{}
			service := newFeedServiceFor(t, testFeedKey, &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}},
				&fakeFetcher{body: tt.body}, leave, tt.weeklyHours)
			_, err := service.Connect(asRole("kc-planner", constants.RolePlanner), ConnectFeedInput{PersonID: otherPersonID, URL: googleLink})
			require.NoError(t, err)
			require.Len(t, leave.applied, 1)
			hours := []float64{}
			for _, l := range leave.applied[0].Leave {
				hours = append(hours, l.HoursPerDay)
			}
			assert.Equal(t, tt.wantHours, hours)
		})
	}
}
