package services

import (
	"context"
	"net/http"
	"testing"

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

func ledProgram(id, leadID string) *models.Program {
	return &models.Program{BaseModel: models.BaseModel{ID: id}, LeadID: &leadID}
}

func TestAuthzService_ProgramReader(t *testing.T) {
	authz := newVisibilityAuthz()

	planner, err := authz.ProgramReader(asRole("kc-planner", constants.RolePlanner))
	require.NoError(t, err)
	assert.True(t, planner.All)

	lead, err := authz.ProgramReader(asRole("kc-lead", constants.RoleLead))
	require.NoError(t, err)
	assert.Equal(t, ProgramReader{LeadID: managerPersonID}, lead)

	unknown, err := authz.ProgramReader(asRole("kc-unknown-lead", constants.RoleLead))
	require.NoError(t, err)
	assert.Equal(t, ProgramReader{}, unknown, "a manager without a person record reads no program")

	_, err = authz.ProgramReader(asRole("kc-specialist", constants.RoleSpecialist))
	requireStatus(t, err, http.StatusForbidden)
}

func TestRequireProgramReader(t *testing.T) {
	projects := ledProjects{projectProgramID: {managerPersonID}}
	tests := []struct {
		name    string
		ctx     context.Context
		program *models.Program
		allowed bool
	}{
		{name: "a planner reads any program", ctx: asRole("kc-planner", constants.RolePlanner),
			program: ledProgram(foreignProgramID, outsiderPersonID), allowed: true},
		{name: "a manager reads a program they lead", ctx: asRole("kc-lead", constants.RoleLead),
			program: ledProgram(ownProgramID, managerPersonID), allowed: true},
		{name: "a manager reads a program where they lead a project", ctx: asRole("kc-lead", constants.RoleLead),
			program: ledProgram(projectProgramID, outsiderPersonID), allowed: true},
		{name: "a manager cannot read another manager's program", ctx: asRole("kc-lead", constants.RoleLead),
			program: ledProgram(foreignProgramID, outsiderPersonID)},
		{name: "a manager without a person record reads none", ctx: asRole("kc-unknown-lead", constants.RoleLead),
			program: ledProgram(ownProgramID, managerPersonID)},
		{name: "a program without a lead is the planner's", ctx: asRole("kc-lead", constants.RoleLead),
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
		{name: "a manager lists the programs they lead", ctx: asRole("kc-lead", constants.RoleLead), readerID: &lead, listed: true},
		{name: "a manager without a person record lists none", ctx: asRole("kc-unknown-lead", constants.RoleLead)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockProgramRepository)
			if tt.listed {
				repo.On("List", mock.Anything, mock.Anything, readerIs(tt.readerID)).
					Return(&dtos.DataResponse[models.Program]{Data: []models.Program{{Name: "Acme"}}}, nil)
			}
			page, err := newProgramService(repo).List(tt.ctx, ListProgramsInput{Page: 1, PageSize: 20})
			require.NoError(t, err)
			assert.Equal(t, tt.listed, len(page.Data) == 1)
			repo.AssertExpectations(t)
		})
	}
}

func TestProgramService_Get_ByReader(t *testing.T) {
	lead := asRole("kc-lead", constants.RoleLead)
	own, other := leadPersonID, otherPersonID2

	repo, projects := new(MockProgramRepository), new(MockProjectRepository)
	repo.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, &other), nil)
	projects.On("LeadsProjectIn", mock.Anything, programID, leadPersonID).Return(false, nil)
	people := leadPeople{}
	_, err := ProvideProgramService(repo, projects, people, ProvideAuthzService(people), passthroughTx{}).Get(lead, programID)
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
		{name: "a manager sees the projects of their programs", ctx: asRole("kc-lead", constants.RoleLead), readerID: &lead},
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

func TestProjectService_Get_ByReader(t *testing.T) {
	lead := asRole("kc-lead", constants.RoleLead)
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
	_, err := newProjectService(projects, programs).ListByProgram(asRole("kc-lead", constants.RoleLead), programID)
	requireStatus(t, err, http.StatusForbidden)
	projects.AssertNotCalled(t, "ListByProgram", mock.Anything, mock.Anything)
}

func TestInsightsService_ByReader(t *testing.T) {
	authz := ProvideAuthzService(selfFinder{"kc-lead": leadPersonID, "kc-other-lead": otherPersonID2})
	service, projects := newMetricsServiceWith(t, map[string]float64{}, authz)
	projects.On("LeadsProjectIn", mock.Anything, metricsProgramID, otherPersonID2).Return(false, nil)
	other := asRole("kc-other-lead", constants.RoleLead)

	own, err := service.Risks(asRole("kc-lead", constants.RoleLead))
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
