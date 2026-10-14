package services

import (
	"context"
	stderrors "errors"
	"net/http"

	"the-slot-finder/server/internal/constants"
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
	people   repositories.PersonRepository
	plans    repositories.PlanRepository
	authz    AuthzService
}

// ProvideProgramTeamService creates the program team service.
func ProvideProgramTeamService(teams repositories.ProgramTeamRepository, programs repositories.ProgramRepository,
	people repositories.PersonRepository, plans repositories.PlanRepository, authz AuthzService,
) ProgramTeamService {
	return &programTeamService{teams: teams, programs: programs, people: people, plans: plans, authz: authz}
}

func (s *programTeamService) List(ctx context.Context, programID string) ([]models.TeamMember, error) {
	program, err := s.programs.FindWithLead(ctx, programID)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.ProgramNotFound, constants.MsgProgramNotFound, resourceProgram, opListTeam)
	}
	if _, err := s.authz.RequireLeadOf(ctx, CanManageAllPrograms, program.LeadID); err != nil {
		return nil, err
	}
	members, err := s.teams.ListByProgram(ctx, programID)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opListTeam).WithResource(resourceTeam)
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

func (s *programTeamService) Add(ctx context.Context, programID, personID string) error {
	if _, err := s.authz.Require(ctx, CanManageAllPrograms); err != nil {
		return err
	}
	if _, err := s.programs.FindWithLead(ctx, programID); err != nil {
		return notFoundOrDatabase(err, constants.ProgramNotFound, constants.MsgProgramNotFound, resourceProgram, opAddTeamMember)
	}
	if _, err := s.people.FindWithSkills(ctx, personID); err != nil {
		return notFoundOrDatabase(err, constants.PersonNotFound, constants.MsgPersonNotFound, resourcePerson, opAddTeamMember)
	}
	if err := s.teams.Add(ctx, &models.ProgramTeamMember{ProgramID: programID, PersonID: personID}); err != nil {
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
	removed, err := s.teams.Remove(ctx, programID, personID)
	if err != nil {
		return errors.DatabaseError(constants.MsgTeamSaveFailed, err).WithOperation(opRemoveTeamMember).WithResource(resourceTeam)
	}
	if !removed {
		return errors.NewAppError(constants.TeamMemberNotFound, constants.MsgTeamMemberNotFound, errors.ErrorTypeNotFound,
			http.StatusNotFound).WithOperation(opRemoveTeamMember).WithResource(resourceTeam)
	}
	return nil
}
