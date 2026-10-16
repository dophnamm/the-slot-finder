//go:build integration

package services

import (
	"context"
	"testing"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sharedOffsite = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:offsite@google.com\r\nSUMMARY:OOO offsite\r\nDTSTART;VALUE=DATE:20261019\r\nDTEND;VALUE=DATE:20261021\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	movedOffsite  = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:offsite@google.com\r\nSUMMARY:OOO offsite\r\nDTSTART;VALUE=DATE:20261026\r\nDTEND;VALUE=DATE:20261027\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	emptyCalendar = "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"
)

func TestCalendarFeedService_Sync_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	alex := models.Person{BaseModel: models.NewBaseModel(), Name: "Alex", Email: "alex@example.test", Role: constants.RoleSpecialist,
		WeeklyHours: 40, Active: true}
	bao := models.Person{BaseModel: models.NewBaseModel(), Name: "Bao", Email: "bao@example.test", Role: constants.RoleSpecialist,
		WeeklyHours: 40, Active: true}
	require.NoError(t, pdb.Create(&alex).Error)
	require.NoError(t, pdb.Create(&bao).Error)
	manual := models.Absence{BaseModel: models.NewBaseModel(), PersonID: &alex.ID, FromDate: syncNow.AddDate(0, 0, 14),
		ToDate: syncNow.AddDate(0, 0, 14), HoursPerDay: 8, Type: constants.AbsenceTypeTraining, Status: constants.AbsenceStatusApproved,
		Source: constants.AbsenceSourceManual}
	require.NoError(t, pdb.Create(&manual).Error)

	people := hoursPeople{weeklyHours: fullTime}
	authz := ProvideAuthzService(people)
	tx := db.ProvideTxManager(pdb)
	recorder := ProvideEventService(repositories.ProvideChangeEventRepository(pdb), repositories.ProvideProjectRepository(pdb),
		authz, clock.NewFixed(syncNow))
	availability := ProvideAvailabilityService(repositories.ProvideAbsenceRepository(pdb), people, authz, tx, recorder)
	fetcher := &fakeFetcher{body: sharedOffsite}
	service, err := ProvideCalendarFeedService(repositories.ProvideCalendarFeedRepository(pdb), people, availability, authz, tx,
		clock.NewFixed(syncNow), feedConfig(testFeedKey), fetcher)
	require.NoError(t, err)
	planner := asRole("kc-planner", constants.RolePlanner)

	calendarLeave := func(t *testing.T, personID string) []models.Absence {
		t.Helper()
		var absences []models.Absence
		require.NoError(t, pdb.Where("person_id = ? AND source = ?", personID, constants.AbsenceSourceCalendar).
			Order("from_date").Find(&absences).Error)
		return absences
	}
	eventTypes := func(t *testing.T) []string {
		t.Helper()
		var events []models.ChangeEvent
		require.NoError(t, pdb.Order("occurred_at, id").Find(&events).Error)
		types := make([]string, len(events))
		for i, e := range events {
			types[i] = e.Type
		}
		return types
	}

	t.Run("two people may share one event", func(t *testing.T) {
		_, err := service.Connect(planner, ConnectFeedInput{PersonID: alex.ID, URL: googleLink})
		require.NoError(t, err)
		_, err = service.Connect(planner, ConnectFeedInput{PersonID: bao.ID, URL: googleLink})
		require.NoError(t, err)
		require.Len(t, calendarLeave(t, alex.ID), 1)
		require.Len(t, calendarLeave(t, bao.ID), 1)
		assert.Equal(t, constants.AbsenceStatusPlanned, calendarLeave(t, alex.ID)[0].Status)
		assert.Equal(t, []string{constants.EventAbsenceAdded, constants.EventAbsenceAdded}, eventTypes(t))
	})

	t.Run("a moved event moves the absence and records both weeks", func(t *testing.T) {
		require.NoError(t, pdb.Model(&models.Absence{}).Where("person_id = ? AND source = ?", alex.ID, constants.AbsenceSourceCalendar).
			Update("status", constants.AbsenceStatusApproved).Error)
		fetcher.body = movedOffsite
		_, err := service.SyncNow(planner, alex.ID)
		require.NoError(t, err)
		leave := calendarLeave(t, alex.ID)
		require.Len(t, leave, 1)
		assert.Equal(t, 26, leave[0].FromDate.Day())
		assert.Equal(t, constants.AbsenceStatusPlanned, leave[0].Status, "a moved event needs approval again")
		types := eventTypes(t)
		assert.Equal(t, []string{constants.EventAbsenceRemoved, constants.EventAbsenceAdded}, types[len(types)-2:])
	})

	t.Run("a deleted event deletes the absence, and it can come back", func(t *testing.T) {
		fetcher.body = emptyCalendar
		_, err := service.SyncNow(planner, alex.ID)
		require.NoError(t, err)
		assert.Empty(t, calendarLeave(t, alex.ID))

		fetcher.body = sharedOffsite
		_, err = service.SyncNow(planner, alex.ID)
		require.NoError(t, err, "a deleted absence does not block the same event UID")
		assert.Len(t, calendarLeave(t, alex.ID), 1)
	})

	t.Run("a sync never touches an absence entered by hand", func(t *testing.T) {
		var kept models.Absence
		require.NoError(t, pdb.First(&kept, "id = ?", manual.ID).Error)
		assert.Equal(t, constants.AbsenceStatusApproved, kept.Status)
		assert.Equal(t, constants.AbsenceSourceManual, kept.Source)
	})
}

const (
	// runningLeave started last Friday and runs to this Friday (syncNow is Tue 13 Oct).
	runningLeave = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:running@google.com\r\nSUMMARY:OOO\r\nDTSTART;VALUE=DATE:20261009\r\nDTEND;VALUE=DATE:20261017\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	// trimmedLeave is the same event, shortened to end yesterday.
	trimmedLeave = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:running@google.com\r\nSUMMARY:OOO\r\nDTSTART;VALUE=DATE:20261009\r\nDTEND;VALUE=DATE:20261013\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
)

// TestCalendarFeedService_KeepsLeave_Postgres covers the ways a sync must not lose leave.
func TestCalendarFeedService_KeepsLeave_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	tx := db.ProvideTxManager(pdb)
	absences := repositories.ProvideAbsenceRepository(pdb)
	feeds := repositories.ProvideCalendarFeedRepository(pdb)
	planner := asRole("kc-planner", constants.RolePlanner)

	// newService builds the services for people whose weekly hours are weeklyHours.
	newService := func(t *testing.T, fetcher *fakeFetcher, weeklyHours float64) CalendarFeedService {
		t.Helper()
		people := hoursPeople{weeklyHours: weeklyHours}
		authz := ProvideAuthzService(people)
		recorder := ProvideEventService(repositories.ProvideChangeEventRepository(pdb), repositories.ProvideProjectRepository(pdb),
			authz, clock.NewFixed(syncNow))
		service, err := ProvideCalendarFeedService(feeds, people, ProvideAvailabilityService(absences, people, authz, tx, recorder),
			authz, tx, clock.NewFixed(syncNow), feedConfig(testFeedKey), fetcher)
		require.NoError(t, err)
		return service
	}
	newPerson := func(t *testing.T, name string) models.Person {
		t.Helper()
		person := models.Person{BaseModel: models.NewBaseModel(), Name: name, Email: name + "@example.test",
			Role: constants.RoleSpecialist, WeeklyHours: fullTime, Active: true}
		require.NoError(t, pdb.Create(&person).Error)
		return person
	}
	leaveOf := func(t *testing.T, personID string) []models.Absence {
		t.Helper()
		var found []models.Absence
		require.NoError(t, pdb.Where("person_id = ?", personID).Order("from_date").Find(&found).Error)
		return found
	}
	approveAll := func(t *testing.T, personID string) {
		t.Helper()
		require.NoError(t, pdb.Model(&models.Absence{}).Where("person_id = ?", personID).
			Update("status", constants.AbsenceStatusApproved).Error)
	}

	t.Run("running leave shortened to end yesterday is shortened, not deleted", func(t *testing.T) {
		cam := newPerson(t, "cam")
		fetcher := &fakeFetcher{body: runningLeave}
		service := newService(t, fetcher, fullTime)
		_, err := service.Connect(planner, ConnectFeedInput{PersonID: cam.ID, URL: googleLink})
		require.NoError(t, err)
		require.Len(t, leaveOf(t, cam.ID), 1)

		fetcher.body = trimmedLeave
		_, err = service.SyncNow(planner, cam.ID)
		require.NoError(t, err)
		leave := leaveOf(t, cam.ID)
		require.Len(t, leave, 1, "the days already taken stay")
		assert.Equal(t, 9, leave[0].FromDate.Day())
		assert.Equal(t, 12, leave[0].ToDate.Day())
	})

	t.Run("new weekly hours change the leave hours and keep the approval", func(t *testing.T) {
		dana := newPerson(t, "dana")
		fetcher := &fakeFetcher{body: sharedOffsite}
		_, err := newService(t, fetcher, fullTime).Connect(planner, ConnectFeedInput{PersonID: dana.ID, URL: googleLink})
		require.NoError(t, err)
		approveAll(t, dana.ID)

		const partTime = 32
		_, err = newService(t, fetcher, partTime).SyncNow(planner, dana.ID)
		require.NoError(t, err)
		leave := leaveOf(t, dana.ID)
		require.Len(t, leave, 1)
		assert.InDelta(t, 6.4, leave[0].HoursPerDay, 0.001)
		assert.Equal(t, constants.AbsenceStatusApproved, leave[0].Status)
	})

	t.Run("disconnecting keeps the leave as ordinary absences, which another calendar leaves alone", func(t *testing.T) {
		eve := newPerson(t, "eve")
		fetcher := &fakeFetcher{body: sharedOffsite}
		service := newService(t, fetcher, fullTime)
		_, err := service.Connect(planner, ConnectFeedInput{PersonID: eve.ID, URL: googleLink})
		require.NoError(t, err)
		approveAll(t, eve.ID)

		require.NoError(t, service.Disconnect(planner, eve.ID))
		leave := leaveOf(t, eve.ID)
		require.Len(t, leave, 1)
		assert.Equal(t, constants.AbsenceSourceManual, leave[0].Source)
		assert.Nil(t, leave[0].ExternalUID)

		fetcher.body = emptyCalendar
		_, err = service.Connect(planner, ConnectFeedInput{PersonID: eve.ID, URL: googleLink})
		require.NoError(t, err)
		leave = leaveOf(t, eve.ID)
		require.Len(t, leave, 1, "connecting another calendar keeps the leave")
		assert.Equal(t, constants.AbsenceStatusApproved, leave[0].Status)
	})
}

// A second sync of the same calendar, even from another connection, cannot take the lock.
func TestCalendarFeedRepository_TryLockSync_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	feeds := repositories.ProvideCalendarFeedRepository(pdb)
	tx := db.ProvideTxManager(pdb)
	const feedID = "feed-1"

	err := tx.Do(context.Background(), func(ctx context.Context) error {
		first, err := feeds.TryLockSync(ctx, feedID)
		require.NoError(t, err)
		assert.True(t, first)
		// A call outside this transaction runs on another connection.
		second, err := feeds.TryLockSync(context.Background(), feedID)
		require.NoError(t, err)
		assert.False(t, second, "the calendar is locked until the first transaction ends")
		other, err := feeds.TryLockSync(ctx, "feed-2")
		require.NoError(t, err)
		assert.True(t, other, "another calendar is not affected")
		return nil
	})
	require.NoError(t, err)
}
