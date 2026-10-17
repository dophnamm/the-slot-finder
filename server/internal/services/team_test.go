package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"slices"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	teamProgramID = "0192f3a4-0000-7000-8000-0000000000e1"
	teamPersonID  = "0192f3a4-0000-7000-8000-0000000000e2"
)

// memoryTeams is an in-memory program team repository.
type memoryTeams struct {
	members map[string][]string
}

func (m *memoryTeams) ListByProgram(_ context.Context, programID string) ([]models.ProgramTeamMember, error) {
	out := []models.ProgramTeamMember{}
	for _, id := range m.members[programID] {
		out = append(out, models.ProgramTeamMember{ProgramID: programID, PersonID: id,
			Person: &models.Person{BaseModel: models.BaseModel{ID: id}, Name: id}})
	}
	return out, nil
}

func (m *memoryTeams) Add(_ context.Context, member *models.ProgramTeamMember) error {
	if slices.Contains(m.members[member.ProgramID], member.PersonID) {
		return gorm.ErrDuplicatedKey
	}
	m.members[member.ProgramID] = append(m.members[member.ProgramID], member.PersonID)
	return nil
}

func (m *memoryTeams) Remove(_ context.Context, programID, personID string) (bool, error) {
	before := len(m.members[programID])
	m.members[programID] = slices.DeleteFunc(m.members[programID], func(id string) bool { return id == personID })
	return len(m.members[programID]) < before, nil
}

func (m *memoryTeams) PersonIDsByProgram(_ context.Context, programIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	for _, id := range programIDs {
		out[id] = append([]string{}, m.members[id]...)
	}
	return out, nil
}

func (m *memoryTeams) PersonIDsLedBy(_ context.Context, _ string) ([]string, error) { return nil, nil }

func newTeamService(t *testing.T, teams *memoryTeams) ProgramTeamService {
	t.Helper()
	return newTeamServiceWith(t, teams, new(MockProjectRepository), ProvideAuthzService(leadPeople{}))
}

// newTeamServiceWith is newTeamService with this project repository and authz service.
func newTeamServiceWith(t *testing.T, teams *memoryTeams, projects *MockProjectRepository, authz AuthzService) ProgramTeamService {
	t.Helper()
	return newTeamServiceRecording(t, teams, projects, authz, &recordedEvents{})
}

// newTeamServiceRecording is newTeamServiceWith with these events to record into.
func newTeamServiceRecording(t *testing.T, teams *memoryTeams, projects *MockProjectRepository, authz AuthzService,
	events *recordedEvents,
) ProgramTeamService {
	t.Helper()
	programs := new(MockProgramRepository)
	lead := leadPersonID
	programs.On("FindWithLead", mock.Anything, teamProgramID).Return(&models.Program{BaseModel: models.BaseModel{ID: teamProgramID}, LeadID: &lead}, nil)
	people := new(MockPersonRepository)
	people.On("FindWithSkills", mock.Anything, teamPersonID).Return(&models.Person{BaseModel: models.BaseModel{ID: teamPersonID}}, nil).Maybe()
	plans := new(MockPlanRepository)
	plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(1), nil).Maybe()
	return ProvideProgramTeamService(teams, programs, projects, people, plans, authz, passthroughTx{}, events)
}

func TestProgramTeamService_List(t *testing.T) {
	teams := &memoryTeams{members: map[string][]string{teamProgramID: {teamPersonID}}}
	service := newTeamService(t, teams)

	members, err := service.List(asRole("kc-manager", constants.RoleManager), teamProgramID)
	require.NoError(t, err, "the program's manager reads its team")
	require.Len(t, members, 1)
	assert.Equal(t, teamPersonID, members[0].Person.ID)
	assert.Nil(t, members[0].Hours, "no hours in the baseline yet")

	_, err = service.List(asRole("kc-other-manager", constants.RoleManager), teamProgramID)
	requireStatus(t, err, http.StatusForbidden)
}

func TestProgramTeamService_List_ProjectLead(t *testing.T) {
	teams := &memoryTeams{members: map[string][]string{teamProgramID: {teamPersonID}}}
	projects := new(MockProjectRepository)
	projects.On("LeadsProjectIn", mock.Anything, teamProgramID, otherPersonID2).Return(true, nil)
	authz := ProvideAuthzService(selfFinder{"kc-project-manager": otherPersonID2})

	members, err := newTeamServiceWith(t, teams, projects, authz).List(asRole("kc-project-manager", constants.RoleManager), teamProgramID)
	require.NoError(t, err, "the lead of one of the program's projects reads its team")
	require.Len(t, members, 1)
}

func TestProgramTeamService_AddAndRemove(t *testing.T) {
	teams := &memoryTeams{members: map[string][]string{}}
	projects, events := new(MockProjectRepository), &recordedEvents{}
	projects.On("ListByProgram", mock.Anything, teamProgramID).Return([]models.Project{
		{BaseModel: models.BaseModel{ID: "p-active"}, Status: constants.ProjectStatusActive},
		{BaseModel: models.BaseModel{ID: "p-cancelled"}, Status: constants.ProjectStatusCancelled},
	}, nil)
	service := newTeamServiceRecording(t, teams, projects, ProvideAuthzService(leadPeople{}), events)
	planner := asRole("kc-planner", constants.RolePlanner)

	require.NoError(t, service.Add(planner, teamProgramID, teamPersonID))
	assert.Equal(t, []string{teamPersonID}, teams.members[teamProgramID])
	require.Len(t, events.events, 1)
	assert.Equal(t, constants.EventTeamChanged, events.events[0].Type)
	assert.Equal(t, []string{"p-active"}, events.events[0].ProjectIDs, "only the projects a draft may cover")
	assert.Equal(t, constants.TeamChangeAdded, events.events[0].Payload.(TeamChange).Change)

	requireAppError(t, service.Add(planner, teamProgramID, teamPersonID), constants.TeamMemberExists, http.StatusConflict)
	requireStatus(t, service.Add(asRole("kc-manager", constants.RoleManager), teamProgramID, teamPersonID), http.StatusForbidden)
	assert.Len(t, events.events, 1, "a refused change records nothing")

	require.NoError(t, service.Remove(planner, teamProgramID, teamPersonID))
	require.Len(t, events.events, 2)
	assert.Equal(t, constants.TeamChangeRemoved, events.events[1].Payload.(TeamChange).Change)
	requireAppError(t, service.Remove(planner, teamProgramID, teamPersonID), constants.TeamMemberNotFound, http.StatusNotFound)
	assert.Len(t, events.events, 2, "removing someone who is not in the team records nothing")

	t.Run("a failed event fails the change", func(t *testing.T) {
		failing := &recordedEvents{fail: stderrors.New("outbox down")}
		svc := newTeamServiceRecording(t, &memoryTeams{members: map[string][]string{}}, projects, ProvideAuthzService(leadPeople{}), failing)
		require.Error(t, svc.Add(planner, teamProgramID, teamPersonID))
	})
}
