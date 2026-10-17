package constants

// Change event types, recorded in the same transaction as the change (outbox).
const (
	EventAbsenceAdded   = "AbsenceAdded"
	EventAbsenceRemoved = "AbsenceRemoved"
	EventPhaseSlipped   = "PhaseSlipped"
	EventProjectSlipped = "ProjectSlipped"
	// EventProjectStatusChanged: a project or program changed status; hours of a project that
	// is no longer plannable free capacity.
	EventProjectStatusChanged = "ProjectStatusChanged"
	// EventDemandChanged: a project's demand was replaced with different demand.
	EventDemandChanged = "DemandChanged"
	// EventPersonCapacityChanged: a person's weekly hours, non-project hours, dates or active
	// flag changed, so their capacity moved.
	EventPersonCapacityChanged = "PersonCapacityChanged"
	// EventTeamChanged: a person joined or left a program's team.
	EventTeamChanged = "TeamChanged"
)

// Change event statuses. The runner claims pending events and ends them as no_impact or
// draft_created; an event that keeps failing becomes failed.
const (
	EventStatusPending      = "pending"
	EventStatusNoImpact     = "no_impact"
	EventStatusDraftCreated = "draft_created"
	EventStatusFailed       = "failed"
)

// EventNoteUnderstaffed is stored on a no_impact event (in last_error) when projects stay
// understaffed that the event did not open; the argument is how many.
const EventNoteUnderstaffed = "No re-plan: %d project(s) stay understaffed, and this change does not free capacity for them"
