package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) FindAll(ctx context.Context, pr *dtos.PageableRequest, preloads ...string) (*dtos.DataResponse[models.Project], error) {
	args := m.Called(ctx, pr, preloads)
	return args.Get(0).(*dtos.DataResponse[models.Project]), args.Error(1)
}

func (m *MockProjectRepository) FindOneByID(ctx context.Context, id string) (*models.Project, error) {
	args := m.Called(ctx, id)
	return projectOrNil(args.Get(0)), args.Error(1)
}

func (m *MockProjectRepository) Create(ctx context.Context, entity *models.Project) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockProjectRepository) Save(ctx context.Context, entity *models.Project) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockProjectRepository) SaveAll(ctx context.Context, entities []models.Project) error {
	return m.Called(ctx, entities).Error(0)
}

func (m *MockProjectRepository) Delete(ctx context.Context, entity *models.Project) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockProjectRepository) Updates(ctx context.Context, entity *models.Project) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockProjectRepository) MaxPriority(ctx context.Context, programID string) (int, error) {
	args := m.Called(ctx, programID)
	return args.Int(0), args.Error(1)
}

func (m *MockProjectRepository) SetPriorities(ctx context.Context, priorities map[string]int) error {
	return m.Called(ctx, priorities).Error(0)
}

func (m *MockProjectRepository) ListByProgram(ctx context.Context, programID string) ([]models.Project, error) {
	args := m.Called(ctx, programID)
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepository) ListForMetrics(ctx context.Context, programID string) ([]models.Project, error) {
	args := m.Called(ctx, programID)
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockProjectRepository) ListByStatus(ctx context.Context, status string, readerID *string, page *dtos.PageableRequest) (*dtos.DataResponse[models.Project], error) {
	args := m.Called(ctx, status, readerID, page)
	return args.Get(0).(*dtos.DataResponse[models.Project]), args.Error(1)
}

func (m *MockProjectRepository) LeadsProjectIn(ctx context.Context, programID, personID string) (bool, error) {
	args := m.Called(ctx, programID, personID)
	return args.Bool(0), args.Error(1)
}

func (m *MockProjectRepository) FindDetail(ctx context.Context, id string) (*models.Project, error) {
	args := m.Called(ctx, id)
	return projectOrNil(args.Get(0)), args.Error(1)
}

func (m *MockProjectRepository) FindPhase(ctx context.Context, id string) (*models.ProjectPhase, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProjectPhase), args.Error(1)
}

func (m *MockProjectRepository) ListPhases(ctx context.Context, projectID string) ([]models.ProjectPhase, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]models.ProjectPhase), args.Error(1)
}

func (m *MockProjectRepository) SavePhases(ctx context.Context, phases []models.ProjectPhase) error {
	return m.Called(ctx, phases).Error(0)
}

func (m *MockProjectRepository) DeletePhases(ctx context.Context, ids []string) error {
	return m.Called(ctx, ids).Error(0)
}

func (m *MockProjectRepository) CountDemandsInPhases(ctx context.Context, phaseIDs []string) (int64, error) {
	args := m.Called(ctx, phaseIDs)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProjectRepository) CountPastBaselineHours(ctx context.Context, demandIDs []string, before isoweek.Week, minHours float64) (int64, error) {
	args := m.Called(ctx, demandIDs, before, minHours)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProjectRepository) ListDemands(ctx context.Context, projectID string) ([]models.Demand, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]models.Demand), args.Error(1)
}

func (m *MockProjectRepository) SaveDemands(ctx context.Context, demands []models.Demand) error {
	return m.Called(ctx, demands).Error(0)
}

func (m *MockProjectRepository) DeleteDemands(ctx context.Context, ids []string) error {
	return m.Called(ctx, ids).Error(0)
}

func (m *MockProjectRepository) CancelOpen(ctx context.Context, programID string) ([]string, error) {
	args := m.Called(ctx, programID)
	ids, _ := args.Get(0).([]string)
	return ids, args.Error(1)
}

func projectOrNil(v any) *models.Project {
	if v == nil {
		return nil
	}
	return v.(*models.Project)
}

// knownExpertisePeople extends leadPeople with an expertise check that knows expQA only.
type knownExpertisePeople struct {
	leadPeople
}

func (knownExpertisePeople) ExpertisesExist(_ context.Context, ids []string) (bool, error) {
	for _, id := range ids {
		if id != expQA {
			return false, nil
		}
	}
	return true, nil
}

const (
	projectID = "0192f3a4-0000-7000-8000-0000000000e1"
	buildID   = "0192f3a4-0000-7000-8000-0000000000e2"
	testID    = "0192f3a4-0000-7000-8000-0000000000e3"
	demandID  = "0192f3a4-0000-7000-8000-0000000000e4"
)

func mustWeek(t *testing.T, text string) isoweek.Week {
	t.Helper()
	w, err := isoweek.Parse(text)
	require.NoError(t, err)
	return w
}

// buildAndTest is a project with Build W42–W47 and Test W46–W51, led by the test lead.
func buildAndTest(t *testing.T) *models.Project {
	lead := leadPersonID
	return &models.Project{
		BaseModel: models.BaseModel{ID: projectID}, ProgramID: programID, Status: constants.ProjectStatusIncoming, LeadID: &lead,
		StartEarliest: time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC),
		Deadline:      time.Date(2026, time.December, 18, 0, 0, 0, 0, time.UTC),
		Program:       &models.Program{BaseModel: models.BaseModel{ID: programID}},
		Phases: []models.ProjectPhase{
			{BaseModel: models.BaseModel{ID: buildID}, ProjectID: projectID, Name: "Build", Sequence: 1, StartWeek: mustWeek(t, "2026-W42"), EndWeek: mustWeek(t, "2026-W47")},
			{BaseModel: models.BaseModel{ID: testID}, ProjectID: projectID, Name: "Test", Sequence: 2, StartWeek: mustWeek(t, "2026-W46"), EndWeek: mustWeek(t, "2026-W51")},
		},
		Demands: []models.Demand{{BaseModel: models.BaseModel{ID: demandID}, ProjectID: projectID, ExpertiseID: expQA, PhaseID: new(testID)}},
	}
}

func projectConfig() *config.Config {
	return &config.Config{ProjectWindowWeeks: 156, PlanningHourStep: 1}
}

func newProjectService(projects *MockProjectRepository, programs *MockProgramRepository) ProjectService {
	people := knownExpertisePeople{}
	return ProvideProjectService(projects, programs, people, ProvideAuthzService(people), passthroughTx{}, &recordedEvents{}, projectConfig(), clock.NewFixed(acceptedAt))
}

func TestProjectService_UpdatePhase_SlipPropagation(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)

	t.Run("a later end moves every later phase by the same weeks", func(t *testing.T) {
		repo := new(MockProjectRepository)
		project := buildAndTest(t)
		build := project.Phases[0]
		repo.On("FindPhase", mock.Anything, buildID).Return(&build, nil)
		repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
		repo.On("SavePhases", mock.Anything, mock.Anything).Return(nil)

		changed, err := newProjectService(repo, nil).UpdatePhase(planner, buildID, UpdatePhaseInput{EndWeek: new(mustWeek(t, "2026-W49"))})
		require.NoError(t, err)
		require.Len(t, changed, 2)
		assert.Equal(t, "2026-W49", changed[0].EndWeek.String())
		assert.Equal(t, "2026-W48", changed[1].StartWeek.String())
		assert.Equal(t, "2026-W53", changed[1].EndWeek.String())
		repo.AssertCalled(t, "SavePhases", mock.Anything, changed)
	})

	t.Run("an earlier end moves nothing else", func(t *testing.T) {
		repo := new(MockProjectRepository)
		project := buildAndTest(t)
		build := project.Phases[0]
		repo.On("FindPhase", mock.Anything, buildID).Return(&build, nil)
		repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
		repo.On("SavePhases", mock.Anything, mock.Anything).Return(nil)

		changed, err := newProjectService(repo, nil).UpdatePhase(planner, buildID, UpdatePhaseInput{EndWeek: new(mustWeek(t, "2026-W46"))})
		require.NoError(t, err)
		require.Len(t, changed, 1)
		assert.Equal(t, "2026-W46", changed[0].EndWeek.String())
	})

	t.Run("end before start", func(t *testing.T) {
		repo := new(MockProjectRepository)
		project := buildAndTest(t)
		build := project.Phases[0]
		repo.On("FindPhase", mock.Anything, buildID).Return(&build, nil)
		repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
		_, err := newProjectService(repo, nil).UpdatePhase(planner, buildID, UpdatePhaseInput{EndWeek: new(mustWeek(t, "2026-W40"))})
		requireAppError(t, err, constants.PhaseWeeksInvalid, http.StatusBadRequest)
		repo.AssertNotCalled(t, "SavePhases", mock.Anything, mock.Anything)
	})

	t.Run("a lead of another project is refused", func(t *testing.T) {
		repo := new(MockProjectRepository)
		project := buildAndTest(t)
		project.LeadID = new(otherPersonID2)
		build := project.Phases[0]
		repo.On("FindPhase", mock.Anything, buildID).Return(&build, nil)
		repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
		_, err := newProjectService(repo, nil).UpdatePhase(asRole("kc-lead", constants.RoleLead), buildID,
			UpdatePhaseInput{EndWeek: new(mustWeek(t, "2026-W49"))})
		requireStatus(t, err, http.StatusForbidden)
	})
}

func TestProjectService_ReplacePhases(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)

	t.Run("removing a phase that has demand", func(t *testing.T) {
		repo := new(MockProjectRepository)
		repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
		repo.On("CountDemandsInPhases", mock.Anything, []string{testID}).Return(int64(1), nil)
		_, err := newProjectService(repo, nil).ReplacePhases(planner, projectID, []PhaseInput{
			{ID: new(buildID), Name: "Build", StartWeek: mustWeek(t, "2026-W42"), EndWeek: mustWeek(t, "2026-W47")},
		})
		requireAppError(t, err, constants.PhaseInUse, http.StatusConflict)
		repo.AssertNotCalled(t, "SavePhases", mock.Anything, mock.Anything)
	})

	t.Run("keeps, adds and orders phases", func(t *testing.T) {
		repo := new(MockProjectRepository)
		repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
		repo.On("CountDemandsInPhases", mock.Anything, []string(nil)).Return(int64(0), nil)
		repo.On("DeletePhases", mock.Anything, []string(nil)).Return(nil)
		repo.On("SavePhases", mock.Anything, mock.MatchedBy(func(rows []models.ProjectPhase) bool {
			return len(rows) == 3 && rows[0].Name == "Discovery" && rows[0].Sequence == 1 && rows[1].ID == buildID && rows[2].Sequence == 3
		})).Return(nil)
		_, err := newProjectService(repo, nil).ReplacePhases(planner, projectID, []PhaseInput{
			{Name: "Discovery", StartWeek: mustWeek(t, "2026-W40"), EndWeek: mustWeek(t, "2026-W41")},
			{ID: new(buildID), Name: "Build", StartWeek: mustWeek(t, "2026-W42"), EndWeek: mustWeek(t, "2026-W47")},
			{ID: new(testID), Name: "Test", StartWeek: mustWeek(t, "2026-W46"), EndWeek: mustWeek(t, "2026-W51")},
		})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("an id from another project", func(t *testing.T) {
		repo := new(MockProjectRepository)
		repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
		_, err := newProjectService(repo, nil).ReplacePhases(planner, projectID, []PhaseInput{
			{ID: new(demandID), Name: "X", StartWeek: mustWeek(t, "2026-W42"), EndWeek: mustWeek(t, "2026-W43")},
		})
		requireAppError(t, err, constants.PhaseNotInProject, http.StatusBadRequest)
	})
}

func TestProjectService_ReplaceDemands(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)

	t.Run("updates, adds and removes; profile defaults to flat", func(t *testing.T) {
		repo := new(MockProjectRepository)
		repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
		repo.On("CountPastBaselineHours", mock.Anything, []string(nil), mock.Anything, mock.Anything).Return(int64(0), nil)
		repo.On("DeleteDemands", mock.Anything, []string(nil)).Return(nil)
		repo.On("SaveDemands", mock.Anything, mock.MatchedBy(func(rows []models.Demand) bool {
			return len(rows) == 2 && rows[0].ID == demandID && rows[1].PhaseID == nil && rows[1].Profile == constants.DemandProfileFlat
		})).Return(nil)
		_, err := newProjectService(repo, nil).ReplaceDemands(planner, projectID, []DemandInput{
			{ID: new(demandID), ExpertiseID: expQA, PhaseID: new(testID), MinLevel: 3, Headcount: 1, TotalHours: 80, Profile: constants.DemandProfileFlat},
			{ExpertiseID: expQA, MinLevel: 2, Headcount: 1, TotalHours: 40},
		})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("removing a demand that holds past baseline hours", func(t *testing.T) {
		repo := new(MockProjectRepository)
		repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
		repo.On("CountPastBaselineHours", mock.Anything, []string{demandID}, mustWeek(t, "2026-W42"), hoursEpsilon).Return(int64(2), nil)
		_, err := newProjectService(repo, nil).ReplaceDemands(planner, projectID, nil)
		requireAppError(t, err, constants.DemandInUse, http.StatusConflict)
		repo.AssertNotCalled(t, "DeleteDemands", mock.Anything, mock.Anything)
	})

	t.Run("removing a demand with future baseline hours only", func(t *testing.T) {
		repo := new(MockProjectRepository)
		repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
		repo.On("CountPastBaselineHours", mock.Anything, []string{demandID}, mustWeek(t, "2026-W42"), hoursEpsilon).Return(int64(0), nil)
		repo.On("DeleteDemands", mock.Anything, []string{demandID}).Return(nil)
		repo.On("SaveDemands", mock.Anything, mock.Anything).Return(nil)
		_, err := newProjectService(repo, nil).ReplaceDemands(planner, projectID, nil)
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("a phase of another project", func(t *testing.T) {
		repo := new(MockProjectRepository)
		repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
		_, err := newProjectService(repo, nil).ReplaceDemands(planner, projectID, []DemandInput{
			{ExpertiseID: expQA, PhaseID: new(demandID), MinLevel: 3, Headcount: 1, TotalHours: 80},
		})
		requireAppError(t, err, constants.PhaseNotInProject, http.StatusBadRequest)
	})

	t.Run("hours that are not a multiple of the planning step", func(t *testing.T) {
		repo := new(MockProjectRepository)
		repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
		_, err := newProjectService(repo, nil).ReplaceDemands(planner, projectID, []DemandInput{
			{ExpertiseID: expQA, MinLevel: 3, Headcount: 1, TotalHours: 40.5},
		})
		requireAppError(t, err, constants.DemandHoursStep, http.StatusBadRequest)
		repo.AssertNotCalled(t, "SaveDemands", mock.Anything, mock.Anything)
	})

	t.Run("an unknown expertise", func(t *testing.T) {
		repo := new(MockProjectRepository)
		repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
		_, err := newProjectService(repo, nil).ReplaceDemands(planner, projectID, []DemandInput{
			{ExpertiseID: expEngineer, MinLevel: 3, Headcount: 1, TotalHours: 80},
		})
		requireAppError(t, err, constants.ExpertiseUnknown, http.StatusBadRequest)
	})
}

func TestProjectService_Create(t *testing.T) {
	start := time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, time.December, 18, 0, 0, 0, 0, time.UTC)
	ownLead := leadPersonID

	t.Run("the program's lead adds a project and leads it", func(t *testing.T) {
		projects := new(MockProjectRepository)
		programs := new(MockProgramRepository)
		programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, &ownLead), nil)
		projects.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Project) bool {
			return p.Status == constants.ProjectStatusIncoming && *p.LeadID == leadPersonID && p.ProgramID == programID
		})).Return(nil)
		projects.On("FindDetail", mock.Anything, mock.Anything).Return(buildAndTest(t), nil)
		_, err := newProjectService(projects, programs).Create(asRole("kc-lead", constants.RoleLead),
			CreateProjectInput{ProgramID: programID, Name: "App", Priority: new(2), StartEarliest: start, Deadline: deadline})
		require.NoError(t, err)
		projects.AssertExpectations(t)
	})

	t.Run("without a priority the project goes to the bottom of the program's list", func(t *testing.T) {
		projects := new(MockProjectRepository)
		programs := new(MockProgramRepository)
		programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, nil), nil)
		projects.On("MaxPriority", mock.Anything, programID).Return(3, nil)
		projects.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Project) bool { return p.Priority == 4 })).Return(nil)
		projects.On("FindDetail", mock.Anything, mock.Anything).Return(buildAndTest(t), nil)
		_, err := newProjectService(projects, programs).Create(asRole("kc-planner", constants.RolePlanner),
			CreateProjectInput{ProgramID: programID, Name: "App", StartEarliest: start, Deadline: deadline})
		require.NoError(t, err)
		projects.AssertExpectations(t)
	})

	t.Run("a closed program takes no projects", func(t *testing.T) {
		programs := new(MockProgramRepository)
		programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusClosed, nil), nil)
		_, err := newProjectService(new(MockProjectRepository), programs).Create(asRole("kc-planner", constants.RolePlanner),
			CreateProjectInput{ProgramID: programID, Name: "App", Priority: new(2), StartEarliest: start, Deadline: deadline})
		requireAppError(t, err, constants.ProgramNotOpen, http.StatusConflict)
	})
}

func TestProjectService_WeekWindow(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	day := func(year int) time.Time { return time.Date(year, time.October, 12, 0, 0, 0, 0, time.UTC) }
	phases := func(t *testing.T, build, test string) []PhaseInput {
		return []PhaseInput{
			{ID: new(buildID), Name: "Build", StartWeek: mustWeek(t, "2026-W42"), EndWeek: mustWeek(t, build)},
			{ID: new(testID), Name: "Test", StartWeek: mustWeek(t, "2026-W46"), EndWeek: mustWeek(t, test)},
		}
	}
	tests := []struct {
		name     string
		run      func(t *testing.T, repo *MockProjectRepository, programs *MockProgramRepository) error
		wantCode string // "" means success
	}{
		{name: "create inside the window", run: func(t *testing.T, repo *MockProjectRepository, programs *MockProgramRepository) error {
			programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, nil), nil)
			repo.On("MaxPriority", mock.Anything, programID).Return(0, nil)
			repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			repo.On("FindDetail", mock.Anything, mock.Anything).Return(buildAndTest(t), nil)
			_, err := newProjectService(repo, programs).Create(planner, CreateProjectInput{ProgramID: programID, Name: "App", StartEarliest: day(2026), Deadline: day(2028)})
			return err
		}},
		{name: "create with a deadline beyond the window", wantCode: constants.WeekOutOfWindow,
			run: func(t *testing.T, repo *MockProjectRepository, programs *MockProgramRepository) error {
				programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, nil), nil)
				_, err := newProjectService(repo, programs).Create(planner, CreateProjectInput{ProgramID: programID, Name: "App", StartEarliest: day(2026), Deadline: day(9999)})
				return err
			}},
		{name: "create with a start before the window", wantCode: constants.WeekOutOfWindow,
			run: func(t *testing.T, repo *MockProjectRepository, programs *MockProgramRepository) error {
				programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, nil), nil)
				_, err := newProjectService(repo, programs).Create(planner, CreateProjectInput{ProgramID: programID, Name: "App", StartEarliest: day(1), Deadline: day(2026)})
				return err
			}},
		{name: "update with a deadline beyond the window", wantCode: constants.WeekOutOfWindow,
			run: func(t *testing.T, repo *MockProjectRepository, _ *MockProgramRepository) error {
				repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
				deadline := day(9999)
				_, err := newProjectService(repo, nil).Update(planner, projectID, UpdateProjectInput{Deadline: &deadline})
				return err
			}},
		{name: "replace phases with a phase far in the future", wantCode: constants.WeekOutOfWindow,
			run: func(t *testing.T, repo *MockProjectRepository, _ *MockProgramRepository) error {
				repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
				_, err := newProjectService(repo, nil).ReplacePhases(planner, projectID, phases(t, "2026-W47", "9000-W10"))
				return err
			}},
		{name: "replace phases with a phase exactly at the window edge",
			run: func(t *testing.T, repo *MockProjectRepository, _ *MockProgramRepository) error {
				repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
				repo.On("CountDemandsInPhases", mock.Anything, mock.Anything).Return(int64(0), nil)
				repo.On("DeletePhases", mock.Anything, mock.Anything).Return(nil)
				repo.On("SavePhases", mock.Anything, mock.Anything).Return(nil)
				edge := isoweek.Of(acceptedAt).AddWeeks(projectConfig().ProjectWindowWeeks)
				_, err := newProjectService(repo, nil).ReplacePhases(planner, projectID, []PhaseInput{{Name: "Edge", StartWeek: edge, EndWeek: edge}})
				return err
			}},
		{name: "replace phases one week past the window edge", wantCode: constants.WeekOutOfWindow,
			run: func(t *testing.T, repo *MockProjectRepository, _ *MockProgramRepository) error {
				repo.On("FindDetail", mock.Anything, projectID).Return(buildAndTest(t), nil)
				past := isoweek.Of(acceptedAt).AddWeeks(projectConfig().ProjectWindowWeeks + 1)
				_, err := newProjectService(repo, nil).ReplacePhases(planner, projectID, []PhaseInput{{Name: "Past", StartWeek: past, EndWeek: past}})
				return err
			}},
		{name: "renaming a stored phase outside the window",
			run: func(t *testing.T, repo *MockProjectRepository, _ *MockProgramRepository) error {
				project := buildAndTest(t)
				project.Phases[0].StartWeek, project.Phases[0].EndWeek = mustWeek(t, "2010-W10"), mustWeek(t, "2010-W12")
				build := project.Phases[0]
				repo.On("FindPhase", mock.Anything, buildID).Return(&build, nil)
				repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
				repo.On("SavePhases", mock.Anything, mock.Anything).Return(nil)
				_, err := newProjectService(repo, nil).UpdatePhase(planner, buildID, UpdatePhaseInput{Name: new("Renamed")})
				return err
			}},
		{name: "replace phases keeping a stored phase outside the window",
			run: func(t *testing.T, repo *MockProjectRepository, _ *MockProgramRepository) error {
				project := buildAndTest(t)
				project.Phases[0].StartWeek, project.Phases[0].EndWeek = mustWeek(t, "2010-W10"), mustWeek(t, "2010-W12")
				repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
				repo.On("CountDemandsInPhases", mock.Anything, mock.Anything).Return(int64(0), nil)
				repo.On("DeletePhases", mock.Anything, mock.Anything).Return(nil)
				repo.On("SavePhases", mock.Anything, mock.Anything).Return(nil)
				_, err := newProjectService(repo, nil).ReplacePhases(planner, projectID, []PhaseInput{
					{ID: new(buildID), Name: "Renamed", StartWeek: mustWeek(t, "2010-W10"), EndWeek: mustWeek(t, "2010-W12")}})
				return err
			}},
		{name: "update a phase end beyond the window", wantCode: constants.WeekOutOfWindow,
			run: func(t *testing.T, repo *MockProjectRepository, _ *MockProgramRepository) error {
				project := buildAndTest(t)
				build := project.Phases[0]
				repo.On("FindPhase", mock.Anything, buildID).Return(&build, nil)
				repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
				_, err := newProjectService(repo, nil).UpdatePhase(planner, buildID, UpdatePhaseInput{EndWeek: new(mustWeek(t, "9000-W10"))})
				return err
			}},
		{name: "a slip that shifts a later phase outside the window", wantCode: constants.WeekOutOfWindow,
			run: func(t *testing.T, repo *MockProjectRepository, _ *MockProgramRepository) error {
				project := buildAndTest(t)
				project.Phases[1].EndWeek = mustWeek(t, "2029-W40") // inside, but the next slip pushes it out
				build := project.Phases[0]
				repo.On("FindPhase", mock.Anything, buildID).Return(&build, nil)
				repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
				_, err := newProjectService(repo, nil).UpdatePhase(planner, buildID, UpdatePhaseInput{EndWeek: new(mustWeek(t, "2026-W50"))})
				return err
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockProjectRepository)
			err := tt.run(t, repo, new(MockProgramRepository))
			if tt.wantCode == "" {
				require.NoError(t, err)
				return
			}
			requireAppError(t, err, tt.wantCode, http.StatusBadRequest)
			repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			repo.AssertNotCalled(t, "SavePhases", mock.Anything, mock.Anything)
		})
	}
}

func TestProjectService_ApplyAction(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		action     string
		wantStatus string
		wantErr    int
	}{
		{name: "hold an incoming project", status: constants.ProjectStatusIncoming, action: constants.ProjectActionHold, wantStatus: constants.ProjectStatusOnHold},
		{name: "resume goes back to incoming", status: constants.ProjectStatusOnHold, action: constants.ProjectActionResume, wantStatus: constants.ProjectStatusIncoming},
		{name: "cancel an on-hold project", status: constants.ProjectStatusOnHold, action: constants.ProjectActionCancel, wantStatus: constants.ProjectStatusCancelled},
		{name: "done is final", status: constants.ProjectStatusDone, action: constants.ProjectActionHold, wantErr: http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockProjectRepository)
			project := buildAndTest(t)
			project.Status = tt.status
			repo.On("FindDetail", mock.Anything, projectID).Return(project, nil)
			if tt.wantErr == 0 {
				repo.On("Save", mock.Anything, mock.MatchedBy(func(p *models.Project) bool {
					return p.Status == tt.wantStatus && p.Phases == nil && p.Program == nil
				})).Return(nil)
			}
			_, err := newProjectService(repo, nil).ApplyAction(asRole("kc-planner", constants.RolePlanner), projectID, tt.action)
			if tt.wantErr != 0 {
				requireStatus(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestProgramService_CancelCascades(t *testing.T) {
	tests := []struct {
		name       string
		cancelled  []string
		wantEvents int
	}{
		{name: "records one event for the projects it cancelled", cancelled: []string{projectID, "0192f3a4-0000-7000-8000-0000000000e2"}, wantEvents: 1},
		{name: "records nothing when no project was open", cancelled: nil, wantEvents: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			programs, projects, events := new(MockProgramRepository), new(MockProjectRepository), &recordedEvents{}
			programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, nil), nil)
			programs.On("Save", mock.Anything, mock.Anything).Return(nil)
			projects.On("CancelOpen", mock.Anything, programID).Return(tt.cancelled, nil)

			people := leadPeople{}
			service := ProvideProgramService(programs, projects, people, ProvideAuthzService(people), passthroughTx{}, events)
			_, err := service.ApplyAction(asRole("kc-planner", constants.RolePlanner), programID, constants.ProgramActionCancel)
			require.NoError(t, err)
			projects.AssertExpectations(t)
			require.Len(t, events.events, tt.wantEvents)
			if tt.wantEvents == 1 {
				assert.Equal(t, constants.EventProjectStatusChanged, events.events[0].Type)
				assert.Equal(t, tt.cancelled, events.events[0].ProjectIDs)
				assert.Equal(t, constants.StatusKindProgram, events.events[0].Payload.(StatusChange).Kind)
			}
		})
	}

	t.Run("a program that is not cancelled records nothing", func(t *testing.T) {
		programs, projects, events := new(MockProgramRepository), new(MockProjectRepository), &recordedEvents{}
		programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, nil), nil)
		programs.On("Save", mock.Anything, mock.Anything).Return(nil)
		people := leadPeople{}
		service := ProvideProgramService(programs, projects, people, ProvideAuthzService(people), passthroughTx{}, events)
		_, err := service.ApplyAction(asRole("kc-planner", constants.RolePlanner), programID, constants.ProgramActionHold)
		require.NoError(t, err)
		assert.Empty(t, events.events)
	})
}

func (m *MockProjectRepository) ListByIDs(ctx context.Context, ids []string) ([]models.Project, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.Project), args.Error(1)
}

func TestProjectService_Reorder(t *testing.T) {
	const (
		first  = "0192f3a4-0000-7000-8000-0000000000f1"
		second = "0192f3a4-0000-7000-8000-0000000000f2"
		third  = "0192f3a4-0000-7000-8000-0000000000f3"
		other  = "0192f3a4-0000-7000-8000-0000000000f9"
	)
	ownLead, otherLead := leadPersonID, other
	listed := []models.Project{
		{BaseModel: models.BaseModel{ID: first}, ProgramID: programID, Priority: 1},
		{BaseModel: models.BaseModel{ID: second}, ProgramID: programID, Priority: 2},
		{BaseModel: models.BaseModel{ID: third}, ProgramID: programID, Priority: 2},
	}
	setup := func(lead *string) (*MockProjectRepository, *MockProgramRepository) {
		projects, programs := new(MockProjectRepository), new(MockProgramRepository)
		programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, lead), nil)
		projects.On("ListByProgram", mock.Anything, programID).Return(listed, nil)
		return projects, programs
	}

	t.Run("the program's manager sets the order; priorities follow it from 1", func(t *testing.T) {
		projects, programs := setup(&ownLead)
		projects.On("SetPriorities", mock.Anything, map[string]int{third: 1, first: 2, second: 3}).Return(nil)
		_, err := newProjectService(projects, programs).Reorder(asRole("kc-lead", constants.RoleLead), programID, []string{third, first, second})
		require.NoError(t, err)
		projects.AssertExpectations(t)
	})

	t.Run("the order must list each of the program's projects once", func(t *testing.T) {
		for name, ids := range map[string][]string{
			"one missing":   {first, second},
			"one repeated":  {first, second, second},
			"one not there": {first, second, other},
		} {
			t.Run(name, func(t *testing.T) {
				projects, programs := setup(nil)
				_, err := newProjectService(projects, programs).Reorder(asRole("kc-planner", constants.RolePlanner), programID, ids)
				requireAppError(t, err, constants.ProjectOrderMismatch, http.StatusUnprocessableEntity)
				projects.AssertNotCalled(t, "SetPriorities", mock.Anything, mock.Anything)
			})
		}
	})

	t.Run("another program's manager may not", func(t *testing.T) {
		projects, programs := setup(&otherLead)
		_, err := newProjectService(projects, programs).Reorder(asRole("kc-lead", constants.RoleLead), programID, []string{first, second, third})
		requireStatus(t, err, http.StatusForbidden)
	})
}
