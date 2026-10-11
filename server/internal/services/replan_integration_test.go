//go:build integration

package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// steppingClock is a clock the test moves forward.
type steppingClock struct{ now time.Time }

func (c *steppingClock) Now() time.Time           { return c.now }
func (c *steppingClock) Location() *time.Location { return time.UTC }

func day(month time.Month, d int) time.Time { return time.Date(2026, month, d, 0, 0, 0, 0, time.UTC) }

func TestReplan_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	ctx := context.Background()
	planner := asRole("kc-planner", constants.RolePlanner)

	qa := models.Expertise{BaseModel: models.NewBaseModel(), Name: "QA", Sequence: 4}
	require.NoError(t, pdb.Create(&qa).Error)
	newPerson := func(name string) models.Person {
		p := models.Person{BaseModel: models.NewBaseModel(), Name: name, Email: name + "@example.test", Role: constants.RoleSpecialist,
			WeeklyHours: 40, NonProjectHours: 4, Active: true}
		require.NoError(t, pdb.Create(&p).Error)
		require.NoError(t, pdb.Create(&models.PersonSkill{PersonID: p.ID, ExpertiseID: qa.ID, Level: 3, IsPrimary: true}).Error)
		return p
	}
	ada, bea := newPerson("ada"), newPerson("bea")
	program := models.Program{BaseModel: models.NewBaseModel(), Name: "Acme", Client: "Acme", Status: constants.ProgramStatusOpen,
		StartDate: day(time.October, 12), EndDate: day(time.December, 31), Objectives: []string{}}
	require.NoError(t, pdb.Create(&program).Error)
	project := models.Project{BaseModel: models.NewBaseModel(), ProgramID: program.ID, Name: "Checkout", Priority: 1,
		Status: constants.ProjectStatusPlanned, StartEarliest: day(time.October, 12), Deadline: day(time.October, 23)}
	require.NoError(t, pdb.Create(&project).Error)
	demand := models.Demand{BaseModel: models.NewBaseModel(), ProjectID: project.ID, ExpertiseID: qa.ID, MinLevel: 2, Headcount: 1,
		TotalHours: 72, Profile: constants.DemandProfileFlat}
	require.NoError(t, pdb.Create(&demand).Error)

	plans := repositories.ProvidePlanRepository(pdb)
	require.NoError(t, plans.EnsureBaseline(ctx))
	baseline, err := plans.FindBaselineSummary(ctx)
	require.NoError(t, err)
	require.NoError(t, plans.CreateAssignment(ctx, &models.Assignment{BaseModel: models.NewBaseModel(), PlanID: baseline.ID,
		PersonID: ada.ID, DemandID: demand.ID, Kind: constants.AssignmentKindPrimary, Reason: []string{},
		Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W42"), Hours: 36}, {Week: mustWeek(t, "2026-W43"), Hours: 36}}}))

	t0 := time.Date(2026, time.October, 11, 9, 0, 0, 0, time.UTC)
	clk := &steppingClock{now: t0}
	cfg := replanConfig()
	tx := db.ProvideTxManager(pdb)
	people := stubPeople{self: &ada}
	authz := ProvideAuthzService(people)
	eventRepo, projectRepo := repositories.ProvideChangeEventRepository(pdb), repositories.ProvideProjectRepository(pdb)
	personRepo, absenceRepo := repositories.ProvidePersonRepository(pdb), repositories.ProvideAbsenceRepository(pdb)
	events := ProvideEventService(eventRepo, projectRepo, authz, clk)
	availability := ProvideAvailabilityService(absenceRepo, people, authz, tx, events)
	replan := ProvideReplanService(plans, eventRepo, projectRepo, personRepo, absenceRepo, tx, clk, cfg)
	staffing := ProvideStaffingService(plans, projectRepo, personRepo, absenceRepo, repositories.ProvideAuditRepository(pdb), eventRepo,
		leadPeople{}, ProvideAuthzService(leadPeople{}), tx, clk, cfg)
	leave := func(person models.Person, from, to time.Time) {
		t.Helper()
		_, err := availability.Create(planner, CreateAbsenceInput{PersonID: person.ID, FromDate: from, ToDate: to, HoursPerDay: 8,
			Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusApproved})
		require.NoError(t, err)
	}
	allEvents := func(t *testing.T) []models.ChangeEvent {
		t.Helper()
		var rows []models.ChangeEvent
		require.NoError(t, pdb.Order("occurred_at ASC").Find(&rows).Error)
		return rows
	}

	// Two leaves within the debounce: nothing before it passes, then one draft for both.
	leave(ada, day(time.October, 12), day(time.October, 13))
	clk.now = t0.Add(time.Minute)
	leave(bea, day(time.October, 21), day(time.October, 21))
	n, err := replan.RunOnce(ctx)
	require.NoError(t, err)
	assert.Zero(t, n, "events younger than the debounce wait")

	clk.now = t0.Add(7 * time.Minute)
	n, err = replan.RunOnce(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	rows := allEvents(t)
	require.Len(t, rows, 2)
	require.NotNil(t, rows[0].DraftID)
	firstDraft := *rows[0].DraftID
	for _, e := range rows {
		assert.Equal(t, constants.EventStatusDraftCreated, e.Status)
		assert.Equal(t, firstDraft, *e.DraftID, "one draft linked to both leaves")
	}

	page, err := events.List(asRole("kc-lead", constants.RoleLead), ListEventsInput{ProgramID: program.ID, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.NotEmpty(t, page.Data, "the program's history drives the banner")
	assert.Equal(t, constants.PlanStatusDraft, page.Data[0].Draft.Status)

	// The Planner locks a week of the draft; a third leave merges into a new version.
	view, err := staffing.Get(planner, firstDraft)
	require.NoError(t, err)
	assert.Len(t, view.Events, 2)
	var locked models.Assignment
	for _, a := range view.Plan.Assignments {
		if a.DemandID == demand.ID && len(a.Allocations) > 0 {
			locked = a
			break
		}
	}
	require.NotEmpty(t, locked.ID)
	week, hours := locked.Allocations[0].Week, locked.Allocations[0].Hours
	_, err = staffing.Edit(planner, firstDraft, EditDraftInput{Set: []SetAllocationInput{
		{AssignmentID: locked.ID, Week: week, Hours: hours, Locked: new(true)}}})
	require.NoError(t, err)

	clk.now = t0.Add(8 * time.Minute)
	leave(ada, day(time.October, 19), day(time.October, 19))
	clk.now = t0.Add(14 * time.Minute)
	n, err = replan.RunOnce(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	rows = allEvents(t)
	require.Len(t, rows, 3)
	secondDraft := *rows[2].DraftID
	assert.NotEqual(t, firstDraft, secondDraft)
	for _, e := range rows {
		assert.Equal(t, secondDraft, *e.DraftID, "every event follows the newest version")
	}
	old, err := plans.FindOneByID(ctx, firstDraft)
	require.NoError(t, err)
	assert.Equal(t, constants.PlanStatusSuperseded, old.Status)
	assert.Equal(t, secondDraft, *old.SupersededBy)

	merged, err := plans.FindWithAssignments(ctx, secondDraft)
	require.NoError(t, err)
	kept := false
	for _, a := range merged.Assignments {
		for _, alloc := range a.Allocations {
			if a.PersonID == locked.PersonID && a.DemandID == demand.ID && alloc.Week == week && alloc.Locked && alloc.Hours == hours {
				kept = true
			}
		}
	}
	assert.True(t, kept, "the Planner's lock survives the merge")

	_, err = staffing.Edit(planner, firstDraft, EditDraftInput{})
	requireAppError(t, err, constants.DraftSuperseded, 409)
}

func TestChangeEvents_ClaimAndFail_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	ctx := context.Background()
	repo := repositories.ProvideChangeEventRepository(pdb)
	tx := db.ProvideTxManager(pdb)
	now := time.Date(2026, time.October, 11, 9, 0, 0, 0, time.UTC)
	for i := range 10 {
		at := now.Add(-time.Hour + time.Duration(i)*time.Minute)
		require.NoError(t, repo.Create(ctx, &models.ChangeEvent{BaseModel: models.NewBaseModel(), Type: constants.EventAbsenceAdded,
			Payload: map[string]any{}, OccurredAt: at, AvailableAt: at, Status: constants.EventStatusPending, ProjectIDs: []string{}}))
	}

	t.Run("two runners never claim the same event", func(t *testing.T) {
		var first, second []models.ChangeEvent
		firstClaimed, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tx.Do(ctx, func(ctx context.Context) error {
				var err error
				first, err = repo.ClaimPending(ctx, now, now, 5)
				close(firstClaimed)
				<-release
				return err
			})
		}()
		go func() {
			defer close(done)
			<-firstClaimed
			_ = tx.Do(ctx, func(ctx context.Context) error {
				var err error
				second, err = repo.ClaimPending(ctx, now, now, 10)
				return err
			})
		}()
		<-done
		close(release)
		wg.Wait()

		require.Len(t, first, 5)
		require.Len(t, second, 5, "the second runner skips the locked rows")
		seen := map[string]bool{}
		for _, e := range append(first, second...) {
			assert.False(t, seen[e.ID], "claimed twice")
			seen[e.ID] = true
		}
	})

	t.Run("an event that keeps failing ends as failed", func(t *testing.T) {
		var event models.ChangeEvent
		require.NoError(t, pdb.Order("occurred_at ASC").First(&event).Error)
		for range 5 {
			require.NoError(t, repo.RecordFailure(ctx, []string{event.ID}, "engine exploded", now, 30*time.Second, 5))
		}
		var after models.ChangeEvent
		require.NoError(t, pdb.First(&after, "id = ?", event.ID).Error)
		assert.Equal(t, constants.EventStatusFailed, after.Status)
		assert.Equal(t, 5, after.Attempts)
		require.NotNil(t, after.LastError)
		assert.Equal(t, "engine exploded", *after.LastError)
		assert.True(t, after.AvailableAt.After(now), "each failure moves the next attempt later")
	})
}
