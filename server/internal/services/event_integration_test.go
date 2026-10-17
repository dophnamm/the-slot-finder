//go:build integration

package services

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
	"the-slot-finder/server/internal/request"

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

func TestEventService_ListScopesManager_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	person := func(name string) models.Person {
		p := models.Person{BaseModel: models.NewBaseModel(), Name: name, Email: name + "@example.test", Role: constants.RoleSpecialist, WeeklyHours: 40, Active: true}
		require.NoError(t, pdb.Create(&p).Error)
		return p
	}
	manager, teammate, stranger := person("manager"), person("teammate"), person("stranger")
	day := time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC)
	program := func(name string, lead *string) models.Program {
		p := models.Program{BaseModel: models.NewBaseModel(), Name: name, Client: "c", Status: constants.ProgramStatusOpen, StartDate: day, EndDate: day.AddDate(0, 6, 0), LeadID: lead}
		require.NoError(t, pdb.Create(&p).Error)
		return p
	}
	mine, other := program("mine", &manager.ID), program("other", nil)
	project := func(prog models.Program) models.Project {
		p := models.Project{BaseModel: models.NewBaseModel(), ProgramID: prog.ID, Name: prog.Name, Priority: 3, Status: constants.ProjectStatusActive, StartEarliest: day, Deadline: day.AddDate(0, 3, 0)}
		require.NoError(t, pdb.Create(&p).Error)
		return p
	}
	mineProject, otherProject := project(mine), project(other)
	require.NoError(t, pdb.Create(&models.ProgramTeamMember{ProgramID: mine.ID, PersonID: teammate.ID}).Error)
	event := func(payload map[string]any, projectIDs ...string) {
		require.NoError(t, pdb.Create(&models.ChangeEvent{BaseModel: models.NewBaseModel(), Type: constants.EventAbsenceAdded, Payload: payload,
			OccurredAt: day, Status: constants.EventStatusPending, AvailableAt: day, ProjectIDs: projectIDs}).Error)
	}
	event(map[string]any{"person_id": teammate.ID}, otherProject.ID)
	event(map[string]any{"person_id": stranger.ID}, mineProject.ID)
	event(map[string]any{"project_id": mineProject.ID}, mineProject.ID)
	event(map[string]any{"project_id": otherProject.ID}, otherProject.ID)

	people := managerPeople{self: &manager}
	service := ProvideEventService(repositories.ProvideChangeEventRepository(pdb), repositories.ProvideProjectRepository(pdb),
		ProvideAuthzService(people), clock.NewFixed(eventTime))
	lead := asRole("kc-manager", constants.RoleManager)

	page, err := service.List(lead, ListEventsInput{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, page.Data, 2, "the teammate's event and the event on the manager's project; not the stranger's or the other program's")
	for _, e := range page.Data {
		assert.NotEqual(t, stranger.ID, e.Payload["person_id"])
		assert.NotEqual(t, otherProject.ID, e.Payload["project_id"])
	}

	page, err = service.List(asRole("kc-planner", constants.RolePlanner), ListEventsInput{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, page.Data, 4, "a planner sees everything")
}

// managerPeople resolves every principal to one manager.
type managerPeople struct {
	PeopleService
	self *models.Person
}

func (p managerPeople) FindForPrincipal(context.Context, request.Principal) (*models.Person, error) {
	return p.self, nil
}
