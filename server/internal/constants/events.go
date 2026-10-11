package constants

// Change event types, recorded in the same transaction as the change (outbox).
const (
	EventAbsenceAdded   = "AbsenceAdded"
	EventAbsenceRemoved = "AbsenceRemoved"
	EventPhaseSlipped   = "PhaseSlipped"
	EventProjectSlipped = "ProjectSlipped"
)

// Change event statuses. The runner claims pending events and ends them as no_impact or
// draft_created; an event that keeps failing becomes failed.
const (
	EventStatusPending      = "pending"
	EventStatusNoImpact     = "no_impact"
	EventStatusDraftCreated = "draft_created"
	EventStatusFailed       = "failed"
)
