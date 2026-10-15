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
)

// Change event statuses. The runner claims pending events and ends them as no_impact or
// draft_created; an event that keeps failing becomes failed.
const (
	EventStatusPending      = "pending"
	EventStatusNoImpact     = "no_impact"
	EventStatusDraftCreated = "draft_created"
	EventStatusFailed       = "failed"
)
