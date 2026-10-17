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
	"the-slot-finder/server/internal/engine"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgramTeams_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	planner := asRole("kc-planner", constants.RolePlanner)
	manager := asRole("kc-manager", constants.RoleManager)

	qa := models.Expertise{BaseModel: models.NewBaseModel(), Name: "QA", Sequence: 4}
	require.NoError(t, pdb.Create(&qa).Error)
	newPerson := func(id, name string, level int) models.Person {
		p := models.Person{BaseModel: models.NewBaseModel(), Name: name, Email: name + "@example.test", Role: constants.RoleSpecialist,
			WeeklyHours: 40, NonProjectHours: 4, Active: true}
		if id != "" {
			p.ID = id
		}
		require.NoError(t, pdb.Create(&p).Error)
		require.NoError(t, pdb.Create(&models.PersonSkill{PersonID: p.ID, ExpertiseID: qa.ID, Level: level, IsPrimary: true}).Error)
		return p
	}
	lee := newPerson(leadPersonID, "lee", 1) // the manager of Acme; leadPeople maps kc-manager to this id
	ada := newPerson("", "ada", 3)
	bea := newPerson("", "bea", 5) // stronger than Ada, but not in Acme's team

	newProgram := func(name string, lead *string) models.Program {
		p := models.Program{BaseModel: models.NewBaseModel(), Name: name, Client: name, Status: constants.ProgramStatusOpen,
			StartDate: day(time.October, 12), EndDate: day(time.December, 31), Objectives: []string{}, LeadID: lead}
		require.NoError(t, pdb.Create(&p).Error)
		return p
	}
	newProject := func(program models.Program, name string) (models.Project, models.Demand) {
		p := models.Project{BaseModel: models.NewBaseModel(), ProgramID: program.ID, Name: name, Priority: 1,
			Status: constants.ProjectStatusPlanned, StartEarliest: day(time.October, 12), Deadline: day(time.October, 23)}
		require.NoError(t, pdb.Create(&p).Error)
		d := models.Demand{BaseModel: models.NewBaseModel(), ProjectID: p.ID, ExpertiseID: qa.ID, MinLevel: 2, Headcount: 1,
			TotalHours: 40, Profile: constants.DemandProfileFlat}
		require.NoError(t, pdb.Create(&d).Error)
		return p, d
	}
	acme, beta, gamma := newProgram("Acme", &lee.ID), newProgram("Beta", nil), newProgram("Gamma", nil)
	checkout, checkoutDemand := newProject(acme, "Checkout")
	portal, _ := newProject(beta, "Portal")
	dashboard, _ := newProject(gamma, "Dashboard")

	teamRepo := repositories.ProvideProgramTeamRepository(pdb)
	personRepo := repositories.ProvidePersonRepository(pdb)
	plans := repositories.ProvidePlanRepository(pdb)
	authz := ProvideTeamAuthzService(leadPeople{}, teamRepo)
	teams := ProvideProgramTeamService(teamRepo, repositories.ProvideProgramRepository(pdb), repositories.ProvideProjectRepository(pdb),
		personRepo, plans, authz, db.ProvideTxManager(pdb), &recordedEvents{})
	staffing := ProvideStaffingService(plans, repositories.ProvideProjectRepository(pdb), personRepo,
		repositories.ProvideAbsenceRepository(pdb), repositories.ProvideAuditRepository(pdb), repositories.ProvideChangeEventRepository(pdb),
		leadPeople{}, authz, db.ProvideTxManager(pdb), clock.NewFixed(acceptedAt), testPlanningConfig(), teamRepo)
	require.NoError(t, plans.EnsureBaseline(planner))

	require.NoError(t, teams.Add(planner, acme.ID, ada.ID))
	require.NoError(t, teams.Add(planner, beta.ID, bea.ID))

	t.Run("a program's project list carries its manager for the client's staffing rights", func(t *testing.T) {
		projects, err := repositories.ProvideProjectRepository(pdb).ListByProgram(planner, acme.ID)
		require.NoError(t, err)
		require.Len(t, projects, 1)
		require.NotNil(t, projects[0].Program)
		assert.Equal(t, &lee.ID, projects[0].Program.LeadID)
	})

	t.Run("a draft staffs a program from its team only", func(t *testing.T) {
		draft, err := staffing.Generate(planner, GenerateDraftInput{ProjectIDs: []string{checkout.ID}})
		require.NoError(t, err)
		require.Len(t, draft.Plan.Assignments, 1)
		assert.Equal(t, ada.ID, draft.Plan.Assignments[0].PersonID, "Bea is stronger but not in Acme's team")
		assert.Equal(t, map[string]bool{ada.ID: true}, draft.Teams[acme.ID], "the draft carries the team for Add person")
	})

	t.Run("a demand nobody in the team can take is unmet", func(t *testing.T) {
		draft, err := staffing.Generate(planner, GenerateDraftInput{ProjectIDs: []string{dashboard.ID}})
		require.NoError(t, err)
		require.Len(t, draft.Plan.Unmet, 1)
		assert.Equal(t, string(engine.UnmetNotInTeam), draft.Plan.Unmet[0].Reason)
		assert.Empty(t, draft.Teams[gamma.ID])
	})

	t.Run("the manager staffs the programs they lead, and only those", func(t *testing.T) {
		draft, err := staffing.Generate(manager, GenerateDraftInput{ProjectIDs: []string{checkout.ID}})
		require.NoError(t, err)

		_, err = staffing.Edit(manager, draft.Plan.ID, EditDraftInput{Add: []AddAssignmentInput{{PersonID: bea.ID, DemandID: checkoutDemand.ID}}})
		requireAppError(t, err, constants.PersonNotInTeam, http.StatusUnprocessableEntity)

		_, err = staffing.Accept(manager, draft.Plan.ID, AcceptDraftInput{BaseVersion: draft.Plan.Version})
		require.NoError(t, err)

		_, err = staffing.Generate(manager, GenerateDraftInput{ProjectIDs: []string{portal.ID}})
		requireStatus(t, err, http.StatusForbidden)
		_, err = staffing.Generate(manager, GenerateDraftInput{ProjectIDs: []string{checkout.ID, portal.ID}})
		requireStatus(t, err, http.StatusForbidden)
	})

	t.Run("the manager sees Acme's team on People, with their hours on the Team tab", func(t *testing.T) {
		scope, err := authz.VisiblePeople(manager)
		require.NoError(t, err)
		assert.True(t, scope.Allows(ada.ID))
		assert.False(t, scope.Allows(bea.ID))

		members, err := teams.List(manager, acme.ID)
		require.NoError(t, err)
		require.Len(t, members, 1)
		require.NotNil(t, members[0].Hours, "the accepted draft gave Ada hours")
		assert.InDelta(t, 40, members[0].Hours.TotalHours, 0.01)
	})
}
