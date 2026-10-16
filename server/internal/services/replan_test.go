package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func allocationsOf(t *testing.T, weekHours map[string]float64) []models.Allocation {
	out := make([]models.Allocation, 0, len(weekHours))
	for week, hours := range weekHours {
		out = append(out, models.Allocation{Week: mustWeek(t, week), Hours: hours})
	}
	return out
}

func projectDemand(t *testing.T, project, from, to string) *models.Demand {
	return &models.Demand{ProjectID: project, Phase: &models.ProjectPhase{StartWeek: mustWeek(t, from), EndWeek: mustWeek(t, to)}}
}

func TestConflictingProjects(t *testing.T) {
	assignments := []models.Assignment{
		{PersonID: "ada", Demand: projectDemand(t, "A", "2026-W42", "2026-W45"), Allocations: allocationsOf(t, map[string]float64{"2026-W42": 30})},
		{PersonID: "ada", Demand: projectDemand(t, "B", "2026-W42", "2026-W45"), Allocations: allocationsOf(t, map[string]float64{"2026-W42": 10})},
		{PersonID: "bea", Demand: projectDemand(t, "C", "2026-W42", "2026-W45"), Allocations: allocationsOf(t, map[string]float64{"2026-W43": 10})},
		{PersonID: "cal", Demand: projectDemand(t, "D", "2026-W42", "2026-W45"), Allocations: allocationsOf(t, map[string]float64{"2026-W50": 8})},
		{PersonID: "gone", Demand: projectDemand(t, "E", "2026-W42", "2026-W45"), Allocations: allocationsOf(t, map[string]float64{"2026-W44": 8})},
	}
	capacityOf := func(person string, _ isoweek.Week) float64 {
		if person == "gone" {
			return 0
		}
		return 36
	}
	assert.Equal(t, []string{"A", "B", "D", "E"}, conflictingProjects(assignments, capacityOf, 1),
		"over capacity (A, B share the week), outside the window (D), a person who left (E)")
}

func TestTouchedProjects(t *testing.T) {
	assignments := []models.Assignment{
		{PersonID: "ada", Demand: &models.Demand{ProjectID: "A"}, Allocations: allocationsOf(t, map[string]float64{"2026-W42": 30})},
		{PersonID: "bea", Demand: &models.Demand{ProjectID: "C"}, Allocations: allocationsOf(t, map[string]float64{"2026-W43": 10})},
	}
	events := []models.ChangeEvent{
		{BaseModel: models.BaseModel{ID: "slip"}, ProjectIDs: []string{"X"}},
		{BaseModel: models.BaseModel{ID: "leave"}, Payload: map[string]any{"person_id": "ada", "from_date": "2026-10-12", "to_date": "2026-10-13"}},
		{BaseModel: models.BaseModel{ID: "holiday"}, Payload: map[string]any{"from_date": "2026-10-21", "to_date": "2026-10-21"}},
		{BaseModel: models.BaseModel{ID: "broken"}, Payload: map[string]any{}},
	}
	touched := touchedProjects(events, assignments)
	assert.Equal(t, []string{"X"}, touched["slip"])
	assert.Equal(t, []string{"A"}, touched["leave"])
	assert.Equal(t, []string{"C"}, touched["holiday"], "a holiday touches everyone who works that week")
	assert.Empty(t, touched["broken"])
}

func TestHandEditsSurviveAMerge(t *testing.T) {
	planner := leadPersonID
	w42, w43 := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")
	byHand := models.Assignment{PersonID: "bea", DemandID: "d1", CreatedBy: &planner, Kind: constants.AssignmentKindCrossExpertise,
		Demand:      &models.Demand{ProjectID: "A"},
		Allocations: []models.Allocation{{Week: w42, Hours: 8}, {Week: w43, Hours: 8, Locked: true}}}
	empty := models.Assignment{PersonID: "cal", DemandID: "d1", CreatedBy: &planner, Kind: constants.AssignmentKindStretch, Demand: &models.Demand{ProjectID: "A"}}

	kept := keepHandEdits(byHand)
	assert.True(t, kept.Allocations[0].Locked && kept.Allocations[1].Locked, "the engine must keep every hand-added week")
	assert.False(t, byHand.Allocations[0].Locked, "the open draft itself is not changed")

	open := &models.Plan{Assignments: []models.Assignment{byHand, empty}}
	draft := &models.Plan{Assignments: []models.Assignment{
		{PersonID: "bea", DemandID: "d1", Kind: constants.AssignmentKindCrossExpertise, Allocations: kept.Allocations},
	}}
	restoreHandEdits(draft, open, map[string]bool{"A": true})
	require.Len(t, draft.Assignments, 2)
	assert.Equal(t, &planner, draft.Assignments[0].CreatedBy)
	assert.False(t, draft.Assignments[0].Allocations[0].Locked, "an unlocked hand-added week is unlocked again")
	assert.True(t, draft.Assignments[0].Allocations[1].Locked, "a locked week stays locked")
	assert.Equal(t, "cal", draft.Assignments[1].PersonID, "an assignment without hours is carried over")
}

func replanConfig() *config.Config {
	cfg := testPlanningConfig()
	cfg.ReplanDebounce, cfg.ReplanBatchSize, cfg.ReplanMaxAttempts, cfg.ReplanRetryBackoff = 5*time.Minute, 50, 5, 30*time.Second
	cfg.PlanningReplanStabilityWeight = 2
	return cfg
}

func newReplanService(plans *MockPlanRepository, events *MockChangeEventRepository, projects *MockProjectRepository) ReplanService {
	return ProvideReplanService(plans, events, projects, new(MockPersonRepository), new(MockAbsenceRepository), passthroughTx{},
		clock.NewFixed(eventTime), replanConfig(), nil)
}

func TestReplanService_RunOnce(t *testing.T) {
	cutoff := eventTime.Add(-5 * time.Minute)
	leave := models.ChangeEvent{BaseModel: models.BaseModel{ID: "e1"}, Type: constants.EventAbsenceAdded,
		Payload: map[string]any{"person_id": "ada", "from_date": "2026-10-12", "to_date": "2026-10-13"}}

	t.Run("only events older than the debounce are claimed", func(t *testing.T) {
		events := new(MockChangeEventRepository)
		events.On("LockReplanning", mock.Anything).Return(nil)
		events.On("ClaimPending", mock.Anything, eventTime, cutoff, 50).Return([]models.ChangeEvent{}, nil)
		n, err := newReplanService(new(MockPlanRepository), events, nil).RunOnce(t.Context())
		require.NoError(t, err)
		assert.Zero(t, n)
		events.AssertExpectations(t)
	})

	t.Run("a single failed event counts an attempt on that event only", func(t *testing.T) {
		events, plans := new(MockChangeEventRepository), new(MockPlanRepository)
		events.On("LockReplanning", mock.Anything).Return(nil)
		events.On("ClaimPending", mock.Anything, eventTime, cutoff, 50).Return([]models.ChangeEvent{leave}, nil)
		plans.On("FindBaseline", mock.Anything).Return(nil, stderrors.New("database down"))
		events.On("RecordFailure", mock.Anything, []string{"e1"}, mock.Anything, eventTime, 30*time.Second, 5).Return(nil)
		n, err := newReplanService(plans, events, nil).RunOnce(t.Context())
		require.Error(t, err)
		assert.Equal(t, 1, n)
		events.AssertExpectations(t)
	})

	t.Run("a failed batch is retried one event at a time, then the batch size returns", func(t *testing.T) {
		second := models.ChangeEvent{BaseModel: models.BaseModel{ID: "e2"}, Type: leave.Type, Payload: leave.Payload}
		events, plans := new(MockChangeEventRepository), new(MockPlanRepository)
		events.On("LockReplanning", mock.Anything).Return(nil)
		events.On("ClaimPending", mock.Anything, eventTime, cutoff, 50).Return([]models.ChangeEvent{leave, second, leave}, nil).Once()
		events.On("ClaimPending", mock.Anything, eventTime, cutoff, 1).Return([]models.ChangeEvent{second}, nil).Once()
		events.On("ClaimPending", mock.Anything, eventTime, cutoff, 50).Return([]models.ChangeEvent{}, nil).Once()
		plans.On("FindBaseline", mock.Anything).Return(nil, stderrors.New("boom")).Once()
		plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(1), nil).Once()
		plans.On("FindOpenAutoDraft", mock.Anything, 1).Return(nil, gorm.ErrRecordNotFound)
		events.On("Finish", mock.Anything, "e2", constants.EventStatusNoImpact, (*string)(nil), []string{}).Return(nil)
		svc := newReplanService(plans, events, new(MockProjectRepository))

		_, err := svc.RunOnce(t.Context())
		require.Error(t, err)
		events.AssertNotCalled(t, "RecordFailure", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		n, err := svc.RunOnce(t.Context())
		require.NoError(t, err)
		assert.Equal(t, 1, n)

		_, err = svc.RunOnce(t.Context())
		require.NoError(t, err)
		events.AssertExpectations(t)
	})

	t.Run("no conflict ends as no_impact", func(t *testing.T) {
		events, plans, projects := new(MockChangeEventRepository), new(MockPlanRepository), new(MockProjectRepository)
		events.On("LockReplanning", mock.Anything).Return(nil)
		events.On("ClaimPending", mock.Anything, eventTime, cutoff, 50).Return([]models.ChangeEvent{leave}, nil)
		plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(1), nil)
		plans.On("FindOpenAutoDraft", mock.Anything, 1).Return(nil, gorm.ErrRecordNotFound)
		events.On("Finish", mock.Anything, "e1", constants.EventStatusNoImpact, (*string)(nil), []string{}).Return(nil)
		n, err := newReplanService(plans, events, projects).RunOnce(t.Context())
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		events.AssertExpectations(t)
		plans.AssertNotCalled(t, "CreateWithAssignments", mock.Anything, mock.Anything)
	})
}

func TestStaffingService_SupersededDraft(t *testing.T) {
	m := newStaffingMocks()
	newer := "0192f3a4-0000-7000-8000-0000000000f9"
	old := openDraftWith()
	old.Status, old.SupersededBy = constants.PlanStatusSuperseded, &newer
	m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(old, nil)
	planner := asRole("kc-planner", constants.RolePlanner)

	_, err := m.service().Edit(planner, staffingDraftID, EditDraftInput{})
	requireAppError(t, err, constants.DraftSuperseded, http.StatusConflict)
	_, err = m.service().Accept(planner, staffingDraftID, AcceptDraftInput{BaseVersion: 2})
	requireAppError(t, err, constants.DraftSuperseded, http.StatusConflict)
	_, err = m.service().Reject(planner, staffingDraftID)
	requireAppError(t, err, constants.DraftSuperseded, http.StatusConflict)
}

func TestUnderstaffedProjects(t *testing.T) {
	demand := func(id string, hours float64) models.Demand {
		return models.Demand{BaseModel: models.BaseModel{ID: id}, TotalHours: hours}
	}
	project := func(id, status string, demands ...models.Demand) models.Project {
		return models.Project{BaseModel: models.BaseModel{ID: id}, Status: status, Demands: demands}
	}
	assignments := []models.Assignment{
		{DemandID: "full", Allocations: allocationsOf(t, map[string]float64{"2026-W44": 40})},
		{DemandID: "half", Allocations: allocationsOf(t, map[string]float64{"2026-W44": 20})},
		{DemandID: "gone", Allocations: allocationsOf(t, map[string]float64{"2026-W44": 5})},
	}
	projects := []models.Project{
		project("a-full", constants.ProjectStatusActive, demand("full", 40)),
		project("b-half", constants.ProjectStatusActive, demand("half", 40)),
		project("c-none", constants.ProjectStatusIncoming, demand("none", 10)),
		project("d-hold", constants.ProjectStatusOnHold, demand("gone", 100)),
		project("e-nodemand", constants.ProjectStatusActive),
	}
	assert.Equal(t, []string{"b-half", "c-none"}, understaffedProjects(projects, assignments),
		"a project that is not plannable or has no demand is never understaffed")
}

func TestReplanService_RunOnce_EventScope(t *testing.T) {
	cutoff := eventTime.Add(-5 * time.Minute)
	staffed := func(t *testing.T, project *models.Project) []models.Assignment {
		return []models.Assignment{{PersonID: qaPersonID, DemandID: project.Demands[0].ID,
			Allocations: allocationsOf(t, map[string]float64{"2026-W46": 20, "2026-W47": 20, "2026-W48": 20})}}
	}
	tests := []struct {
		name     string
		event    models.ChangeEvent
		baseline func(t *testing.T, project *models.Project) []models.Assignment
		want     string
	}{
		{name: "changed demand on an understaffed project plans it again",
			event: models.ChangeEvent{Type: constants.EventDemandChanged, ProjectIDs: []string{projectID}}, want: constants.EventStatusDraftCreated},
		{name: "changed demand on a staffed project is no impact",
			event: models.ChangeEvent{Type: constants.EventDemandChanged, ProjectIDs: []string{projectID}}, baseline: staffed, want: constants.EventStatusNoImpact},
		{name: "a team change plans the program's projects again",
			event: models.ChangeEvent{Type: constants.EventTeamChanged, ProjectIDs: []string{projectID}}, want: constants.EventStatusDraftCreated},
		{name: "a team change on another project is no impact",
			event: models.ChangeEvent{Type: constants.EventTeamChanged, ProjectIDs: []string{"other"}}, want: constants.EventStatusNoImpact},
		{name: "a person's capacity change may free hours for an understaffed project",
			event: models.ChangeEvent{Type: constants.EventPersonCapacityChanged, Payload: map[string]any{"person_id": qaPersonID}},
			want:  constants.EventStatusDraftCreated},
		{name: "a person's capacity change with nothing understaffed is no impact",
			event:    models.ChangeEvent{Type: constants.EventPersonCapacityChanged, Payload: map[string]any{"person_id": qaPersonID}},
			baseline: staffed, want: constants.EventStatusNoImpact},
		{name: "a cancelled project frees capacity for an understaffed one",
			event: models.ChangeEvent{Type: constants.EventProjectStatusChanged, ProjectIDs: []string{"cancelled"},
				Payload: map[string]any{"new_status": constants.ProjectStatusCancelled}}, want: constants.EventStatusDraftCreated},
		{name: "a project that is plannable again is planned",
			event: models.ChangeEvent{Type: constants.EventProjectStatusChanged, ProjectIDs: []string{projectID},
				Payload: map[string]any{"new_status": constants.ProjectStatusIncoming}}, want: constants.EventStatusDraftCreated},
		{name: "a project that is plannable again but staffed is no impact",
			event: models.ChangeEvent{Type: constants.EventProjectStatusChanged, ProjectIDs: []string{projectID},
				Payload: map[string]any{"new_status": constants.ProjectStatusIncoming}}, baseline: staffed, want: constants.EventStatusNoImpact},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.event.BaseModel = models.BaseModel{ID: "e1"}
			tt.event.OccurredAt = eventTime.Add(-time.Hour)
			events, plans, projects := new(MockChangeEventRepository), new(MockPlanRepository), new(MockProjectRepository)
			people, absences := new(MockPersonRepository), new(MockAbsenceRepository)
			project := qaProject(t)
			baseline := publishedBaseline(1)
			if tt.baseline != nil {
				baseline.Assignments = tt.baseline(t, project)
			}
			events.On("LockReplanning", mock.Anything).Return(nil)
			events.On("ClaimPending", mock.Anything, eventTime, cutoff, 50).Return([]models.ChangeEvent{tt.event}, nil)
			plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
			plans.On("FindOpenAutoDraft", mock.Anything, 1).Return(nil, gorm.ErrRecordNotFound)
			projects.On("ListForMetrics", mock.Anything, "").Return([]models.Project{*project}, nil)
			projects.On("FindDetail", mock.Anything, projectID).Return(project, nil)
			people.On("ListActive", mock.Anything).Return([]models.Person{qaPerson()}, nil)
			absences.On("ListOverlapping", mock.Anything, mock.Anything, mock.Anything).Return([]models.Absence{}, nil)
			plans.On("CreateWithAssignments", mock.Anything, mock.Anything).Return(nil).Maybe()
			var finished string
			events.On("Finish", mock.Anything, "e1", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				finished = args.String(2)
			}).Return(nil)

			service := ProvideReplanService(plans, events, projects, people, absences, passthroughTx{}, clock.NewFixed(eventTime), replanConfig(), nil)
			n, err := service.RunOnce(t.Context())
			require.NoError(t, err)
			assert.Equal(t, 1, n)
			assert.Equal(t, tt.want, finished)
		})
	}
}

// An auto re-plan in 2026-W48 plans the rest of the QA project from W48 on and leaves what
// the baseline holds in W46 and W47 as it was.
func TestReplanService_RunOnce_KeepsPastWeeks(t *testing.T) {
	cutoff := inWeek48.Add(-5 * time.Minute)
	events, plans, projects := new(MockChangeEventRepository), new(MockPlanRepository), new(MockProjectRepository)
	people, absences := new(MockPersonRepository), new(MockAbsenceRepository)
	project := qaProject(t)
	baseline := publishedBaseline(1)
	baseline.Assignments = []models.Assignment{{BaseModel: models.BaseModel{ID: draftAssignmentID}, PersonID: qaPersonID, DemandID: demandID,
		Kind: constants.AssignmentKindPrimary, Demand: &project.Demands[0],
		Allocations: allocationsOf(t, map[string]float64{"2026-W46": 10, "2026-W47": 10})}}
	event := models.ChangeEvent{BaseModel: models.BaseModel{ID: "e1"}, Type: constants.EventDemandChanged, ProjectIDs: []string{projectID},
		OccurredAt: inWeek48.Add(-time.Hour)}
	events.On("LockReplanning", mock.Anything).Return(nil)
	events.On("ClaimPending", mock.Anything, inWeek48, cutoff, 50).Return([]models.ChangeEvent{event}, nil)
	plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
	plans.On("FindOpenAutoDraft", mock.Anything, 1).Return(nil, gorm.ErrRecordNotFound)
	projects.On("ListForMetrics", mock.Anything, "").Return([]models.Project{*project}, nil)
	projects.On("FindDetail", mock.Anything, projectID).Return(project, nil)
	people.On("ListActive", mock.Anything).Return([]models.Person{qaPerson()}, nil)
	absences.On("ListOverlapping", mock.Anything, mock.Anything, mock.Anything).Return([]models.Absence{}, nil)
	var draft *models.Plan
	plans.On("CreateWithAssignments", mock.Anything, mock.Anything).Run(func(args mock.Arguments) { draft = args.Get(1).(*models.Plan) }).Return(nil)
	events.On("Finish", mock.Anything, "e1", constants.EventStatusDraftCreated, mock.Anything, mock.Anything).Return(nil)

	service := ProvideReplanService(plans, events, projects, people, absences, passthroughTx{}, clock.NewFixed(inWeek48), replanConfig(), nil)
	_, err := service.RunOnce(t.Context())
	require.NoError(t, err)

	require.NotNil(t, draft)
	require.Len(t, draft.Assignments, 1)
	weeks := hoursInWeeks(draft.Assignments[0])
	assert.InDelta(t, 10, weeks["2026-W46"], hoursEpsilon, "history stays as it was")
	assert.InDelta(t, 10, weeks["2026-W47"], hoursEpsilon, "history stays as it was")
	assert.InDelta(t, 40, weeks["2026-W48"]+weeks["2026-W49"]+weeks["2026-W50"]+weeks["2026-W51"], hoursEpsilon,
		"the 60 h demand minus the 20 h worked is planned from the current week on")
	assert.InDelta(t, 60, draft.Assignments[0].TotalHours(), hoursEpsilon)
}

func TestReplanService_Run_SurvivesAPanic(t *testing.T) {
	secondTick := make(chan struct{})
	events := new(MockChangeEventRepository)
	events.On("LockReplanning", mock.Anything).Run(func(mock.Arguments) { panic("boom") }).Return(nil).Once()
	events.On("LockReplanning", mock.Anything).Run(func(mock.Arguments) { close(secondTick) }).Return(nil).Once()
	events.On("ClaimPending", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]models.ChangeEvent{}, nil)
	cfg := replanConfig()
	cfg.ReplanPollInterval = time.Millisecond
	service := ProvideReplanService(new(MockPlanRepository), events, nil, new(MockPersonRepository), new(MockAbsenceRepository),
		passthroughTx{}, clock.NewFixed(eventTime), cfg, nil)

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() { service.Run(ctx); close(done) }()

	select {
	case <-secondTick:
	case <-time.After(5 * time.Second):
		t.Fatal("the runner stopped after a panic")
	}
	cancel()
	<-done
}
