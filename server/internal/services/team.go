package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"slices"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"gorm.io/gorm"
)

const (
	opListTeam         = "list_program_team"
	opAddTeamMember    = "add_team_member"
	opRemoveTeamMember = "remove_team_member"
	resourceTeam       = "program_team"
)

// ProgramTeamService manages the people a planner gives each program. The program's manager
// reads the team and staffs its projects from it.
type ProgramTeamService interface {
	// List returns the team, each member with their skills and their hours in the program's
	// baseline. The planner and the program's manager may read it.
	List(ctx context.Context, programID string) ([]models.TeamMember, error)
	// Add puts a person in the team; planners only.
	Add(ctx context.Context, programID, personID string) error
	// Remove takes a person out of the team; their baseline hours stay. Planners only.
	Remove(ctx context.Context, programID, personID string) error
}

type programTeamService struct {
	teams    repositories.ProgramTeamRepository
	programs repositories.ProgramRepository
	projects repositories.ProjectRepository
	people   repositories.PersonRepository
	plans    repositories.PlanRepository
	authz    AuthzService
	tx       db.TxManager
	events   EventRecorder
}

// ProvideProgramTeamService creates the program team service.
func ProvideProgramTeamService(teams repositories.ProgramTeamRepository, programs repositories.ProgramRepository,
	projects repositories.ProjectRepository, people repositories.PersonRepository, plans repositories.PlanRepository,
	authz AuthzService, tx db.TxManager, events EventRecorder,
) ProgramTeamService {
	return &programTeamService{teams: teams, programs: programs, projects: projects, people: people, plans: plans, authz: authz,
		tx: tx, events: events}
}

func (s *programTeamService) List(ctx context.Context, programID string) ([]models.TeamMember, error) {
	program, err := s.programs.FindWithLead(ctx, programID)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.ProgramNotFound, constants.MsgProgramNotFound, resourceProgram, opListTeam)
	}
	if err := requireProgramReader(ctx, s.authz, s.projects, program, opListTeam); err != nil {
		return nil, err
	}
	principal, err := s.authz.Principal(ctx)
	if err != nil {
		return nil, err
	}
	members, err := s.teams.ListByProgram(ctx, programID)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opListTeam).WithResource(resourceTeam)
	}
	if !CanUsePrograms(principal) {
		return specialistTeam(members), nil
	}
	baseline, err := findBaseline(ctx, s.plans, opListTeam, true)
	if err != nil {
		return nil, err
	}
	hours := map[string]models.ProgramMember{}
	for _, member := range programMembers(baseline.Assignments, programID) {
		hours[member.Person.ID] = member
	}
	out := make([]models.TeamMember, 0, len(members))
	for _, m := range members {
		member := models.TeamMember{Person: *m.Person}
		if h, ok := hours[m.PersonID]; ok {
			member.Hours = &h
		}
		out = append(out, member)
	}
	return out, nil
}

// specialistTeam is a program's team as a specialist may see it: names only,
// with no skills, hours, levels, contact details, capacity or dates.
func specialistTeam(members []models.ProgramTeamMember) []models.TeamMember {
	out := make([]models.TeamMember, 0, len(members))
	for _, m := range members {
		person := models.Person{BaseModel: models.BaseModel{ID: m.Person.ID}, Name: m.Person.Name, Active: m.Person.Active,
			Skills: []models.PersonSkill{}}
		out = append(out, models.TeamMember{Person: person})
	}
	return out
}

func (s *programTeamService) Add(ctx context.Context, programID, personID string) error {
	if _, err := s.authz.Require(ctx, CanManageAllPrograms); err != nil {
		return err
	}
	program, person, err := s.findProgramAndPerson(ctx, programID, personID, opAddTeamMember)
	if err != nil {
		return err
	}
	affected, err := s.plannableProjects(ctx, programID, opAddTeamMember)
	if err != nil {
		return err
	}
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.teams.Add(ctx, &models.ProgramTeamMember{ProgramID: programID, PersonID: personID}); err != nil {
			return err
		}
		return s.events.Record(ctx, ChangeInput{Type: constants.EventTeamChanged, ProjectIDs: affected,
			Payload: TeamChange{ProgramID: programID, Program: program.Name, PersonID: personID, PersonName: person.Name,
				Change: constants.TeamChangeAdded}})
	})
	if err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.WrapError(err, constants.TeamMemberExists, constants.MsgTeamMemberExists, errors.ErrorTypeConflict,
				http.StatusConflict).WithOperation(opAddTeamMember).WithResource(resourceTeam)
		}
		return errors.DatabaseError(constants.MsgTeamSaveFailed, err).WithOperation(opAddTeamMember).WithResource(resourceTeam)
	}
	return nil
}

func (s *programTeamService) Remove(ctx context.Context, programID, personID string) error {
	if _, err := s.authz.Require(ctx, CanManageAllPrograms); err != nil {
		return err
	}
	program, person, err := s.findProgramAndPerson(ctx, programID, personID, opRemoveTeamMember)
	if err != nil {
		return err
	}
	affected, err := s.plannableProjects(ctx, programID, opRemoveTeamMember)
	if err != nil {
		return err
	}
	removed := false
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		var err error
		if removed, err = s.teams.Remove(ctx, programID, personID); err != nil || !removed {
			return err
		}
		return s.events.Record(ctx, ChangeInput{Type: constants.EventTeamChanged, ProjectIDs: affected,
			Payload: TeamChange{ProgramID: programID, Program: program.Name, PersonID: personID, PersonName: person.Name,
				Change: constants.TeamChangeRemoved}})
	})
	if err != nil {
		return errors.DatabaseError(constants.MsgTeamSaveFailed, err).WithOperation(opRemoveTeamMember).WithResource(resourceTeam)
	}
	if !removed {
		return errors.NewAppError(constants.TeamMemberNotFound, constants.MsgTeamMemberNotFound, errors.ErrorTypeNotFound,
			http.StatusNotFound).WithOperation(opRemoveTeamMember).WithResource(resourceTeam)
	}
	return nil
}

// findProgramAndPerson loads both, so a change event can name them.
func (s *programTeamService) findProgramAndPerson(ctx context.Context, programID, personID, op string) (*models.Program, *models.Person, error) {
	program, err := s.programs.FindWithLead(ctx, programID)
	if err != nil {
		return nil, nil, notFoundOrDatabase(err, constants.ProgramNotFound, constants.MsgProgramNotFound, resourceProgram, op)
	}
	person, err := s.people.FindWithSkills(ctx, personID)
	if err != nil {
		return nil, nil, notFoundOrDatabase(err, constants.PersonNotFound, constants.MsgPersonNotFound, resourcePerson, op)
	}
	return program, person, nil
}

// plannableProjects is the ids of the program's projects a draft may cover: the ones a team
// change can affect.
func (s *programTeamService) plannableProjects(ctx context.Context, programID, op string) ([]string, error) {
	projects, err := s.projects.ListByProgram(ctx, programID)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceTeam)
	}
	ids := make([]string, 0, len(projects))
	for _, p := range projects {
		if slices.Contains(constants.PlannableProjectStatuses, p.Status) {
			ids = append(ids, p.ID)
		}
	}
	return ids, nil
}
