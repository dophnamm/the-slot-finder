//go:build integration

package services

import (
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A manager lists, opens and sees in the intake queue only the programs they lead or lead a
// project in; a planner sees them all.
func TestProgramVisibility_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)

	person := func(name string) models.Person {
		p := models.Person{BaseModel: models.NewBaseModel(), Name: name, Email: name + "@example.test", Role: constants.RoleManager,
			WeeklyHours: 40, Active: true}
		require.NoError(t, pdb.Create(&p).Error)
		return p
	}
	lee, minh := person("lee"), person("minh")
	program := func(name string, lead models.Person) models.Program {
		p := models.Program{BaseModel: models.NewBaseModel(), Name: name, Client: name, Status: constants.ProgramStatusOpen, LeadID: &lead.ID,
			StartDate: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC),
			Objectives: []string{}}
		require.NoError(t, pdb.Create(&p).Error)
		return p
	}
	acme, beta, gamma := program("Acme", lee), program("Beta", minh), program("Gamma", minh)
	project := func(of models.Program, name string, lead models.Person) models.Project {
		p := models.Project{BaseModel: models.NewBaseModel(), ProgramID: of.ID, Name: name, Priority: 1, LeadID: &lead.ID,
			Status: constants.ProjectStatusIncoming, StartEarliest: time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC),
			Deadline: time.Date(2026, time.December, 11, 0, 0, 0, 0, time.UTC)}
		require.NoError(t, pdb.Create(&p).Error)
		return p
	}
	checkout := project(acme, "Checkout", lee)
	patientApp := project(beta, "Patient App", lee) // Lee leads a project in Minh's program
	records := project(gamma, "Records", minh)

	people := selfFinder{"kc-lee": lee.ID}
	authz := ProvideAuthzService(people)
	projects, programRepo := repositories.ProvideProjectRepository(pdb), repositories.ProvideProgramRepository(pdb)
	programs := ProvideProgramService(programRepo, projects, leadPeople{}, authz, db.ProvideTxManager(pdb), &recordedEvents{})
	projectService := ProvideProjectService(projects, programRepo, knownExpertisePeople{}, authz, db.ProvideTxManager(pdb), &recordedEvents{}, projectConfig(), clock.NewFixed(acceptedAt))
	manager, planner := asRole("kc-lee", constants.RoleManager), asRole("kc-planner", constants.RolePlanner)
	page := ListProgramsInput{Page: 1, PageSize: 20}

	listed, err := programs.List(manager, page)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Acme", "Beta"}, programNames(listed.Data))
	assert.EqualValues(t, 2, listed.Pageable.Total)

	all, err := programs.List(planner, page)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Acme", "Beta", "Gamma"}, programNames(all.Data))

	_, err = programs.Get(manager, beta.ID)
	require.NoError(t, err, "a project they lead opens the program")
	_, err = programs.Get(manager, gamma.ID)
	requireStatus(t, err, http.StatusForbidden)

	intake, err := projectService.ListByStatus(manager, ListProjectsInput{Status: constants.ProjectStatusIncoming, Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Checkout", "Patient App"}, names(intake.Data))

	// A specialist reads the programs where the published baseline gives them hours, and no others.
	sam := person("sam")
	sam.Role = constants.RoleSpecialist
	require.NoError(t, pdb.Save(&sam).Error)
	qa := models.Expertise{BaseModel: models.NewBaseModel(), Name: "QA"}
	require.NoError(t, pdb.Create(&qa).Error)
	plan := func(kind, status string) models.Plan {
		p := models.Plan{BaseModel: models.NewBaseModel(), Kind: kind, Status: status, ProjectIDs: []string{}, Unmet: []models.PlanUnmet{}}
		require.NoError(t, pdb.Create(&p).Error)
		return p
	}
	assign := func(in models.Plan, on models.Project, hours float64) (models.Assignment, models.Allocation) {
		d := models.Demand{BaseModel: models.NewBaseModel(), ProjectID: on.ID, ExpertiseID: qa.ID, MinLevel: 1, Headcount: 1, TotalHours: 8, Profile: "flat"}
		require.NoError(t, pdb.Create(&d).Error)
		a := models.Assignment{BaseModel: models.NewBaseModel(), PlanID: in.ID, PersonID: sam.ID, DemandID: d.ID, Kind: constants.AssignmentKindPrimary,
			Reason: []string{}}
		require.NoError(t, pdb.Create(&a).Error)
		al := models.Allocation{BaseModel: models.NewBaseModel(), AssignmentID: a.ID, Week: isoweek.Of(acceptedAt), Hours: hours}
		require.NoError(t, pdb.Create(&al).Error)
		return a, al
	}
	baseline, draft := plan(constants.PlanKindBaseline, constants.PlanStatusPublished), plan(constants.PlanKindDraft, constants.PlanStatusDraft)
	assign(baseline, records, 8)  // hours in the baseline
	assign(baseline, checkout, 0) // an assignment without hours
	assign(draft, patientApp, 8)  // hours only in a draft
	people["kc-sam"] = sam.ID
	specialist := asRole("kc-sam", constants.RoleSpecialist)

	seen, err := programs.List(specialist, page)
	require.NoError(t, err)
	assert.Equal(t, []string{"Gamma"}, programNames(seen.Data))
	_, err = programs.Get(specialist, gamma.ID)
	require.NoError(t, err)
	for _, closed := range []models.Program{acme, beta} {
		_, err = programs.Get(specialist, closed.ID)
		requireStatus(t, err, http.StatusForbidden)
	}
	opened, err := projectService.Get(specialist, records.ID)
	require.NoError(t, err)
	assert.Empty(t, opened.Demands, "no budget hours for a specialist")
	_, err = projectService.ListByStatus(specialist, ListProjectsInput{Page: 1, PageSize: 20})
	requireStatus(t, err, http.StatusForbidden)

	// Hours that were deleted, or sit in a superseded baseline, open nothing.
	deletedAssignment, deletedAllocation, oldBaseline := program("Delta", minh), program("Epsilon", minh), program("Zeta", minh)
	deadAssignment, _ := assign(baseline, project(deletedAssignment, "Dead Assignment", minh), 8)
	require.NoError(t, pdb.Delete(&deadAssignment).Error)
	_, deadAllocation := assign(baseline, project(deletedAllocation, "Dead Allocation", minh), 8)
	require.NoError(t, pdb.Delete(&deadAllocation).Error)
	assign(plan(constants.PlanKindBaseline, constants.PlanStatusSuperseded), project(oldBaseline, "Old Baseline", minh), 8)
	seen, err = programs.List(specialist, page)
	require.NoError(t, err)
	assert.Equal(t, []string{"Gamma"}, programNames(seen.Data))
	for _, closed := range []models.Program{deletedAssignment, deletedAllocation, oldBaseline} {
		_, err = programs.Get(specialist, closed.ID)
		requireStatus(t, err, http.StatusForbidden)
	}

	// Inside Gamma, a specialist reads only the projects they work on: not one without hours,
	// one with a deleted allocation, or one with hours only in a draft.
	_ = project(gamma, "Untouched", minh)
	_, deadHere := assign(baseline, project(gamma, "Dead Here", minh), 8)
	require.NoError(t, pdb.Delete(&deadHere).Error)
	assign(draft, project(gamma, "Draft Here", minh), 8)
	inGamma, err := projectService.ListByProgram(specialist, gamma.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"Records"}, names(inGamma))
	for _, closed := range []string{"Untouched", "Dead Here", "Draft Here"} {
		var p models.Project
		require.NoError(t, pdb.Where("name = ?", closed).First(&p).Error)
		_, err = projectService.Get(specialist, p.ID)
		requireStatus(t, err, http.StatusForbidden)
	}

	// People: the teams of the programs Lee leads or leads a project in, and Lee.
	ada, bao, cam := person("ada"), person("bao"), person("cam")
	for _, member := range []models.ProgramTeamMember{{ProgramID: acme.ID, PersonID: ada.ID}, {ProgramID: beta.ID, PersonID: bao.ID},
		{ProgramID: gamma.ID, PersonID: cam.ID}} {
		require.NoError(t, pdb.Create(&member).Error)
	}
	teamRepo := repositories.ProvideProgramTeamRepository(pdb)
	teamAuthz := ProvideTeamAuthzService(people, teamRepo)
	visible, err := teamAuthz.VisiblePeople(manager)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{lee.ID, ada.ID, bao.ID}, visible.IDs())

	teams := ProvideProgramTeamService(teamRepo, programRepo, projects, repositories.ProvidePersonRepository(pdb),
		repositories.ProvidePlanRepository(pdb), teamAuthz, db.ProvideTxManager(pdb), &recordedEvents{})
	_, err = teams.List(manager, beta.ID)
	require.NoError(t, err, "a project they lead opens the program's team")
	_, err = teams.List(manager, gamma.ID)
	requireStatus(t, err, http.StatusForbidden)
}

func programNames(programs []models.Program) []string {
	out := make([]string, 0, len(programs))
	for _, p := range programs {
		out = append(out, p.Name)
	}
	return out
}
