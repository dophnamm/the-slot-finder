//go:build integration

package services

import (
	"net/http"
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

// A manager lists, opens and sees in the intake queue only the programs they lead or lead a
// project in; a planner sees them all.
func TestProgramVisibility_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)

	person := func(name string) models.Person {
		p := models.Person{BaseModel: models.NewBaseModel(), Name: name, Email: name + "@example.test", Role: constants.RoleLead,
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
	project := func(of models.Program, name string, lead models.Person) {
		p := models.Project{BaseModel: models.NewBaseModel(), ProgramID: of.ID, Name: name, Priority: 1, LeadID: &lead.ID,
			Status: constants.ProjectStatusIncoming, StartEarliest: time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC),
			Deadline: time.Date(2026, time.December, 11, 0, 0, 0, 0, time.UTC)}
		require.NoError(t, pdb.Create(&p).Error)
	}
	project(acme, "Checkout", lee)
	project(beta, "Patient App", lee) // Lee leads a project in Minh's program
	project(gamma, "Records", minh)

	people := selfFinder{"kc-lee": lee.ID}
	authz := ProvideAuthzService(people)
	projects, programRepo := repositories.ProvideProjectRepository(pdb), repositories.ProvideProgramRepository(pdb)
	programs := ProvideProgramService(programRepo, projects, leadPeople{}, authz, db.ProvideTxManager(pdb))
	projectService := ProvideProjectService(projects, programRepo, knownExpertisePeople{}, authz, db.ProvideTxManager(pdb), &recordedEvents{})
	manager, planner := asRole("kc-lee", constants.RoleLead), asRole("kc-planner", constants.RolePlanner)
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
		repositories.ProvidePlanRepository(pdb), teamAuthz)
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
