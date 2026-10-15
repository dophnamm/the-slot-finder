package constants

// Program statuses. They change only through program actions.
const (
	ProgramStatusOpen      = "open"
	ProgramStatusPending   = "pending"
	ProgramStatusClosing   = "closing"
	ProgramStatusClosed    = "closed"
	ProgramStatusCancelled = "cancelled"
)

// Program actions, sent as POST /programs/{id}/actions/{action}.
const (
	ProgramActionHold         = "hold"
	ProgramActionResume       = "resume"
	ProgramActionStartClosing = "start-closing"
	ProgramActionClose        = "close"
	ProgramActionCancel       = "cancel"
)

// Project statuses. planned and active are set by staffing; the others through project actions.
const (
	ProjectStatusIncoming  = "incoming"
	ProjectStatusPlanned   = "planned"
	ProjectStatusActive    = "active"
	ProjectStatusOnHold    = "on_hold"
	ProjectStatusDone      = "done"
	ProjectStatusCancelled = "cancelled"
	// StatusKindProject and StatusKindProgram say what a StatusChange event is about.
	StatusKindProject = "project"
	StatusKindProgram = "program"
)

// ProjectStatuses lists every project status, e.g. for the ?status= filter.
var ProjectStatuses = []string{
	ProjectStatusIncoming, ProjectStatusPlanned, ProjectStatusActive,
	ProjectStatusOnHold, ProjectStatusDone, ProjectStatusCancelled,
}

// Project actions, sent as POST /projects/{id}/actions/{action}.
const (
	ProjectActionHold     = "hold"
	ProjectActionResume   = "resume"
	ProjectActionComplete = "complete"
	ProjectActionCancel   = "cancel"
)

// Demand profiles. front and back are designed (README §5) and come later.
const (
	DemandProfileFlat = "flat"
)

// SupportedDemandProfiles are the profiles the engine can lay out today.
var SupportedDemandProfiles = []string{DemandProfileFlat}
