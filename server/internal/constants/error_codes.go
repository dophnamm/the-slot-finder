package constants

// Custom error codes for the application
// Using text-based error codes for better readability and maintainability

const (
	// Success codes
	Success = "SUCCESS"

	// Rate Limit errors
	RateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	PayloadTooLarge   = "PAYLOAD_TOO_LARGE"

	// People and skills errors
	PersonNotFound           = "PERSON_NOT_FOUND"
	PersonEmailAlreadyExists = "PERSON_EMAIL_ALREADY_EXISTS"
	PersonDatesInvalid       = "PERSON_DATES_INVALID"
	PersonSkillsInvalid      = "PERSON_SKILLS_INVALID"

	// Absence errors
	AbsenceNotFound     = "ABSENCE_NOT_FOUND"
	AbsenceDatesInvalid = "ABSENCE_DATES_INVALID"
	// AbsenceNotPlanned: only a person's planned absence can be approved (not an approved one, not a holiday).
	AbsenceNotPlanned = "ABSENCE_NOT_PLANNED"

	// Calendar feed errors. The link is a secret, so a problem is only ever stored as one of these.
	CalendarSyncDisabled       = "CALENDAR_SYNC_DISABLED"
	CalendarFeedNotFound       = "CALENDAR_FEED_NOT_FOUND"
	CalendarFeedHostNotAllowed = "CALENDAR_FEED_HOST_NOT_ALLOWED"
	CalendarFeedUnreachable    = "CALENDAR_FEED_UNREACHABLE"
	CalendarFeedNotICS         = "CALENDAR_FEED_NOT_ICS"
	CalendarFeedTooLarge       = "CALENDAR_FEED_TOO_LARGE"

	// Program and project errors
	ProgramNotFound        = "PROGRAM_NOT_FOUND"
	ProgramDatesInvalid    = "PROGRAM_DATES_INVALID"
	InvalidStateTransition = "INVALID_STATE_TRANSITION"
	UnknownAction          = "UNKNOWN_ACTION"
	NoPersonRecord         = "NO_PERSON_RECORD"
	LeadNotFound           = "LEAD_NOT_FOUND"
	ProjectNotFound        = "PROJECT_NOT_FOUND"
	ProjectDatesInvalid    = "PROJECT_DATES_INVALID"
	ProgramNotOpen         = "PROGRAM_NOT_OPEN"
	PhaseNotFound          = "PHASE_NOT_FOUND"
	PhaseWeeksInvalid      = "PHASE_WEEKS_INVALID"
	PhaseInUse             = "PHASE_IN_USE"
	PhaseNotInProject      = "PHASE_NOT_IN_PROJECT"
	DemandNotInProject     = "DEMAND_NOT_IN_PROJECT"
	ExpertiseUnknown       = "EXPERTISE_UNKNOWN"
	WeekRangeInvalid       = "WEEK_RANGE_INVALID"
	PlanNotFound           = "PLAN_NOT_FOUND"
	ProjectNotPlannable    = "PROJECT_NOT_PLANNABLE"
	NoDemandToStaff        = "NO_DEMAND_TO_STAFF"
	DraftNotEditable       = "DRAFT_NOT_EDITABLE"
	PlanVersionConflict    = "PLAN_VERSION_CONFLICT"
	StretchNotConfirmed    = "STRETCH_NOT_CONFIRMED"
	StretchStaysLocked     = "STRETCH_STAYS_LOCKED"
	AssignmentNotFound     = "ASSIGNMENT_NOT_FOUND"
	AssignmentExists       = "ASSIGNMENT_EXISTS"
	DemandNotInDraft       = "DEMAND_NOT_IN_DRAFT"

	// Program team errors
	TeamMemberExists   = "TEAM_MEMBER_EXISTS"
	TeamMemberNotFound = "TEAM_MEMBER_NOT_FOUND"
	PersonNotInTeam    = "PERSON_NOT_IN_TEAM"
	WeekOutsideWindow  = "WEEK_OUTSIDE_WINDOW"
	WeekRangeTooLong   = "WEEK_RANGE_TOO_LONG"
	DraftSuperseded    = "DRAFT_SUPERSEDED"
	ScenarioNotFound   = "SCENARIO_NOT_FOUND"
	AssumptionInvalid  = "ASSUMPTION_INVALID"

	// Timesheet errors
	TimesheetInvalid      = "TIMESHEET_INVALID"
	TimesheetTooLarge     = "TIMESHEET_TOO_LARGE"
	TimesheetWeekInFuture = "TIMESHEET_WEEK_IN_FUTURE"
	TimesheetFileMissing  = "TIMESHEET_FILE_MISSING"

	// Email errors
	EmailSendError = "EMAIL_SEND_ERROR"

	// Legacy error codes - used by the new error handling system
	// These are kept for backward compatibility with the new AppError system
	ValidationError      = "VALIDATION_ERROR"
	UserNotFound         = "USER_NOT_FOUND"
	Unauthorized         = "UNAUTHORIZED"
	Forbidden            = "FORBIDDEN"
	BadRequest           = "BAD_REQUEST"
	Conflict             = "CONFLICT"
	InternalError        = "INTERNAL_ERROR"
	DatabaseError        = "DATABASE_ERROR"
	ExternalServiceError = "EXTERNAL_SERVICE_ERROR"
)
