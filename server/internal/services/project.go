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
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
	"the-slot-finder/server/internal/request"
)

const (
	opListProjects     = "list_projects"
	opGetProject       = "get_project"
	opCreateProject    = "create_project"
	opReorderProjects  = "reorder_projects"
	opUpdateProject    = "update_project"
	opProjectAction    = "project_action"
	opReplacePhases    = "replace_phases"
	opUpdatePhase      = "update_phase"
	opReplaceDemands   = "replace_demands"
	resourceProject    = "project"
	resourcePhase      = "phase"
	resourceDemand     = "demand"
	firstPhaseSequence = 1
)

// openProjectStatuses are the statuses a project can be held, completed or cancelled from.
var openProjectStatuses = []string{constants.ProjectStatusIncoming, constants.ProjectStatusPlanned, constants.ProjectStatusActive}

// projectTransitions lists, per action, the statuses it may start from and the status it sets.
// Until staffing exists, resume goes back to incoming.
var projectTransitions = map[string]transition{
	constants.ProjectActionHold:     {from: openProjectStatuses, to: constants.ProjectStatusOnHold},
	constants.ProjectActionResume:   {from: []string{constants.ProjectStatusOnHold}, to: constants.ProjectStatusIncoming},
	constants.ProjectActionComplete: {from: openProjectStatuses, to: constants.ProjectStatusDone},
	constants.ProjectActionCancel: {
		from: append(slices.Clone(openProjectStatuses), constants.ProjectStatusOnHold),
		to:   constants.ProjectStatusCancelled,
	},
}

// programStatusesAcceptingProjects are the program statuses in which projects can be added.
var programStatusesAcceptingProjects = []string{constants.ProgramStatusOpen, constants.ProgramStatusPending}

// CreateProjectInput holds a new project in a program. LeadID is honoured for planners;
// a lead who adds a project becomes its lead.
type CreateProjectInput struct {
	ProgramID string
	Name      string
	// Priority is the place in the program's order; nil puts the project last.
	Priority      *int
	StartEarliest time.Time
	Deadline      time.Time
	ScopePoints   float64
	DonePoints    float64
	LeadID        *string
}

// UpdateProjectInput holds the fields that change; nil fields stay as they are.
// Only planners may change the lead.
type UpdateProjectInput struct {
	Name          *string
	Priority      *int
	StartEarliest *time.Time
	Deadline      *time.Time
	ScopePoints   *float64
	DonePoints    *float64
	LeadID        OptionalString
}

// ListProjectsInput filters and pages projects across programs; empty Status means all.
type ListProjectsInput struct {
	Status   string
	Page     int
	PageSize int
}

// PhaseInput is one phase of a full list; ID nil adds a phase. Order in the list is the sequence.
type PhaseInput struct {
	ID        *string
	Name      string
	StartWeek isoweek.Week
	EndWeek   isoweek.Week
}

// UpdatePhaseInput holds the phase fields that change.
type UpdatePhaseInput struct {
	Name      *string
	StartWeek *isoweek.Week
	EndWeek   *isoweek.Week
}

// DemandInput is one demand row of a full list; ID nil adds a row, PhaseID nil means the whole project.
type DemandInput struct {
	ID          *string
	ExpertiseID string
	PhaseID     *string
	MinLevel    int
	Headcount   int
	TotalHours  float64
	Profile     string
}

// ProjectService manages projects, their phases and their demand.
type ProjectService interface {
	ListByProgram(ctx context.Context, programID string) ([]models.Project, error)
	ListByStatus(ctx context.Context, in ListProjectsInput) (*dtos.DataResponse[models.Project], error)
	Get(ctx context.Context, id string) (*models.Project, error)
	Create(ctx context.Context, in CreateProjectInput) (*models.Project, error)
	// Reorder sets the order of a program's projects: the first id gets priority 1, the next 2,
	// and so on. The ids must be each of the program's projects exactly once.
	Reorder(ctx context.Context, programID string, projectIDs []string) ([]models.Project, error)
	Update(ctx context.Context, id string, in UpdateProjectInput) (*models.Project, error)
	ApplyAction(ctx context.Context, id, action string) (*models.Project, error)
	// ReplacePhases saves the full ordered list: listed ids are updated, new rows added, missing ones removed.
	ReplacePhases(ctx context.Context, projectID string, phases []PhaseInput) (*models.Project, error)
	// UpdatePhase edits one phase. Moving its end week later by N weeks moves every later phase
	// by N weeks in the same transaction; moving it earlier moves nothing else. It returns the
	// phases that changed.
	UpdatePhase(ctx context.Context, phaseID string, in UpdatePhaseInput) ([]models.ProjectPhase, error)
	// ReplaceDemands saves the full demand list the same way as ReplacePhases.
	ReplaceDemands(ctx context.Context, projectID string, demands []DemandInput) (*models.Project, error)
}

type projectService struct {
	projects repositories.ProjectRepository
	programs repositories.ProgramRepository
	people   PeopleService
	authz    AuthzService
	tx       db.TxManager
	events   EventRecorder
}

// ProvideProjectService creates the project service.
func ProvideProjectService(projects repositories.ProjectRepository, programs repositories.ProgramRepository,
	people PeopleService, authz AuthzService, tx db.TxManager, events EventRecorder,
) ProjectService {
	return &projectService{projects: projects, programs: programs, people: people, authz: authz, tx: tx, events: events}
}

func (s *projectService) ListByProgram(ctx context.Context, programID string) ([]models.Project, error) {
	if _, err := s.authz.Require(ctx, CanUsePrograms); err != nil {
		return nil, err
	}
	program, err := s.findProgram(ctx, programID, opListProjects)
	if err != nil {
		return nil, err
	}
	if err := requireProgramReader(ctx, s.authz, s.projects, program, opListProjects); err != nil {
		return nil, err
	}
	projects, err := s.projects.ListByProgram(ctx, programID)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgProjectListFailed, err).WithOperation(opListProjects).WithResource(resourceProject)
	}
	return projects, nil
}

func (s *projectService) ListByStatus(ctx context.Context, in ListProjectsInput) (*dtos.DataResponse[models.Project], error) {
	reader, err := s.authz.ProgramReader(ctx)
	if err != nil {
		return nil, err
	}
	request := &dtos.PageableRequest{Page: in.Page, PageSize: in.PageSize}
	if !reader.All && reader.LeadID == "" {
		return emptyPage[models.Project](request), nil
	}
	page, err := s.projects.ListByStatus(ctx, in.Status, readerLeadID(reader), request)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgProjectListFailed, err).WithOperation(opListProjects).WithResource(resourceProject)
	}
	return page, nil
}

func (s *projectService) Get(ctx context.Context, id string) (*models.Project, error) {
	if _, err := s.authz.Require(ctx, CanUsePrograms); err != nil {
		return nil, err
	}
	project, err := s.find(ctx, id, opGetProject)
	if err != nil {
		return nil, err
	}
	if err := requireProgramReader(ctx, s.authz, s.projects, programOf(project), opGetProject); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *projectService) Create(ctx context.Context, in CreateProjectInput) (*models.Project, error) {
	program, err := s.findProgram(ctx, in.ProgramID, opCreateProject)
	if err != nil {
		return nil, err
	}
	principal, err := s.authz.RequireLeadOf(ctx, CanManageAllPrograms, program.LeadID)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(programStatusesAcceptingProjects, program.Status) {
		return nil, errors.NewAppError(constants.ProgramNotOpen, constants.MsgProgramNotOpen, errors.ErrorTypeConflict, http.StatusConflict).
			WithOperation(opCreateProject).WithResource(resourceProject)
	}
	if err := checkProjectDates(in.StartEarliest, in.Deadline, opCreateProject); err != nil {
		return nil, err
	}
	leadID, err := chooseLead(ctx, s.people, principal, in.LeadID, opCreateProject)
	if err != nil {
		return nil, err
	}
	priority, err := s.newPriority(ctx, program.ID, in.Priority)
	if err != nil {
		return nil, err
	}

	project := &models.Project{
		BaseModel:     models.NewBaseModel(),
		ProgramID:     program.ID,
		Name:          in.Name,
		Priority:      priority,
		Status:        constants.ProjectStatusIncoming,
		StartEarliest: in.StartEarliest,
		Deadline:      in.Deadline,
		ScopePoints:   in.ScopePoints,
		DonePoints:    in.DonePoints,
		LeadID:        leadID,
	}
	if err := s.projects.Create(ctx, project); err != nil {
		return nil, errors.DatabaseError(constants.MsgProjectSaveFailed, err).WithOperation(opCreateProject).WithResource(resourceProject)
	}
	return s.find(ctx, project.ID, opCreateProject)
}

// newPriority is the given priority, or the place after the program's last project.
func (s *projectService) newPriority(ctx context.Context, programID string, priority *int) (int, error) {
	if priority != nil {
		return *priority, nil
	}
	last, err := s.projects.MaxPriority(ctx, programID)
	if err != nil {
		return 0, errors.DatabaseError(constants.MsgProjectSaveFailed, err).WithOperation(opCreateProject).WithResource(resourceProject)
	}
	return last + 1, nil
}

func (s *projectService) Reorder(ctx context.Context, programID string, projectIDs []string) ([]models.Project, error) {
	program, err := s.findProgram(ctx, programID, opReorderProjects)
	if err != nil {
		return nil, err
	}
	if _, err := s.authz.RequireLeadOf(ctx, CanManageAllPrograms, program.LeadID); err != nil {
		return nil, err
	}
	projects, err := s.projects.ListByProgram(ctx, program.ID)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgProjectListFailed, err).WithOperation(opReorderProjects).WithResource(resourceProject)
	}
	priorities, ok := orderPriorities(projects, projectIDs)
	if !ok {
		return nil, errors.NewAppError(constants.ProjectOrderMismatch, constants.MsgProjectOrderMismatch, errors.ErrorTypeValidation,
			http.StatusUnprocessableEntity).WithOperation(opReorderProjects).WithResource(resourceProject)
	}
	if err := s.tx.Do(ctx, func(ctx context.Context) error { return s.projects.SetPriorities(ctx, priorities) }); err != nil {
		return nil, errors.DatabaseError(constants.MsgProjectSaveFailed, err).WithOperation(opReorderProjects).WithResource(resourceProject)
	}
	return s.ListByProgram(ctx, program.ID)
}

// orderPriorities maps each id to its place in the order, from 1. It is false unless the order
// lists each of the projects exactly once.
func orderPriorities(projects []models.Project, order []string) (map[string]int, bool) {
	if len(order) != len(projects) {
		return nil, false
	}
	known := make(map[string]bool, len(projects))
	for _, p := range projects {
		known[p.ID] = true
	}
	priorities := make(map[string]int, len(order))
	for i, id := range order {
		if _, seen := priorities[id]; seen || !known[id] {
			return nil, false
		}
		priorities[id] = i + 1
	}
	return priorities, true
}

func (s *projectService) Update(ctx context.Context, id string, in UpdateProjectInput) (*models.Project, error) {
	project, err := s.find(ctx, id, opUpdateProject)
	if err != nil {
		return nil, err
	}
	principal, err := s.requireEditor(ctx, project)
	if err != nil {
		return nil, err
	}

	oldDeadline := project.Deadline
	applyProjectChanges(project, in)
	if in.LeadID.Set {
		if !CanManageAllPrograms(principal) {
			return nil, errors.ForbiddenError(constants.MsgInsufficientPermissions, nil).
				WithOperation(opUpdateProject).WithResource(resourceProject)
		}
		if project.LeadID, err = checkLeadExists(ctx, s.people, in.LeadID.Value, opUpdateProject); err != nil {
			return nil, err
		}
	}
	if err := checkProjectDates(project.StartEarliest, project.Deadline, opUpdateProject); err != nil {
		return nil, err
	}
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.saveProject(ctx, project, opUpdateProject); err != nil {
			return err
		}
		if project.Deadline.Equal(oldDeadline) {
			return nil
		}
		return s.events.Record(ctx, ChangeInput{Type: constants.EventProjectSlipped, ProjectIDs: []string{project.ID},
			Payload: DeadlineChange{ProjectID: project.ID, Project: project.Name, OldDeadline: oldDeadline.Format(constants.DateLayout),
				NewDeadline: project.Deadline.Format(constants.DateLayout)}})
	})
	if err != nil {
		return nil, asAppError(err, constants.MsgProjectSaveFailed, opUpdateProject, resourceProject)
	}
	return s.find(ctx, id, opUpdateProject)
}

func (s *projectService) ApplyAction(ctx context.Context, id, action string) (*models.Project, error) {
	project, err := s.find(ctx, id, opProjectAction)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireEditor(ctx, project); err != nil {
		return nil, err
	}
	next, err := nextStatus(projectTransitions, project.Status, action, opProjectAction, resourceProject)
	if err != nil {
		return nil, err
	}
	previous := project.Status
	project.Status = next
	// The status decides whether the project's hours are load, so the change is an event.
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.saveProject(ctx, project, opProjectAction); err != nil {
			return err
		}
		return s.events.Record(ctx, ChangeInput{Type: constants.EventProjectStatusChanged, ProjectIDs: []string{project.ID},
			Payload: StatusChange{Kind: constants.StatusKindProject, ID: project.ID, Name: project.Name, OldStatus: previous, NewStatus: next}})
	})
	if err != nil {
		return nil, asAppError(err, constants.MsgProjectSaveFailed, opProjectAction, resourceProject)
	}
	return s.find(ctx, id, opProjectAction)
}

func (s *projectService) ReplacePhases(ctx context.Context, projectID string, phases []PhaseInput) (*models.Project, error) {
	project, err := s.find(ctx, projectID, opReplacePhases)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireEditor(ctx, project); err != nil {
		return nil, err
	}

	existing := map[string]bool{}
	for _, phase := range project.Phases {
		existing[phase.ID] = true
	}
	kept := map[string]bool{}
	rows := make([]models.ProjectPhase, 0, len(phases))
	for i, in := range phases {
		if in.EndWeek.Before(in.StartWeek) {
			return nil, phaseWeeksInvalid(opReplacePhases)
		}
		row := models.ProjectPhase{BaseModel: models.NewBaseModel(), ProjectID: projectID, Name: in.Name,
			Sequence: i + firstPhaseSequence, StartWeek: in.StartWeek, EndWeek: in.EndWeek}
		if in.ID != nil {
			if !existing[*in.ID] {
				return nil, phaseNotInProject(opReplacePhases)
			}
			row.ID = *in.ID
			kept[*in.ID] = true
		}
		rows = append(rows, row)
	}
	removed := missingIDs(existing, kept)

	err = s.tx.Do(ctx, func(ctx context.Context) error {
		inUse, err := s.projects.CountDemandsInPhases(ctx, removed)
		if err != nil {
			return err
		}
		if inUse > 0 {
			return errors.NewAppError(constants.PhaseInUse, constants.MsgPhaseInUse, errors.ErrorTypeConflict, http.StatusConflict).
				WithOperation(opReplacePhases).WithResource(resourcePhase)
		}
		if err := s.projects.DeletePhases(ctx, removed); err != nil {
			return err
		}
		return s.projects.SavePhases(ctx, rows)
	})
	if err != nil {
		return nil, asAppError(err, constants.MsgProjectSaveFailed, opReplacePhases, resourcePhase)
	}
	return s.find(ctx, projectID, opReplacePhases)
}

func (s *projectService) UpdatePhase(ctx context.Context, phaseID string, in UpdatePhaseInput) ([]models.ProjectPhase, error) {
	phase, err := s.projects.FindPhase(ctx, phaseID)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.PhaseNotFound, constants.MsgPhaseNotFound, resourcePhase, opUpdatePhase)
	}
	project, err := s.find(ctx, phase.ProjectID, opUpdatePhase)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireEditor(ctx, project); err != nil {
		return nil, err
	}

	oldEnd := phase.EndWeek
	if in.Name != nil {
		phase.Name = *in.Name
	}
	if in.StartWeek != nil {
		phase.StartWeek = *in.StartWeek
	}
	if in.EndWeek != nil {
		phase.EndWeek = *in.EndWeek
	}
	if phase.EndWeek.Before(phase.StartWeek) {
		return nil, phaseWeeksInvalid(opUpdatePhase)
	}

	changed := []models.ProjectPhase{*phase}
	slip := isoweek.WeeksBetween(oldEnd, phase.EndWeek)
	if slip > 0 {
		for _, later := range project.Phases {
			if later.Sequence <= phase.Sequence {
				continue
			}
			later.StartWeek = later.StartWeek.AddWeeks(slip)
			later.EndWeek = later.EndWeek.AddWeeks(slip)
			changed = append(changed, later)
		}
	}

	save := func(ctx context.Context) error {
		if err := s.projects.SavePhases(ctx, changed); err != nil {
			return err
		}
		if slip <= 0 {
			return nil
		}
		return s.events.Record(ctx, ChangeInput{Type: constants.EventPhaseSlipped, ProjectIDs: []string{project.ID},
			Payload: PhaseChange{ProjectID: project.ID, Project: project.Name, PhaseID: phase.ID, Phase: phase.Name,
				OldEndWeek: oldEnd.String(), NewEndWeek: phase.EndWeek.String(), Weeks: slip, MovedPhases: len(changed) - 1}})
	}
	if err := s.tx.Do(ctx, save); err != nil {
		return nil, asAppError(err, constants.MsgProjectSaveFailed, opUpdatePhase, resourcePhase)
	}
	return changed, nil
}

func (s *projectService) ReplaceDemands(ctx context.Context, projectID string, demands []DemandInput) (*models.Project, error) {
	project, err := s.find(ctx, projectID, opReplaceDemands)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireEditor(ctx, project); err != nil {
		return nil, err
	}

	phases := map[string]bool{}
	for _, phase := range project.Phases {
		phases[phase.ID] = true
	}
	existing := map[string]bool{}
	for _, demand := range project.Demands {
		existing[demand.ID] = true
	}
	kept := map[string]bool{}
	expertiseIDs := make([]string, 0, len(demands))
	rows := make([]models.Demand, 0, len(demands))
	for _, in := range demands {
		if in.PhaseID != nil && !phases[*in.PhaseID] {
			return nil, phaseNotInProject(opReplaceDemands)
		}
		row := models.Demand{BaseModel: models.NewBaseModel(), ProjectID: projectID, ExpertiseID: in.ExpertiseID,
			PhaseID: in.PhaseID, MinLevel: in.MinLevel, Headcount: in.Headcount, TotalHours: in.TotalHours, Profile: in.Profile}
		if row.Profile == "" {
			row.Profile = constants.DemandProfileFlat
		}
		if in.ID != nil {
			if !existing[*in.ID] {
				return nil, errors.NewAppError(constants.DemandNotInProject, constants.MsgDemandNotInProject, errors.ErrorTypeValidation, http.StatusBadRequest).
					WithOperation(opReplaceDemands).WithResource(resourceDemand)
			}
			row.ID = *in.ID
			kept[*in.ID] = true
		}
		expertiseIDs = append(expertiseIDs, in.ExpertiseID)
		rows = append(rows, row)
	}
	known, err := s.people.ExpertisesExist(ctx, expertiseIDs)
	if err != nil {
		return nil, err
	}
	if !known {
		return nil, errors.NewAppError(constants.ExpertiseUnknown, constants.MsgExpertiseUnknown, errors.ErrorTypeValidation, http.StatusBadRequest).
			WithOperation(opReplaceDemands).WithResource(resourceDemand)
	}

	removed := missingIDs(existing, kept)
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.projects.DeleteDemands(ctx, removed); err != nil {
			return err
		}
		return s.projects.SaveDemands(ctx, rows)
	})
	if err != nil {
		return nil, asAppError(err, constants.MsgProjectSaveFailed, opReplaceDemands, resourceDemand)
	}
	return s.find(ctx, projectID, opReplaceDemands)
}

// requireEditor passes planners and the lead of the project or of its program.
func (s *projectService) requireEditor(ctx context.Context, project *models.Project) (request.Principal, error) {
	var programLead *string
	if project.Program != nil {
		programLead = project.Program.LeadID
	}
	return s.authz.RequireLeadOf(ctx, CanManageAllPrograms, project.LeadID, programLead)
}

func (s *projectService) find(ctx context.Context, id, op string) (*models.Project, error) {
	project, err := s.projects.FindDetail(ctx, id)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.ProjectNotFound, constants.MsgProjectNotFound, resourceProject, op)
	}
	return project, nil
}

func (s *projectService) findProgram(ctx context.Context, id, op string) (*models.Program, error) {
	program, err := s.programs.FindWithLead(ctx, id)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.ProgramNotFound, constants.MsgProgramNotFound, resourceProgram, op)
	}
	return program, nil
}

// saveProject stores the project's own columns, never its relations.
func (s *projectService) saveProject(ctx context.Context, project *models.Project, op string) error {
	row := *project
	row.Program, row.Lead, row.Phases, row.Demands = nil, nil, nil, nil
	if err := s.projects.Save(ctx, &row); err != nil {
		return errors.DatabaseError(constants.MsgProjectSaveFailed, err).WithOperation(op).WithResource(resourceProject)
	}
	return nil
}

func applyProjectChanges(project *models.Project, in UpdateProjectInput) {
	if in.Name != nil {
		project.Name = *in.Name
	}
	if in.Priority != nil {
		project.Priority = *in.Priority
	}
	if in.StartEarliest != nil {
		project.StartEarliest = *in.StartEarliest
	}
	if in.Deadline != nil {
		project.Deadline = *in.Deadline
	}
	if in.ScopePoints != nil {
		project.ScopePoints = *in.ScopePoints
	}
	if in.DonePoints != nil {
		project.DonePoints = *in.DonePoints
	}
}

func checkProjectDates(start, deadline time.Time, op string) error {
	if deadline.Before(start) {
		return errors.NewAppError(constants.ProjectDatesInvalid, constants.MsgProjectDatesInvalid, errors.ErrorTypeValidation, http.StatusBadRequest).
			WithOperation(op).WithResource(resourceProject)
	}
	return nil
}

func phaseWeeksInvalid(op string) error {
	return errors.NewAppError(constants.PhaseWeeksInvalid, constants.MsgPhaseWeeksInvalid, errors.ErrorTypeValidation, http.StatusBadRequest).
		WithOperation(op).WithResource(resourcePhase)
}

func phaseNotInProject(op string) error {
	return errors.NewAppError(constants.PhaseNotInProject, constants.MsgPhaseNotInProject, errors.ErrorTypeValidation, http.StatusBadRequest).
		WithOperation(op).WithResource(resourcePhase)
}

// missingIDs lists the ids in all that are not in kept.
func missingIDs(all, kept map[string]bool) []string {
	var missing []string
	for id := range all {
		if !kept[id] {
			missing = append(missing, id)
		}
	}
	slices.Sort(missing)
	return missing
}

// asAppError keeps an AppError raised inside a transaction, and wraps anything else as a database error.
func asAppError(err error, message, op, resource string) error {
	if appErr := errors.GetAppError(err); appErr != nil {
		return appErr
	}
	return errors.DatabaseError(message, err).WithOperation(op).WithResource(resource)
}
