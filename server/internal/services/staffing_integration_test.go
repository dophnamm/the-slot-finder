//go:build integration

package services

import (
	"context"
	"sync"
	"testing"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func emptyDraft(t *testing.T, plans repositories.PlanRepository, version int) string {
	t.Helper()
	draft := models.Plan{BaseModel: models.NewBaseModel(), Kind: constants.PlanKindDraft, Status: constants.PlanStatusDraft,
		Version: version, ProjectIDs: []string{}, Unmet: []models.PlanUnmet{}}
	require.NoError(t, plans.CreateWithAssignments(context.Background(), &draft))
	return draft.ID
}

func TestStaffingService_Accept_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	plans := repositories.ProvidePlanRepository(pdb)
	finder := leadPeople{}
	service := ProvideStaffingService(plans, repositories.ProvideProjectRepository(pdb), repositories.ProvidePersonRepository(pdb),
		repositories.ProvideAbsenceRepository(pdb), repositories.ProvideAuditRepository(pdb), repositories.ProvideChangeEventRepository(pdb),
		finder, ProvideAuthzService(finder),
		db.ProvideTxManager(pdb), clock.NewFixed(acceptedAt), testPlanningConfig())
	planner := asRole("kc-planner", constants.RolePlanner)
	ctx := context.Background()
	require.NoError(t, plans.EnsureBaseline(ctx))

	publishedBaselines := func(t *testing.T) []models.Plan {
		t.Helper()
		var rows []models.Plan
		require.NoError(t, pdb.Where("kind = ? AND status = ?", constants.PlanKindBaseline, constants.PlanStatusPublished).Find(&rows).Error)
		return rows
	}

	first, second := emptyDraft(t, plans, 0), emptyDraft(t, plans, 0)

	t.Run("two accepts at the same time: exactly one wins", func(t *testing.T) {
		start := make(chan struct{})
		errs := make([]error, 2)
		var wg sync.WaitGroup
		for i, id := range []string{first, second} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				_, errs[i] = service.Accept(planner, id, AcceptDraftInput{BaseVersion: 0})
			}()
		}
		close(start)
		wg.Wait()

		failed := 0
		for _, err := range errs {
			if err != nil {
				failed++
				appErr := errors.GetAppError(err)
				require.NotNil(t, appErr, err)
				assert.Equal(t, constants.PlanVersionConflict, appErr.Code)
			}
		}
		assert.Equal(t, 1, failed)
		baselines := publishedBaselines(t)
		require.Len(t, baselines, 1)
		assert.Equal(t, 1, baselines[0].Version)
	})

	t.Run("a stale draft gets a conflict and leaves the baseline as it is", func(t *testing.T) {
		stale := emptyDraft(t, plans, 0)
		_, err := service.Accept(planner, stale, AcceptDraftInput{BaseVersion: 0})
		requireAppError(t, err, constants.PlanVersionConflict, 409)
		baselines := publishedBaselines(t)
		require.Len(t, baselines, 1)
		assert.Equal(t, 1, baselines[0].Version)
	})

	t.Run("the index allows one published baseline", func(t *testing.T) {
		extra := models.Plan{BaseModel: models.NewBaseModel(), Kind: constants.PlanKindBaseline, Status: constants.PlanStatusPublished,
			Version: 9, ProjectIDs: []string{}, Unmet: []models.PlanUnmet{}}
		err := plans.Create(ctx, &extra)
		assert.ErrorIs(t, err, gorm.ErrDuplicatedKey)
	})
}
