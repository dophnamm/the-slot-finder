package constants

// Plan kinds.
const (
	PlanKindBaseline = "baseline"
	PlanKindDraft    = "draft"
)

// Plan statuses. A draft is published (it becomes the baseline) or rejected; a published
// baseline is superseded by the next one.
const (
	PlanStatusDraft      = "draft"
	PlanStatusPublished  = "published"
	PlanStatusRejected   = "rejected"
	PlanStatusSuperseded = "superseded"
)

// Assignment kinds; they match the engine's kinds.
const (
	AssignmentKindPrimary        = "primary"
	AssignmentKindCrossExpertise = "cross_expertise"
	AssignmentKindStretch        = "stretch"
)

// Plan change kinds in a draft's diff against the baseline.
const (
	PlanChangeAdded   = "added"
	PlanChangeRemoved = "removed"
	PlanChangeChanged = "changed"
)

// PlannableProjectStatuses are the project statuses a staffing draft may cover.
var PlannableProjectStatuses = []string{ProjectStatusIncoming, ProjectStatusPlanned, ProjectStatusActive}

// Audit actions and resources.
const (
	AuditActionPlanAccepted = "plan.accepted"
	AuditResourcePlan       = "plan"
)
