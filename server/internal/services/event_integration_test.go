//go:build integration

package services

import (
	stderrors "errors"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvailabilityService_Outbox_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	person := models.Person{BaseModel: models.NewBaseModel(), Name: "Quinn", Email: "quinn@example.test", Role: constants.RoleSpecialist,
		WeeklyHours: 40, Active: true}
	require.NoError(t, pdb.Create(&person).Error)
	people := stubPeople{self: &person}
	absences := repositories.ProvideAbsenceRepository(pdb)
	recorder := ProvideEventService(repositories.ProvideChangeEventRepository(pdb), repositories.ProvideProjectRepository(pdb),
		ProvideAuthzService(people), clock.NewFixed(eventTime))
	from := time.Date(2026, time.October, 19, 0, 0, 0, 0, time.UTC)
	in := CreateAbsenceInput{PersonID: person.ID, FromDate: from, ToDate: from.AddDate(0, 0, 1), HoursPerDay: 8,
		Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusApproved}
	planner := asRole("kc-planner", constants.RolePlanner)
	count := func(t *testing.T, model any) int64 {
		t.Helper()
		var n int64
		require.NoError(t, pdb.Model(model).Count(&n).Error)
		return n
	}

	t.Run("a failed event leaves no absence", func(t *testing.T) {
		failing := &recordedEvents{fail: stderrors.New("outbox down")}
		_, err := ProvideAvailabilityService(absences, people, ProvideAuthzService(people), db.ProvideTxManager(pdb), failing).
			Create(planner, in)
		require.Error(t, err)
		assert.Equal(t, int64(0), count(t, &models.Absence{}))
	})

	t.Run("an absence and its event commit together", func(t *testing.T) {
		_, err := ProvideAvailabilityService(absences, people, ProvideAuthzService(people), db.ProvideTxManager(pdb), recorder).
			Create(planner, in)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count(t, &models.Absence{}))
		var events []models.ChangeEvent
		require.NoError(t, pdb.Find(&events).Error)
		require.Len(t, events, 1)
		assert.Equal(t, constants.EventAbsenceAdded, events[0].Type)
		assert.Equal(t, constants.EventStatusPending, events[0].Status)
		assert.Equal(t, person.ID, events[0].Payload["person_id"])
	})
}
