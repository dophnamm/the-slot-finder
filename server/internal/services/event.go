package services

import (
	"context"
	"encoding/json"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
)

const (
	opRecordEvent = "record_event"
	opListEvents  = "list_events"
	resourceEvent = "change_event"
)

// ChangeInput is a change to record as an event. Payload is a struct with json tags, such as
// AbsenceChange; it is stored as a JSON object for the change history.
type ChangeInput struct {
	Type       string
	Payload    any
	ProjectIDs []string
}

// AbsenceChange is the payload of AbsenceAdded and AbsenceRemoved. A holiday has no person.
type AbsenceChange struct {
	AbsenceID  string  `json:"absence_id"`
	PersonID   *string `json:"person_id"`
	PersonName string  `json:"person_name"`
	FromDate   string  `json:"from_date"`
	ToDate     string  `json:"to_date"`
	Type       string  `json:"absence_type"`
	Status     string  `json:"absence_status"`
}

// StatusChange is the payload of ProjectStatusChanged. Kind is "project" or "program"; a
// cancelled program records one event for all the projects it cancelled.
type StatusChange struct {
	Kind      string `json:"kind"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}

// DemandChange is the payload of DemandChanged.
type DemandChange struct {
	ProjectID string `json:"project_id"`
	Project   string `json:"project"`
}

// PersonCapacityChange is the payload of PersonCapacityChanged; Fields are the changed fields.
type PersonCapacityChange struct {
	PersonID   string   `json:"person_id"`
	PersonName string   `json:"person_name"`
	Fields     []string `json:"fields"`
}

// TeamChange is the payload of TeamChanged; Change is constants.TeamChangeAdded or Removed.
type TeamChange struct {
	ProgramID string `json:"program_id"`
	PersonID  string `json:"person_id"`
	Change    string `json:"change"`
}

// PhaseChange is the payload of PhaseSlipped.
type PhaseChange struct {
	ProjectID   string `json:"project_id"`
	Project     string `json:"project"`
	PhaseID     string `json:"phase_id"`
	Phase       string `json:"phase"`
	OldEndWeek  string `json:"old_end_week"`
	NewEndWeek  string `json:"new_end_week"`
	Weeks       int    `json:"weeks"`
	MovedPhases int    `json:"moved_phases"`
}

// DeadlineChange is the payload of ProjectSlipped.
type DeadlineChange struct {
	ProjectID   string `json:"project_id"`
	Project     string `json:"project"`
	OldDeadline string `json:"old_deadline"`
	NewDeadline string `json:"new_deadline"`
}

// EventRecorder records a change event. Call it inside the TxManager.Do that saves the change,
// so the event and the change commit or roll back together.
type EventRecorder interface {
	Record(ctx context.Context, in ChangeInput) error
}

// ListEventsInput pages the change history, optionally for one program.
type ListEventsInput struct {
	ProgramID string
	Page      int
	PageSize  int
}

// EventService records change events and lists the change history.
type EventService interface {
	EventRecorder
	// List returns events newest first; with a program, only events that touch its projects.
	// A manager sees only the events of their programs and their teams.
	List(ctx context.Context, in ListEventsInput) (*dtos.DataResponse[models.ChangeEvent], error)
}

type eventService struct {
	events   repositories.ChangeEventRepository
	projects repositories.ProjectRepository
	authz    AuthzService
	clock    clock.Clock
}

// ProvideEventService creates the event service.
func ProvideEventService(events repositories.ChangeEventRepository, projects repositories.ProjectRepository, authz AuthzService,
	clk clock.Clock,
) EventService {
	return &eventService{events: events, projects: projects, authz: authz, clock: clk}
}

// eventRecorder writes events to the outbox. It needs no permission check, so services the
// authorization depends on (people) can use it without a dependency cycle.
type eventRecorder struct {
	events repositories.ChangeEventRepository
	clock  clock.Clock
}

// ProvideEventRecorder gives services that save changes the recording part of the event service.
func ProvideEventRecorder(events repositories.ChangeEventRepository, clk clock.Clock) EventRecorder {
	return eventRecorder{events: events, clock: clk}
}

func (s *eventService) Record(ctx context.Context, in ChangeInput) error {
	return eventRecorder{events: s.events, clock: s.clock}.Record(ctx, in)
}

func (s eventRecorder) Record(ctx context.Context, in ChangeInput) error {
	payload, err := toPayload(in.Payload)
	if err != nil {
		return errors.InternalError(constants.MsgEventRecordFailed, err).WithOperation(opRecordEvent).WithResource(resourceEvent)
	}
	now := s.clock.Now().UTC()
	projectIDs := in.ProjectIDs
	if projectIDs == nil {
		projectIDs = []string{}
	}
	event := &models.ChangeEvent{BaseModel: models.NewBaseModel(), Type: in.Type, Payload: payload, OccurredAt: now,
		Status: constants.EventStatusPending, AvailableAt: now, ProjectIDs: projectIDs}
	if err := s.events.Create(ctx, event); err != nil {
		return errors.DatabaseError(constants.MsgEventRecordFailed, err).WithOperation(opRecordEvent).WithResource(resourceEvent)
	}
	return nil
}

func (s *eventService) List(ctx context.Context, in ListEventsInput) (*dtos.DataResponse[models.ChangeEvent], error) {
	if _, err := s.authz.Require(ctx, CanViewEvents); err != nil {
		return nil, err
	}
	reader, err := s.authz.ProgramReader(ctx)
	if err != nil {
		return nil, err
	}
	if !reader.All && reader.LeadID == "" {
		return emptyPage[models.ChangeEvent](&dtos.PageableRequest{Page: in.Page, PageSize: in.PageSize}), nil
	}
	var projectIDs []string
	if in.ProgramID != "" {
		projects, err := s.projects.ListByProgram(ctx, in.ProgramID)
		if err != nil {
			return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opListEvents).WithResource(resourceEvent)
		}
		projectIDs = make([]string, 0, len(projects))
		for _, p := range projects {
			projectIDs = append(projectIDs, p.ID)
		}
		if len(projectIDs) == 0 {
			return &dtos.DataResponse[models.ChangeEvent]{Data: []models.ChangeEvent{}, Pageable: &dtos.Pageable{Page: in.Page, PageSize: in.PageSize}}, nil
		}
	}
	page, err := s.events.List(ctx, projectIDs, readerLeadID(reader), &dtos.PageableRequest{Page: in.Page, PageSize: in.PageSize})
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opListEvents).WithResource(resourceEvent)
	}
	return page, nil
}

// toPayload turns a payload struct into the JSON object stored with the event.
func toPayload(v any) (map[string]any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
