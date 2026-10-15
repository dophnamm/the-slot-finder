package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockChangeEventRepository struct {
	mock.Mock
}

func (m *MockChangeEventRepository) FindAll(ctx context.Context, pr *dtos.PageableRequest, preloads ...string) (*dtos.DataResponse[models.ChangeEvent], error) {
	args := m.Called(ctx, pr, preloads)
	return args.Get(0).(*dtos.DataResponse[models.ChangeEvent]), args.Error(1)
}

func (m *MockChangeEventRepository) FindOneByID(ctx context.Context, id string) (*models.ChangeEvent, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ChangeEvent), args.Error(1)
}

func (m *MockChangeEventRepository) Create(ctx context.Context, entity *models.ChangeEvent) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockChangeEventRepository) Save(ctx context.Context, entity *models.ChangeEvent) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockChangeEventRepository) SaveAll(ctx context.Context, entities []models.ChangeEvent) error {
	return m.Called(ctx, entities).Error(0)
}

func (m *MockChangeEventRepository) Delete(ctx context.Context, entity *models.ChangeEvent) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockChangeEventRepository) Updates(ctx context.Context, entity *models.ChangeEvent) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockChangeEventRepository) List(ctx context.Context, projectIDs []string, page *dtos.PageableRequest) (*dtos.DataResponse[models.ChangeEvent], error) {
	args := m.Called(ctx, projectIDs, page)
	return args.Get(0).(*dtos.DataResponse[models.ChangeEvent]), args.Error(1)
}

func (m *MockChangeEventRepository) ClaimPending(ctx context.Context, now, cutoff time.Time, limit int) ([]models.ChangeEvent, error) {
	args := m.Called(ctx, now, cutoff, limit)
	return args.Get(0).([]models.ChangeEvent), args.Error(1)
}

func (m *MockChangeEventRepository) Finish(ctx context.Context, id, status string, draftID *string, projectIDs []string) error {
	return m.Called(ctx, id, status, draftID, projectIDs).Error(0)
}

func (m *MockChangeEventRepository) RelinkDraft(ctx context.Context, fromDraftID, toDraftID string) error {
	return m.Called(ctx, fromDraftID, toDraftID).Error(0)
}

func (m *MockChangeEventRepository) ListByDraft(ctx context.Context, draftID string) ([]models.ChangeEvent, error) {
	args := m.Called(ctx, draftID)
	return args.Get(0).([]models.ChangeEvent), args.Error(1)
}

func (m *MockChangeEventRepository) RecordFailure(ctx context.Context, ids []string, message string, now time.Time, backoff time.Duration, maxAttempts int) error {
	return m.Called(ctx, ids, message, now, backoff, maxAttempts).Error(0)
}

func (m *MockChangeEventRepository) LockReplanning(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// newEmptyEventRepository links no events to any draft.
func newEmptyEventRepository() *MockChangeEventRepository {
	repo := new(MockChangeEventRepository)
	repo.On("ListByDraft", mock.Anything, mock.Anything).Return([]models.ChangeEvent{}, nil)
	return repo
}

var eventTime = time.Date(2026, time.October, 12, 9, 30, 0, 0, time.UTC)

func TestEventService_Record(t *testing.T) {
	repo := new(MockChangeEventRepository)
	var saved *models.ChangeEvent
	repo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) { saved = args.Get(1).(*models.ChangeEvent) }).Return(nil)
	service := ProvideEventService(repo, nil, ProvideAuthzService(nil), clock.NewFixed(eventTime))

	err := service.Record(context.Background(), ChangeInput{Type: constants.EventPhaseSlipped, ProjectIDs: []string{projectID},
		Payload: PhaseChange{ProjectID: projectID, Phase: "Build", OldEndWeek: "2026-W47", NewEndWeek: "2026-W49", Weeks: 2}})
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, constants.EventStatusPending, saved.Status)
	assert.Equal(t, eventTime, saved.OccurredAt)
	assert.Equal(t, eventTime, saved.AvailableAt, "available as soon as it happens; the runner applies the debounce")
	assert.Equal(t, "Build", saved.Payload["phase"])
	assert.InDelta(t, 2, saved.Payload["weeks"], 0)
	assert.Equal(t, []string{projectID}, saved.ProjectIDs)
}

func TestEventService_List(t *testing.T) {
	lead := asRole("kc-lead", constants.RoleLead)

	t.Run("a program keeps the events of its projects", func(t *testing.T) {
		events, projects := new(MockChangeEventRepository), new(MockProjectRepository)
		projects.On("ListByProgram", mock.Anything, programID).Return([]models.Project{{BaseModel: models.BaseModel{ID: projectID}}}, nil)
		events.On("List", mock.Anything, []string{projectID}, mock.Anything).
			Return(&dtos.DataResponse[models.ChangeEvent]{Data: []models.ChangeEvent{{Type: constants.EventAbsenceAdded}}}, nil)
		page, err := ProvideEventService(events, projects, ProvideAuthzService(leadPeople{}), clock.NewFixed(eventTime)).
			List(lead, ListEventsInput{ProgramID: programID, Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Len(t, page.Data, 1)
	})

	t.Run("a program without projects has no events", func(t *testing.T) {
		events, projects := new(MockChangeEventRepository), new(MockProjectRepository)
		projects.On("ListByProgram", mock.Anything, programID).Return([]models.Project{}, nil)
		page, err := ProvideEventService(events, projects, ProvideAuthzService(leadPeople{}), clock.NewFixed(eventTime)).
			List(lead, ListEventsInput{ProgramID: programID, Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Empty(t, page.Data)
		events.AssertNotCalled(t, "List", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("a specialist is refused", func(t *testing.T) {
		_, err := ProvideEventService(nil, nil, ProvideAuthzService(leadPeople{}), clock.NewFixed(eventTime)).
			List(asRole("kc-specialist", constants.RoleSpecialist), ListEventsInput{})
		requireStatus(t, err, http.StatusForbidden)
	})
}

func TestAvailabilityService_RecordsAbsenceEvents(t *testing.T) {
	from := time.Date(2026, time.October, 19, 0, 0, 0, 0, time.UTC)
	in := CreateAbsenceInput{PersonID: selfPersonID, FromDate: from, ToDate: from.AddDate(0, 0, 1), HoursPerDay: 8,
		Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusApproved}
	people := stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: selfPersonID}}}

	t.Run("adding leave records AbsenceAdded", func(t *testing.T) {
		repo, events := new(MockAbsenceRepository), &recordedEvents{}
		repo.On("Create", mock.Anything, mock.Anything).Return(nil)
		_, err := ProvideAvailabilityService(repo, people, ProvideAuthzService(people), passthroughTx{}, events).
			Create(asRole("kc-planner", constants.RolePlanner), in)
		require.NoError(t, err)
		require.Len(t, events.events, 1)
		assert.Equal(t, constants.EventAbsenceAdded, events.events[0].Type)
		payload := events.events[0].Payload.(AbsenceChange)
		assert.Equal(t, "2026-10-19", payload.FromDate)
		assert.Equal(t, constants.AbsenceStatusApproved, payload.Status)
	})

	t.Run("a failed event fails the save", func(t *testing.T) {
		repo, events := new(MockAbsenceRepository), &recordedEvents{fail: stderrors.New("outbox down")}
		repo.On("Create", mock.Anything, mock.Anything).Return(nil)
		_, err := ProvideAvailabilityService(repo, people, ProvideAuthzService(people), passthroughTx{}, events).
			Create(asRole("kc-planner", constants.RolePlanner), in)
		require.Error(t, err)
	})

	t.Run("removing a holiday records AbsenceRemoved", func(t *testing.T) {
		repo, events := new(MockAbsenceRepository), &recordedEvents{}
		holiday := &models.Absence{BaseModel: models.BaseModel{ID: "h1"}, FromDate: from, ToDate: from, Type: constants.AbsenceTypeHoliday}
		repo.On("FindOneByID", mock.Anything, "h1").Return(holiday, nil)
		repo.On("Delete", mock.Anything, holiday).Return(nil)
		err := ProvideAvailabilityService(repo, people, ProvideAuthzService(people), passthroughTx{}, events).
			Delete(asRole("kc-planner", constants.RolePlanner), "h1")
		require.NoError(t, err)
		require.Len(t, events.events, 1)
		assert.Equal(t, constants.EventAbsenceRemoved, events.events[0].Type)
		assert.Nil(t, events.events[0].Payload.(AbsenceChange).PersonID)
	})
}

func TestProjectService_RecordsSlips(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)

	t.Run("a later phase end records PhaseSlipped", func(t *testing.T) {
		repo, events := new(MockProjectRepository), &recordedEvents{}
		project := buildAndTest(t)
		build := project.Phases[0]
		repo.On("FindPhase", mock.Anything, buildID).Return(&build, nil)
		repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
		repo.On("SavePhases", mock.Anything, mock.Anything).Return(nil)
		people := knownExpertisePeople{}
		_, err := ProvideProjectService(repo, nil, people, ProvideAuthzService(people), passthroughTx{}, events).
			UpdatePhase(planner, buildID, UpdatePhaseInput{EndWeek: new(mustWeek(t, "2026-W49"))})
		require.NoError(t, err)
		require.Len(t, events.events, 1)
		change := events.events[0].Payload.(PhaseChange)
		assert.Equal(t, constants.EventPhaseSlipped, events.events[0].Type)
		assert.Equal(t, 2, change.Weeks)
		assert.Equal(t, 1, change.MovedPhases)
	})

	t.Run("a moved deadline records ProjectSlipped, an unchanged one nothing", func(t *testing.T) {
		people := knownExpertisePeople{}
		for _, tt := range []struct {
			name     string
			deadline time.Time
			want     int
		}{
			{name: "moved", deadline: time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC), want: 1},
			{name: "unchanged", deadline: time.Date(2026, time.December, 18, 0, 0, 0, 0, time.UTC), want: 0},
		} {
			repo, events := new(MockProjectRepository), &recordedEvents{}
			repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
			repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			_, err := ProvideProjectService(repo, nil, people, ProvideAuthzService(people), passthroughTx{}, events).
				Update(planner, projectID, UpdateProjectInput{Deadline: &tt.deadline})
			require.NoError(t, err, tt.name)
			assert.Len(t, events.events, tt.want, tt.name)
		}
	})
}

func TestProjectService_RecordsStatusChange(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	people := knownExpertisePeople{}
	tests := []struct {
		name       string
		from       string
		action     string
		wantStatus string
	}{
		{name: "cancel", from: constants.ProjectStatusActive, action: constants.ProjectActionCancel, wantStatus: constants.ProjectStatusCancelled},
		{name: "hold", from: constants.ProjectStatusActive, action: constants.ProjectActionHold, wantStatus: constants.ProjectStatusOnHold},
		{name: "complete", from: constants.ProjectStatusActive, action: constants.ProjectActionComplete, wantStatus: constants.ProjectStatusDone},
		{name: "resume", from: constants.ProjectStatusOnHold, action: constants.ProjectActionResume, wantStatus: constants.ProjectStatusIncoming},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, events := new(MockProjectRepository), &recordedEvents{}
			project := buildAndTest(t)
			project.Status = tt.from
			repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
			repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			_, err := ProvideProjectService(repo, nil, people, ProvideAuthzService(people), passthroughTx{}, events).
				ApplyAction(planner, projectID, tt.action)
			require.NoError(t, err)
			require.Len(t, events.events, 1)
			assert.Equal(t, constants.EventProjectStatusChanged, events.events[0].Type)
			assert.Equal(t, []string{projectID}, events.events[0].ProjectIDs)
			change := events.events[0].Payload.(StatusChange)
			assert.Equal(t, tt.from, change.OldStatus)
			assert.Equal(t, tt.wantStatus, change.NewStatus)
		})
	}

	t.Run("a failed event fails the action", func(t *testing.T) {
		repo, events := new(MockProjectRepository), &recordedEvents{fail: stderrors.New("outbox down")}
		repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
		repo.On("Save", mock.Anything, mock.Anything).Return(nil)
		_, err := ProvideProjectService(repo, nil, people, ProvideAuthzService(people), passthroughTx{}, events).
			ApplyAction(planner, projectID, constants.ProjectActionCancel)
		require.Error(t, err)
	})
}
