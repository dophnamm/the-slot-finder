package services

import (
	"net/http"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const otherProgramID = "0192f3a4-0000-7000-8000-00000000d104"

func TestProgramMembers(t *testing.T) {
	checkout := &models.Project{BaseModel: models.BaseModel{ID: checkoutID}, ProgramID: metricsProgramID, Name: "Checkout"}
	loyalty := &models.Project{BaseModel: models.BaseModel{ID: loyaltyID}, ProgramID: metricsProgramID, Name: "Loyalty"}
	elsewhere := &models.Project{BaseModel: models.BaseModel{ID: "0192f3a4-0000-7000-8000-00000000d105"}, ProgramID: otherProgramID, Name: "Elsewhere"}
	demand := func(project *models.Project, expertise models.Expertise) *models.Demand {
		return &models.Demand{ProjectID: project.ID, Project: project, ExpertiseID: expertise.ID, Expertise: expertise}
	}
	bo := &models.Person{BaseModel: models.BaseModel{ID: busyPersonID}, Name: "Bo"}
	quinn := &models.Person{BaseModel: models.BaseModel{ID: qaOnlyID}, Name: "Quinn"}
	alloc := func(week string, hours float64) models.Allocation {
		return models.Allocation{Week: mustWeek(t, week), Hours: hours}
	}
	assignments := []models.Assignment{
		{PersonID: bo.ID, Person: bo, Kind: constants.AssignmentKindPrimary, Demand: demand(checkout, engineerExp),
			Allocations: []models.Allocation{alloc("2026-W45", 30), alloc("2026-W46", 30)}},
		{PersonID: bo.ID, Person: bo, Kind: constants.AssignmentKindCrossExpertise, Demand: demand(loyalty, qaExp),
			Allocations: []models.Allocation{alloc("2026-W44", 10), alloc("2026-W48", 10)}},
		{PersonID: quinn.ID, Person: quinn, Kind: constants.AssignmentKindPrimary, Demand: demand(checkout, qaExp),
			Allocations: []models.Allocation{alloc("2026-W47", 40)}},
		{PersonID: quinn.ID, Person: quinn, Kind: constants.AssignmentKindPrimary, Demand: demand(elsewhere, qaExp),
			Allocations: []models.Allocation{alloc("2026-W47", 99)}},
		{PersonID: quinn.ID, Person: quinn, Kind: constants.AssignmentKindPrimary, Demand: demand(loyalty, qaExp),
			Allocations: []models.Allocation{alloc("2026-W50", 0)}},
	}

	members := programMembers(assignments, metricsProgramID)

	require.Len(t, members, 2)
	assert.Equal(t, "Bo", members[0].Person.Name, "most hours first")
	assert.InDelta(t, 80, members[0].TotalHours, hoursEpsilon)
	assert.Equal(t, "2026-W44", members[0].FromWeek.String())
	assert.Equal(t, "2026-W48", members[0].ToWeek.String())
	require.Len(t, members[0].Expertises, 2)
	assert.Equal(t, "Engineer", members[0].Expertises[0].Expertise.Name)
	assert.Equal(t, constants.AssignmentKindPrimary, members[0].Expertises[0].Kind)
	assert.Equal(t, constants.AssignmentKindCrossExpertise, members[0].Expertises[1].Kind)
	assert.Equal(t, []string{"Checkout", "Loyalty"}, []string{members[0].Projects[0].Name, members[0].Projects[1].Name})

	assert.Equal(t, "Quinn", members[1].Person.Name)
	assert.InDelta(t, 40, members[1].TotalHours, hoursEpsilon, "another program's hours and an empty assignment are left out")
	require.Len(t, members[1].Projects, 1)
	assert.Equal(t, "Checkout", members[1].Projects[0].Name)
}

func TestInsightsService_ProgramMembers(t *testing.T) {
	lead := asRole("kc-lead", constants.RoleLead)

	t.Run("members of the program from the baseline", func(t *testing.T) {
		checkout := &models.Project{BaseModel: models.BaseModel{ID: checkoutID}, ProgramID: metricsProgramID, Name: "Checkout"}
		baseline := publishedBaseline(2)
		baseline.Assignments = []models.Assignment{{PersonID: qaOnlyID, Person: &models.Person{BaseModel: models.BaseModel{ID: qaOnlyID}, Name: "Quinn"},
			Kind: constants.AssignmentKindPrimary, Demand: &models.Demand{ProjectID: checkoutID, Project: checkout, ExpertiseID: qaExp.ID, Expertise: qaExp},
			Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W47"), Hours: 40}}}}
		plans, programs := new(MockPlanRepository), new(MockProgramRepository)
		plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
		leadID := leadPersonID
		programs.On("FindWithLead", mock.Anything, metricsProgramID).Return(&models.Program{BaseModel: models.BaseModel{ID: metricsProgramID}, LeadID: &leadID}, nil)
		svc := ProvideInsightsService(plans, nil, nil, programs, nil, nil, ProvideAuthzService(leadPeople{}), nil, heatmapConfig())

		members, err := svc.ProgramMembers(lead, metricsProgramID)
		require.NoError(t, err)
		require.Len(t, members, 1)
		assert.Equal(t, "Quinn", members[0].Person.Name)
		assert.InDelta(t, 40, members[0].TotalHours, hoursEpsilon)
	})

	t.Run("an unknown program", func(t *testing.T) {
		programs := new(MockProgramRepository)
		programs.On("FindWithLead", mock.Anything, otherProgramID).Return(nil, gorm.ErrRecordNotFound)
		svc := ProvideInsightsService(new(MockPlanRepository), nil, nil, programs, nil, nil, ProvideAuthzService(leadPeople{}), nil, heatmapConfig())
		_, err := svc.ProgramMembers(lead, otherProgramID)
		requireAppError(t, err, constants.ProgramNotFound, http.StatusNotFound)
	})

	t.Run("a specialist cannot", func(t *testing.T) {
		_, err := newMetricsService(t).ProgramMembers(asRole("kc-specialist", constants.RoleSpecialist), metricsProgramID)
		requireStatus(t, err, http.StatusForbidden)
	})
}
