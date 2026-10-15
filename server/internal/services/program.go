package services

import (
	"context"
	"net/http"
	"slices"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
	"the-slot-finder/server/internal/request"
)

const (
	opListPrograms  = "list_programs"
	opGetProgram    = "get_program"
	opCreateProgram = "create_program"
	opUpdateProgram = "update_program"
	opProgramAction = "program_action"
	resourceProgram = "program"
)

// programTransitions lists, per action, the statuses it may start from and the status it sets.
var programTransitions = map[string]transition{
	constants.ProgramActionHold:         {from: []string{constants.ProgramStatusOpen}, to: constants.ProgramStatusPending},
	constants.ProgramActionResume:       {from: []string{constants.ProgramStatusPending}, to: constants.ProgramStatusOpen},
	constants.ProgramActionStartClosing: {from: []string{constants.ProgramStatusOpen}, to: constants.ProgramStatusClosing},
	constants.ProgramActionClose:        {from: []string{constants.ProgramStatusClosing}, to: constants.ProgramStatusClosed},
	constants.ProgramActionCancel: {
		from: []string{constants.ProgramStatusOpen, constants.ProgramStatusPending, constants.ProgramStatusClosing},
		to:   constants.ProgramStatusCancelled,
	},
}

// transition is one allowed status change.
type transition struct {
	from []string
	to   string
}

// ListProgramsInput pages the program list.
type ListProgramsInput struct {
	Page     int
	PageSize int
}

// CreateProgramInput holds a new program. LeadID is honoured for planners; a lead who creates
// a program becomes its lead.
type CreateProgramInput struct {
	Name       string
	Client     string
	StartDate  time.Time
	EndDate    time.Time
	Goal       string
	Objectives []string
	CharterURL *string
	LeadID     *string
}

// OptionalString is an optional text edit: Set false leaves it alone, Value nil clears it.
type OptionalString struct {
	Set   bool
	Value *string
}

// UpdateProgramInput holds the fields that change; nil fields stay as they are.
// Only planners may change the lead.
type UpdateProgramInput struct {
	Name       *string
	Client     *string
	StartDate  *time.Time
	EndDate    *time.Time
	Goal       *string
	Objectives *[]string
	CharterURL OptionalString
	LeadID     OptionalString
}

// ProgramService manages programs and their lifecycle.
type ProgramService interface {
	List(ctx context.Context, in ListProgramsInput) (*dtos.DataResponse[models.Program], error)
	Get(ctx context.Context, id string) (*models.Program, error)
	Create(ctx context.Context, in CreateProgramInput) (*models.Program, error)
	Update(ctx context.Context, id string, in UpdateProgramInput) (*models.Program, error)
	// ApplyAction changes the status through an action such as hold or cancel.
	ApplyAction(ctx context.Context, id, action string) (*models.Program, error)
}

type programService struct {
	programs repositories.ProgramRepository
	projects repositories.ProjectRepository
	people   PeopleService
	authz    AuthzService
	tx       db.TxManager
}

// ProvideProgramService creates the program service.
func ProvideProgramService(programs repositories.ProgramRepository, projects repositories.ProjectRepository,
	people PeopleService, authz AuthzService, tx db.TxManager,
) ProgramService {
	return &programService{programs: programs, projects: projects, people: people, authz: authz, tx: tx}
}

func (s *programService) List(ctx context.Context, in ListProgramsInput) (*dtos.DataResponse[models.Program], error) {
	reader, err := s.authz.ProgramReader(ctx)
	if err != nil {
		return nil, err
	}
	request := &dtos.PageableRequest{Page: in.Page, PageSize: in.PageSize}
	if !reader.All && reader.LeadID == "" {
		return emptyPage[models.Program](request), nil
	}
	page, err := s.programs.List(ctx, request, readerLeadID(reader))
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgProgramListFailed, err).WithOperation(opListPrograms).WithResource(resourceProgram)
	}
	return page, nil
}

func (s *programService) Get(ctx context.Context, id string) (*models.Program, error) {
	if _, err := s.authz.Require(ctx, CanUsePrograms); err != nil {
		return nil, err
	}
	program, err := s.find(ctx, id, opGetProgram)
	if err != nil {
		return nil, err
	}
	if err := requireProgramReader(ctx, s.authz, s.projects, program, opGetProgram); err != nil {
		return nil, err
	}
	return program, nil
}

func (s *programService) Create(ctx context.Context, in CreateProgramInput) (*models.Program, error) {
	principal, err := s.authz.Require(ctx, CanUsePrograms)
	if err != nil {
		return nil, err
	}
	if err := checkProgramDates(in.StartDate, in.EndDate, opCreateProgram); err != nil {
		return nil, err
	}
	leadID, err := chooseLead(ctx, s.people, principal, in.LeadID, opCreateProgram)
	if err != nil {
		return nil, err
	}

	program := &models.Program{
		BaseModel:  models.NewBaseModel(),
		Name:       in.Name,
		Client:     in.Client,
		Status:     constants.ProgramStatusOpen,
		StartDate:  in.StartDate,
		EndDate:    in.EndDate,
		Goal:       in.Goal,
		Objectives: nonNilStrings(in.Objectives),
		CharterURL: in.CharterURL,
		LeadID:     leadID,
	}
	if err := s.programs.Create(ctx, program); err != nil {
		return nil, errors.DatabaseError(constants.MsgProgramSaveFailed, err).WithOperation(opCreateProgram).WithResource(resourceProgram)
	}
	return s.find(ctx, program.ID, opCreateProgram)
}

func (s *programService) Update(ctx context.Context, id string, in UpdateProgramInput) (*models.Program, error) {
	program, err := s.find(ctx, id, opUpdateProgram)
	if err != nil {
		return nil, err
	}
	principal, err := s.authz.RequireLeadOf(ctx, CanManageAllPrograms, program.LeadID)
	if err != nil {
		return nil, err
	}

	applyProgramChanges(program, in)
	if in.LeadID.Set {
		if !CanManageAllPrograms(principal) {
			return nil, errors.ForbiddenError(constants.MsgInsufficientPermissions, nil).
				WithOperation(opUpdateProgram).WithResource(resourceProgram)
		}
		if program.LeadID, err = checkLeadExists(ctx, s.people, in.LeadID.Value, opUpdateProgram); err != nil {
			return nil, err
		}
	}
	if err := checkProgramDates(program.StartDate, program.EndDate, opUpdateProgram); err != nil {
		return nil, err
	}

	program.Lead = nil
	if err := s.programs.Save(ctx, program); err != nil {
		return nil, errors.DatabaseError(constants.MsgProgramSaveFailed, err).WithOperation(opUpdateProgram).WithResource(resourceProgram)
	}
	return s.find(ctx, id, opUpdateProgram)
}

func (s *programService) ApplyAction(ctx context.Context, id, action string) (*models.Program, error) {
	program, err := s.find(ctx, id, opProgramAction)
	if err != nil {
		return nil, err
	}
	if _, err := s.authz.RequireLeadOf(ctx, CanManageAllPrograms, program.LeadID); err != nil {
		return nil, err
	}
	next, err := nextStatus(programTransitions, program.Status, action, opProgramAction, resourceProgram)
	if err != nil {
		return nil, err
	}

	program.Status = next
	program.Lead = nil
	// Cancelling a program cancels its open projects in the same transaction.
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.programs.Save(ctx, program); err != nil {
			return err
		}
		if next == constants.ProgramStatusCancelled {
			return s.projects.CancelOpen(ctx, program.ID)
		}
		return nil
	})
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgProgramSaveFailed, err).WithOperation(opProgramAction).WithResource(resourceProgram)
	}
	return s.find(ctx, id, opProgramAction)
}

func (s *programService) find(ctx context.Context, id, op string) (*models.Program, error) {
	program, err := s.programs.FindWithLead(ctx, id)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.ProgramNotFound, constants.MsgProgramNotFound, resourceProgram, op)
	}
	return program, nil
}

// chooseLead picks the lead of a new program or project: planners choose (or leave it empty),
// a lead becomes the lead themselves.
func chooseLead(ctx context.Context, people PeopleService, principal request.Principal, requested *string, op string) (*string, error) {
	if CanManageAllPrograms(principal) {
		return checkLeadExists(ctx, people, requested, op)
	}
	self, err := people.FindForPrincipal(ctx, principal)
	if err != nil {
		return nil, err
	}
	if self == nil {
		return nil, errors.NewAppError(constants.NoPersonRecord, constants.MsgNoPersonRecord, errors.ErrorTypeForbidden, http.StatusForbidden).
			WithOperation(op)
	}
	return &self.ID, nil
}

// checkLeadExists returns leadID when it is nil or a known person, else a 400 error.
func checkLeadExists(ctx context.Context, people PeopleService, leadID *string, op string) (*string, error) {
	if leadID == nil {
		return nil, nil
	}
	if _, err := people.Get(ctx, *leadID); err != nil {
		if appErr := errors.GetAppError(err); appErr != nil && appErr.HTTPStatus == http.StatusNotFound {
			return nil, errors.NewAppError(constants.LeadNotFound, constants.MsgLeadNotFound, errors.ErrorTypeValidation, http.StatusBadRequest).
				WithOperation(op)
		}
		return nil, err
	}
	return leadID, nil
}

func applyProgramChanges(program *models.Program, in UpdateProgramInput) {
	if in.Name != nil {
		program.Name = *in.Name
	}
	if in.Client != nil {
		program.Client = *in.Client
	}
	if in.StartDate != nil {
		program.StartDate = *in.StartDate
	}
	if in.EndDate != nil {
		program.EndDate = *in.EndDate
	}
	if in.Goal != nil {
		program.Goal = *in.Goal
	}
	if in.Objectives != nil {
		program.Objectives = nonNilStrings(*in.Objectives)
	}
	if in.CharterURL.Set {
		program.CharterURL = in.CharterURL.Value
	}
}

// nextStatus returns the status an action leads to, or 409 when it is not allowed now.
func nextStatus(transitions map[string]transition, current, action, op, resource string) (string, error) {
	t, ok := transitions[action]
	if !ok {
		return "", errors.NewAppError(constants.UnknownAction, constants.MsgActionUnknown, errors.ErrorTypeValidation, http.StatusBadRequest).
			WithOperation(op).WithResource(resource)
	}
	if !slices.Contains(t.from, current) {
		return "", errors.NewAppError(constants.InvalidStateTransition, constants.MsgStateTransitionInvalid, errors.ErrorTypeConflict, http.StatusConflict).
			WithOperation(op).WithResource(resource).WithContext("status", current).WithContext("action", action)
	}
	return t.to, nil
}

func checkProgramDates(start, end time.Time, op string) error {
	if end.Before(start) {
		return errors.NewAppError(constants.ProgramDatesInvalid, constants.MsgProgramDatesInvalid, errors.ErrorTypeValidation, http.StatusBadRequest).
			WithOperation(op).WithResource(resourceProgram)
	}
	return nil
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
