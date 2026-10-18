package services

import (
	"context"
	"net/http"
	"testing"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	ownProgramID     = "0192f3a4-0000-7000-8000-000000000091"
	projectProgramID = "0192f3a4-0000-7000-8000-000000000092"
	foreignProgramID = "0192f3a4-0000-7000-8000-000000000093"
)

// ledProjects says, per program, who leads one of its projects.
type ledProjects map[string][]string

func (l ledProjects) LeadsProjectIn(_ context.Context, programID, personID string) (bool, error) {
	for _, id := range l[programID] {
		if id == personID {
			return true, nil
		}
	}
	return false, nil
}

func (ledProjects) OnBaselineIn(context.Context, string, string) (bool, error) { return false, nil }

// workedPrograms adds, per program, who has baseline hours on one of its projects.
type workedPrograms struct {
	ledProjects
	worked ledProjects
}

func (w workedPrograms) OnBaselineIn(ctx context.Context, programID, personID string) (bool, error) {
	return w.worked.LeadsProjectIn(ctx, programID, personID)
}

func ledProgram(id, leadID string) *models.Program {
	return &models.Program{BaseModel: models.BaseModel{ID: id}, LeadID: &leadID}
}

func TestAuthzService_ProgramReader(t *testing.T) {
	authz := newVisibilityAuthz()

	planner, err := authz.ProgramReader(asRole("kc-planner", constants.RolePlanner))
	require.NoError(t, err)
	assert.True(t, planner.All)

	lead, err := authz.ProgramReader(asRole("kc-manager", constants.RoleManager))
	require.NoError(t, err)
	assert.Equal(t, ProgramReader{LeadID: managerPersonID}, lead)

	unknown, err := authz.ProgramReader(asRole("kc-unknown-manager", constants.RoleManager))
	require.NoError(t, err)
	assert.Equal(t, ProgramReader{}, unknown, "a manager without a person record reads no program")

	specialist, err := authz.ProgramReader(asRole("kc-specialist", constants.RoleSpecialist))
	require.NoError(t, err)
	assert.Equal(t, ProgramReader{LeadID: selfPersonID, OnBaseline: true}, specialist, "a specialist also reads where they have baseline hours")

	none, err := authz.ProgramReader(asRole("kc-unknown-specialist", constants.RoleSpecialist))
	require.NoError(t, err)
	assert.Equal(t, ProgramReader{}, none, "a specialist without a person record reads no program")

	_, err = authz.ProgramReader(asRole("kc-nobody", "offline_access"))
	requireStatus(t, err, http.StatusForbidden)
}

func TestRequireProgramReader(t *testing.T) {
	projects := workedPrograms{ledProjects: ledProjects{projectProgramID: {managerPersonID}}, worked: ledProjects{ownProgramID: {selfPersonID}}}
	tests := []struct {
		name    string
		ctx     context.Context
		program *models.Program
		allowed bool
	}{
		{name: "a planner reads any program", ctx: asRole("kc-planner", constants.RolePlanner),
			program: ledProgram(foreignProgramID, outsiderPersonID), allowed: true},
		{name: "a manager reads a program they lead", ctx: asRole("kc-manager", constants.RoleManager),
			program: ledProgram(ownProgramID, managerPersonID), allowed: true},
		{name: "a manager reads a program where they lead a project", ctx: asRole("kc-manager", constants.RoleManager),
			program: ledProgram(projectProgramID, outsiderPersonID), allowed: true},
		{name: "a manager cannot read another manager's program", ctx: asRole("kc-manager", constants.RoleManager),
			program: ledProgram(foreignProgramID, outsiderPersonID)},
		{name: "a manager without a person record reads none", ctx: asRole("kc-unknown-manager", constants.RoleManager),
			program: ledProgram(ownProgramID, managerPersonID)},
		{name: "a specialist reads a program where they have baseline hours", ctx: asRole("kc-specialist", constants.RoleSpecialist),
			program: ledProgram(ownProgramID, outsiderPersonID), allowed: true},
		{name: "a specialist cannot read a program where they have none", ctx: asRole("kc-specialist", constants.RoleSpecialist),
			program: ledProgram(foreignProgramID, outsiderPersonID)},
		{name: "a manager does not read a program by baseline hours", ctx: asRole("kc-manager", constants.RoleManager),
			program: ledProgram(ownProgramID, outsiderPersonID)},
		{name: "a program without a lead is the planner's", ctx: asRole("kc-manager", constants.RoleManager),
			program: &models.Program{BaseModel: models.BaseModel{ID: foreignProgramID}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireProgramReader(tt.ctx, newVisibilityAuthz(), projects, tt.program, opGetProgram)
			if tt.allowed {
				require.NoError(t, err)
				return
			}
			requireStatus(t, err, http.StatusForbidden)
		})
	}
}

func TestEmptyPage(t *testing.T) {
	page := emptyPage[models.Program](&dtos.PageableRequest{Page: 2, PageSize: 20})
	assert.Empty(t, page.Data)
	assert.NotNil(t, page.Data, "an empty page is sent as []")
	assert.Equal(t, 2, page.Pageable.Page)
	assert.Equal(t, 20, page.Pageable.PageSize)
}

// readerIs matches the reader id a list is limited to; nil matches a list of everything.
func readerIs(personID *string) any {
	return mock.MatchedBy(func(id *string) bool {
		if personID == nil || id == nil {
			return personID == nil && id == nil
		}
		return *id == *personID
	})
}

func TestProgramService_List_ByReader(t *testing.T) {
	lead := leadPersonID
	tests := []struct {
		name     string
		ctx      context.Context
		readerID *string
		listed   bool
	}{
		{name: "a planner lists every program", ctx: asRole("kc-planner", constants.RolePlanner), listed: true},
		{name: "a manager lists the programs they lead", ctx: asRole("kc-manager", constants.RoleManager), readerID: &lead, listed: true},
		{name: "a manager without a person record lists none", ctx: asRole("kc-unknown-manager", constants.RoleManager)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockProgramRepository)
			if tt.listed {
				repo.On("List", mock.Anything, mock.Anything, readerIs(tt.readerID), false).
					Return(&dtos.DataResponse[models.Program]{Data: []models.Program{{Name: "Acme"}}}, nil)
			}
			page, err := newProgramService(repo).List(tt.ctx, ListProgramsInput{Page: 1, PageSize: 20})
			require.NoError(t, err)
			assert.Equal(t, tt.listed, len(page.Data) == 1)
			repo.AssertExpectations(t)
		})
	}
}

func TestProgramService_List_Specialist(t *testing.T) {
	self := selfPersonID
	repo := new(MockProgramRepository)
	repo.On("List", mock.Anything, mock.Anything, readerIs(&self), true).
		Return(&dtos.DataResponse[models.Program]{Data: []models.Program{{Name: "Acme"}}}, nil)
	service := ProvideProgramService(repo, nil, leadPeople{}, ProvideAuthzService(selfFinder{"kc-specialist": selfPersonID}),
		passthroughTx{}, &recordedEvents{})

	page, err := service.List(asRole("kc-specialist", constants.RoleSpecialist), ListProgramsInput{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, page.Data, 1)
	repo.AssertExpectations(t)

	page, err = service.List(asRole("kc-unknown-specialist", constants.RoleSpecialist), ListProgramsInput{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Empty(t, page.Data, "a specialist without a person record lists none")
}

func TestProjectService_Specialist_ReadsWithoutWork(t *testing.T) {
	other := otherPersonID2
	specialist := asRole("kc-specialist", constants.RoleSpecialist)
	newService := func(projects *MockProjectRepository, programs *MockProgramRepository) ProjectService {
		people := knownExpertisePeople{}
		return ProvideProjectService(projects, programs, people, ProvideAuthzService(selfFinder{"kc-specialist": selfPersonID}),
			passthroughTx{}, &recordedEvents{}, projectConfig(), clock.NewFixed(acceptedAt))
	}
	detail := func() *models.Project {
		return &models.Project{BaseModel: models.BaseModel{ID: projectID}, ProgramID: programID, ScopePoints: 40, DonePoints: 10,
			Program: &models.Program{BaseModel: models.BaseModel{ID: programID}, LeadID: &other},
			Demands: []models.Demand{{TotalHours: 120}}}
	}

	t.Run("project detail has no demand or points", func(t *testing.T) {
		projects := new(MockProjectRepository)
		projects.On("FindDetail", mock.Anything, projectID).Return(detail(), nil)
		projects.On("LeadsProjectIn", mock.Anything, programID, selfPersonID).Return(false, nil)
		projects.On("OnBaselineIn", mock.Anything, programID, selfPersonID).Return(true, nil)
		projects.On("OnBaselineProjects", mock.Anything, programID, selfPersonID).Return([]string{projectID}, nil)
		project, err := newService(projects, nil).Get(specialist, projectID)
		require.NoError(t, err)
		assert.Empty(t, project.Demands)
		assert.Zero(t, project.ScopePoints+project.DonePoints)
	})

	t.Run("a manager still gets the demand", func(t *testing.T) {
		projects := new(MockProjectRepository)
		projects.On("FindDetail", mock.Anything, projectID).Return(detail(), nil)
		projects.On("LeadsProjectIn", mock.Anything, programID, leadPersonID).Return(true, nil)
		project, err := newProjectService(projects, nil).Get(asRole("kc-manager", constants.RoleManager), projectID)
		require.NoError(t, err)
		assert.Len(t, project.Demands, 1)
		assert.Equal(t, 40.0, project.ScopePoints)
	})

	t.Run("project list has no points", func(t *testing.T) {
		projects, programs := new(MockProjectRepository), new(MockProgramRepository)
		programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, &other), nil)
		projects.On("LeadsProjectIn", mock.Anything, programID, selfPersonID).Return(false, nil)
		projects.On("OnBaselineIn", mock.Anything, programID, selfPersonID).Return(true, nil)
		projects.On("OnBaselineProjects", mock.Anything, programID, selfPersonID).Return([]string{projectID}, nil)
		projects.On("ListByProgram", mock.Anything, programID).Return([]models.Project{*detail()}, nil)
		list, err := newService(projects, programs).ListByProgram(specialist, programID)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Zero(t, list[0].ScopePoints+list[0].DonePoints)
	})

	t.Run("project list keeps only the projects they work on or lead", func(t *testing.T) {
		const worked, led, notMine = "p-worked", "p-led", "p-not-mine"
		self := selfPersonID
		projects, programs := new(MockProjectRepository), new(MockProgramRepository)
		programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, &other), nil)
		projects.On("LeadsProjectIn", mock.Anything, programID, selfPersonID).Return(true, nil)
		projects.On("OnBaselineProjects", mock.Anything, programID, selfPersonID).Return([]string{worked}, nil)
		projects.On("ListByProgram", mock.Anything, programID).Return([]models.Project{
			{BaseModel: models.BaseModel{ID: worked}}, {BaseModel: models.BaseModel{ID: notMine}}, {BaseModel: models.BaseModel{ID: led}, LeadID: &self}}, nil)
		list, err := newService(projects, programs).ListByProgram(specialist, programID)
		require.NoError(t, err)
		require.Len(t, list, 2)
		assert.Equal(t, []string{worked, led}, []string{list[0].ID, list[1].ID})
	})

	t.Run("the program lead sees every project", func(t *testing.T) {
		self := selfPersonID
		projects, programs := new(MockProjectRepository), new(MockProgramRepository)
		programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, &self), nil)
		projects.On("ListByProgram", mock.Anything, programID).Return([]models.Project{{BaseModel: models.BaseModel{ID: "a"}}, {BaseModel: models.BaseModel{ID: "b"}}}, nil)
		list, err := newService(projects, programs).ListByProgram(specialist, programID)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("a project they lead is readable without hours", func(t *testing.T) {
		self := selfPersonID
		d := detail()
		d.LeadID = &self
		projects := new(MockProjectRepository)
		projects.On("FindDetail", mock.Anything, projectID).Return(d, nil)
		projects.On("LeadsProjectIn", mock.Anything, programID, selfPersonID).Return(true, nil)
		projects.On("OnBaselineProjects", mock.Anything, programID, selfPersonID).Return(nil, nil)
		_, err := newService(projects, nil).Get(specialist, projectID)
		require.NoError(t, err)
	})

	t.Run("another project of a readable program gets 403, for the project and its phases", func(t *testing.T) {
		projects := new(MockProjectRepository)
		projects.On("FindDetail", mock.Anything, projectID).Return(detail(), nil)
		projects.On("LeadsProjectIn", mock.Anything, programID, selfPersonID).Return(false, nil)
		projects.On("OnBaselineIn", mock.Anything, programID, selfPersonID).Return(true, nil)
		projects.On("OnBaselineProjects", mock.Anything, programID, selfPersonID).Return([]string{"another-project"}, nil)
		// The phases route reads through Get.
		_, err := newService(projects, nil).Get(specialist, projectID)
		requireStatus(t, err, http.StatusForbidden)
	})

	t.Run("not on the baseline gets 403", func(t *testing.T) {
		projects := new(MockProjectRepository)
		projects.On("FindDetail", mock.Anything, projectID).Return(detail(), nil)
		projects.On("LeadsProjectIn", mock.Anything, programID, selfPersonID).Return(false, nil)
		projects.On("OnBaselineIn", mock.Anything, programID, selfPersonID).Return(false, nil)
		_, err := newService(projects, nil).Get(specialist, projectID)
		requireStatus(t, err, http.StatusForbidden)
	})
}

func TestInsightsService_ProjectMetrics_NotForSpecialist(t *testing.T) {
	authz := ProvideAuthzService(selfFinder{"kc-specialist": selfPersonID})
	service, _ := newMetricsServiceWith(t, map[string]float64{}, authz)
	_, err := service.ProjectMetrics(asRole("kc-specialist", constants.RoleSpecialist), metricsProgramID)
	requireStatus(t, err, http.StatusForbidden)
}

func TestProgramService_Get_ByReader(t *testing.T) {
	lead := asRole("kc-manager", constants.RoleManager)
	own, other := leadPersonID, otherPersonID2

	repo, projects := new(MockProgramRepository), new(MockProjectRepository)
	repo.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, &other), nil)
	projects.On("LeadsProjectIn", mock.Anything, programID, leadPersonID).Return(false, nil)
	people := leadPeople{}
	_, err := ProvideProgramService(repo, projects, people, ProvideAuthzService(people), passthroughTx{}, &recordedEvents{}).Get(lead, programID)
	requireStatus(t, err, http.StatusForbidden)

	repo = new(MockProgramRepository)
	repo.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, &own), nil)
	_, err = newProgramService(repo).Get(lead, programID)
	require.NoError(t, err)
}

func TestProjectService_ListByStatus_ByReader(t *testing.T) {
	lead := leadPersonID
	for _, tt := range []struct {
		name     string
		ctx      context.Context
		readerID *string
	}{
		{name: "a planner sees every program's projects", ctx: asRole("kc-planner", constants.RolePlanner)},
		{name: "a manager sees the projects of their programs", ctx: asRole("kc-manager", constants.RoleManager), readerID: &lead},
	} {
		t.Run(tt.name, func(t *testing.T) {
			projects := new(MockProjectRepository)
			projects.On("ListByStatus", mock.Anything, constants.ProjectStatusIncoming, readerIs(tt.readerID), mock.Anything).
				Return(&dtos.DataResponse[models.Project]{Data: []models.Project{}}, nil)
			_, err := newProjectService(projects, nil).ListByStatus(tt.ctx,
				ListProjectsInput{Status: constants.ProjectStatusIncoming, Page: 1, PageSize: 20})
			require.NoError(t, err)
			projects.AssertExpectations(t)
		})
	}
}

func TestProjectService_ListByStatus_NotForSpecialist(t *testing.T) {
	projects := new(MockProjectRepository)
	people := knownExpertisePeople{}
	service := ProvideProjectService(projects, nil, people, ProvideAuthzService(selfFinder{"kc-specialist": selfPersonID}),
		passthroughTx{}, &recordedEvents{}, projectConfig(), clock.NewFixed(acceptedAt))
	_, err := service.ListByStatus(asRole("kc-specialist", constants.RoleSpecialist), ListProjectsInput{Page: 1, PageSize: 20})
	requireStatus(t, err, http.StatusForbidden)
	projects.AssertNotCalled(t, "ListByStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestProjectService_Get_ByReader(t *testing.T) {
	lead := asRole("kc-manager", constants.RoleManager)
	other := otherPersonID2
	detail := &models.Project{BaseModel: models.BaseModel{ID: projectID}, ProgramID: programID,
		Program: &models.Program{BaseModel: models.BaseModel{ID: programID}, LeadID: &other}}

	for _, tt := range []struct {
		name    string
		leads   bool
		allowed bool
	}{
		{name: "a manager of another program cannot open it"},
		{name: "the lead of one of the program's projects can", leads: true, allowed: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			projects := new(MockProjectRepository)
			projects.On("FindDetail", mock.Anything, projectID).Return(detail, nil)
			projects.On("LeadsProjectIn", mock.Anything, programID, leadPersonID).Return(tt.leads, nil)
			_, err := newProjectService(projects, nil).Get(lead, projectID)
			if tt.allowed {
				require.NoError(t, err)
				return
			}
			requireStatus(t, err, http.StatusForbidden)
		})
	}
}

func TestProjectService_ListByProgram_ByReader(t *testing.T) {
	other := otherPersonID2
	projects, programs := new(MockProjectRepository), new(MockProgramRepository)
	programs.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, &other), nil)
	projects.On("LeadsProjectIn", mock.Anything, programID, leadPersonID).Return(false, nil)
	_, err := newProjectService(projects, programs).ListByProgram(asRole("kc-manager", constants.RoleManager), programID)
	requireStatus(t, err, http.StatusForbidden)
	projects.AssertNotCalled(t, "ListByProgram", mock.Anything, mock.Anything)
}

func TestInsightsService_ByReader(t *testing.T) {
	authz := ProvideAuthzService(selfFinder{"kc-manager": leadPersonID, "kc-other-manager": otherPersonID2})
	service, projects := newMetricsServiceWith(t, map[string]float64{}, authz)
	projects.On("LeadsProjectIn", mock.Anything, metricsProgramID, otherPersonID2).Return(false, nil)
	other := asRole("kc-other-manager", constants.RoleManager)

	own, err := service.Risks(asRole("kc-manager", constants.RoleManager))
	require.NoError(t, err)
	assert.NotEmpty(t, own.Projects, "the manager of Acme sees its project risks")

	risks, err := service.Risks(other)
	require.NoError(t, err)
	assert.Empty(t, risks.Projects, "another program's manager does not")

	_, err = service.ProgramOverview(other, metricsProgramID)
	requireStatus(t, err, http.StatusForbidden)
	_, err = service.ProgramMembers(other, metricsProgramID)
	requireStatus(t, err, http.StatusForbidden)
}
