package services

import (
	"context"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func calendarAbsence(uid string, from, to time.Time, hours float64, status string) models.Absence {
	person := otherPersonID
	return models.Absence{BaseModel: models.BaseModel{ID: "abs-" + uid}, PersonID: &person, FromDate: from, ToDate: to,
		HoursPerDay: hours, Type: constants.AbsenceTypeLeave, Status: status, Source: constants.AbsenceSourceCalendar,
		ExternalUID: &uid}
}

func TestAvailabilityService_ApplyCalendarLeave(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, time.October, d, 0, 0, 0, 0, time.UTC) }
	window := ApplyCalendarLeaveInput{PersonID: otherPersonID, From: day(13), To: day(13).AddDate(0, 0, 182)}
	// kept started before the window and is still in the calendar: it stays as it is.
	kept := calendarAbsence("kept", day(9), day(16), 8, constants.AbsenceStatusApproved)
	moved := calendarAbsence("moved", day(21), day(21), 8, constants.AbsenceStatusApproved)
	gone := calendarAbsence("gone", day(26), day(27), 8, constants.AbsenceStatusPlanned)
	past := calendarAbsence("past", time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.September, 2, 0, 0, 0, 0, time.UTC), 8, constants.AbsenceStatusApproved)

	repo := new(MockAbsenceRepository)
	repo.On("ListCalendarLeave", mock.Anything, otherPersonID).Return([]models.Absence{past, kept, moved, gone}, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
		return *a.ExternalUID == "new" && a.Status == constants.AbsenceStatusPlanned && a.Source == constants.AbsenceSourceCalendar &&
			a.Type == constants.AbsenceTypeLeave && a.HoursPerDay == 4
	})).Return(nil)
	repo.On("Save", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
		return a.ID == moved.ID && a.FromDate.Equal(day(22)) && a.Status == constants.AbsenceStatusPlanned
	})).Return(nil)
	repo.On("Delete", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool { return a.ID == gone.ID })).Return(nil)
	events := &recordedEvents{}
	people := stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: selfPersonID}}}
	service := ProvideAvailabilityService(repo, people, ProvideAuthzService(people), passthroughTx{}, events)

	window.Leave = []CalendarLeave{
		{UID: "kept", FromDate: day(9), ToDate: day(16), HoursPerDay: 8},
		{UID: "moved", FromDate: day(22), ToDate: day(22), HoursPerDay: 8},
		{UID: "new", FromDate: day(23), ToDate: day(23), HoursPerDay: 4},
	}
	result, err := service.ApplyCalendarLeave(context.Background(), window)

	require.NoError(t, err)
	assert.Equal(t, CalendarLeaveResult{Created: 1, Updated: 1, Deleted: 1}, result)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool { return a.ID == past.ID }))
	var types []string
	for _, e := range events.events {
		types = append(types, e.Type)
	}
	assert.Equal(t, []string{constants.EventAbsenceRemoved, constants.EventAbsenceAdded, constants.EventAbsenceAdded,
		constants.EventAbsenceRemoved}, types, "moved: removed + added; new: added; gone: removed")
}

func TestAvailabilityService_ApplyCalendarLeave_PastAndHours(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, time.October, d, 0, 0, 0, 0, time.UTC) }
	approved := calendarAbsence("hours", day(19), day(20), 8, constants.AbsenceStatusApproved)
	repo := new(MockAbsenceRepository)
	repo.On("ListCalendarLeave", mock.Anything, otherPersonID).Return([]models.Absence{approved}, nil)
	repo.On("Save", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
		return a.ID == approved.ID && a.HoursPerDay == 6.4 && a.Status == constants.AbsenceStatusApproved
	})).Return(nil)
	people := stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: selfPersonID}}}
	service := ProvideAvailabilityService(repo, people, ProvideAuthzService(people), passthroughTx{}, &recordedEvents{})

	result, err := service.ApplyCalendarLeave(context.Background(), ApplyCalendarLeaveInput{
		PersonID: otherPersonID, From: day(13), To: day(13).AddDate(0, 0, 182),
		Leave: []CalendarLeave{
			// Last month's trip was never synced: past leave is not added.
			{UID: "old-trip", FromDate: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
				ToDate: time.Date(2026, time.September, 2, 0, 0, 0, 0, time.UTC), HoursPerDay: 8},
			// The person's weekly hours changed: same days, fewer hours, still approved.
			{UID: "hours", FromDate: day(19), ToDate: day(20), HoursPerDay: 6.4},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, CalendarLeaveResult{Updated: 1}, result)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}
