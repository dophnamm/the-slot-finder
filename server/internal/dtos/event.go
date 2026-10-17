package dtos

import (
	"strings"
	"time"

	"the-slot-finder/server/internal/models"
)

// ListEventsRequest pages GET /events, optionally for one program.
type ListEventsRequest struct {
	PageableRequest
	ProgramID string `json:"program_id" query:"program_id" validate:"omitempty,uuid"`
}

// Normalize applies the paging defaults and trims the program id.
func (r *ListEventsRequest) Normalize() {
	r.PageableRequest.Normalize()
	r.ProgramID = strings.TrimSpace(r.ProgramID)
}

// EventDraftResponse is the re-plan draft an event led to.
type EventDraftResponse struct {
	ID     string `json:"id"`
	Status string `json:"status" enums:"draft,published,rejected,superseded"`
}

// ChangeEventResponse is one change in the history with its processing status.
type ChangeEventResponse struct {
	ID   string `json:"id"`
	Type string `json:"type" enums:"AbsenceAdded,AbsenceRemoved,PhaseSlipped,ProjectSlipped,ProjectStatusChanged,DemandChanged,PersonCapacityChanged,TeamChanged,SkillChanged,ProgramSlipped"`
	// Payload depends on the type, e.g. person_name, from_date and to_date for absence.
	Payload    map[string]any      `json:"payload"`
	OccurredAt time.Time           `json:"occurred_at"`
	Status     string              `json:"status" enums:"pending,no_impact,draft_created,failed"`
	Attempts   int                 `json:"attempts"`
	LastError  *string             `json:"last_error"`
	ProjectIDs []string            `json:"project_ids"`
	Draft      *EventDraftResponse `json:"draft"`
}

// NewChangeEventResponses maps events; it never returns nil.
func NewChangeEventResponses(events []models.ChangeEvent) []ChangeEventResponse {
	out := make([]ChangeEventResponse, 0, len(events))
	for _, e := range events {
		out = append(out, NewChangeEventResponse(e))
	}
	return out
}

// NewChangeEventResponse maps one event with its draft loaded when it has one.
func NewChangeEventResponse(e models.ChangeEvent) ChangeEventResponse {
	item := ChangeEventResponse{ID: e.ID, Type: e.Type, Payload: e.Payload, OccurredAt: e.OccurredAt, Status: e.Status,
		Attempts: e.Attempts, LastError: e.LastError, ProjectIDs: nonNilStrings(e.ProjectIDs)}
	if item.Payload == nil {
		item.Payload = map[string]any{}
	}
	if e.Draft != nil {
		item.Draft = &EventDraftResponse{ID: e.Draft.ID, Status: e.Draft.Status}
	}
	return item
}
