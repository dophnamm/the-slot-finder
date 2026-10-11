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

// Audit actions and resources.
const (
	AuditActionPlanAccepted = "plan.accepted"
	AuditResourcePlan       = "plan"
)
