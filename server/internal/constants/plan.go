package constants

// Plan kinds.
const (
	PlanKindBaseline = "baseline"
	PlanKindDraft    = "draft"
	// PlanKindScenario is a what-if: the baseline planned again under hypothetical changes.
	PlanKindScenario = "scenario"
)

// Scenario assumption types: a project's dates move, or a person is away.
const (
	AssumptionShiftProject = "shift_project"
	AssumptionAddAbsence   = "add_absence"
)

// AssumptionTypes lists the assumption types a scenario accepts.
var AssumptionTypes = []string{AssumptionShiftProject, AssumptionAddAbsence}

// MaxAbsenceHoursPerDay is the most hours of capacity a day of leave can take.
const MaxAbsenceHoursPerDay = 24

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

// Reasons the server records on assignments a Planner adds by hand.
const (
	ReasonAddedByHand          = "added by a Planner"
	ReasonStretchLevelFormat   = "stretch: %s level %d is below %d"
	ReasonStretchNoSkillFormat = "stretch: no %s skill"
)

// Audit actions and resources.
const (
	AuditActionPlanAccepted = "plan.accepted"
	AuditResourcePlan       = "plan"
)
