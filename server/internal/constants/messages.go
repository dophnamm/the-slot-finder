package constants

// Response messages sent in meta.message. Group them by module.
const (
	MsgServiceHealthy          = "Service is healthy"
	MsgDatabaseHealthy         = "Database is healthy"
	MsgDatabaseUnhealthy       = "Database is unhealthy"
	MsgDatabaseNotInitialized  = "Database not initialized"
	MsgDatabaseMetricsReturned = "Database metrics retrieved successfully"
)

// Authentication and authorization messages.
const (
	MsgAuthHeaderRequired         = "Authorization header required"
	MsgAuthHeaderInvalid          = "Invalid authorization header format"
	MsgTokenRequired              = "Token required"
	MsgTokenInvalid               = "Invalid token"
	MsgTokenInactive              = "Token is not active"
	MsgTokenClaimsInvalid         = "Invalid token claims"
	MsgUserNotAuthenticated       = "User not authenticated"
	MsgInsufficientPermissions    = "Insufficient permissions"
	MsgPermissionEvaluationFailed = "Permission evaluation failed"
	MsgRPTClaimsInvalid           = "Invalid RPT claims"
	MsgCurrentUserRetrieved       = "Current user retrieved"
)

// People and skills messages.
const (
	MsgPeopleListed          = "People retrieved"
	MsgPersonCreated         = "Person created"
	MsgPersonUpdated         = "Person updated"
	MsgPersonSkillsUpdated   = "Skills updated"
	MsgExpertisesListed      = "Expertises retrieved"
	MsgPersonNotFound        = "Person not found"
	MsgPersonEmailTaken      = "A person with this email already exists"
	MsgPersonListFailed      = "Could not list people"
	MsgPersonCreateFailed    = "Could not create the person"
	MsgPersonUpdateFailed    = "Could not update the person"
	MsgPersonDatesInvalid    = "end_date must not be before start_date"
	MsgSkillsPrimaryRequired = "Exactly one skill must be primary"
	MsgSkillsDuplicate       = "Each expertise can appear only once"
	MsgSkillsUnknown         = "Unknown expertise"
	MsgExpertiseListFailed   = "Could not list expertises"
	MsgInvalidRequest        = "Request is invalid"
	MsgInvalidID             = "id must be a UUID"
	MsgDateInvalid           = "must be a date like 2026-10-12"
	MsgDatabaseReadFailed    = "Could not read from the database"
)

// Absence and holiday messages.
const (
	MsgAbsencesListed      = "Absences retrieved"
	MsgAbsenceCreated      = "Absence created"
	MsgAbsenceDeleted      = "Absence deleted"
	MsgHolidaysListed      = "Holidays retrieved"
	MsgHolidayCreated      = "Holiday created"
	MsgAbsenceNotFound     = "Absence not found"
	MsgAbsenceDatesInvalid = "to_date must not be before from_date"
	MsgAbsenceListFailed   = "Could not list absences"
	MsgAbsenceCreateFailed = "Could not create the absence"
	MsgAbsenceDeleteFailed = "Could not delete the absence"
)

// Program and project messages.
const (
	MsgProgramsListed         = "Programs retrieved"
	MsgProgramRetrieved       = "Program retrieved"
	MsgProgramCreated         = "Program created"
	MsgProgramUpdated         = "Program updated"
	MsgProgramNotFound        = "Program not found"
	MsgProgramDatesInvalid    = "end_date must not be before start_date"
	MsgProgramSaveFailed      = "Could not save the program"
	MsgProgramListFailed      = "Could not list programs"
	MsgStateTransitionInvalid = "This action is not allowed in the current status"
	MsgActionUnknown          = "Unknown action"
	MsgNoPersonRecord         = "Your account has no person record; ask a planner to add you"
	MsgLeadNotFound           = "lead_id does not match a person"
	MsgMustUseAction          = "cannot be set here; use an action"
	MsgWeekInvalid            = "must be an ISO week like 2026-W42"
	MsgProjectsListed         = "Projects retrieved"
	MsgProjectRetrieved       = "Project retrieved"
	MsgProjectCreated         = "Project created"
	MsgProjectUpdated         = "Project updated"
	MsgProjectNotFound        = "Project not found"
	MsgProjectDatesInvalid    = "deadline must not be before start_earliest"
	MsgProjectSaveFailed      = "Could not save the project"
	MsgProjectListFailed      = "Could not list projects"
	MsgProgramNotOpen         = "Projects can be added only while the program is open or pending"
	MsgPhasesListed           = "Phases retrieved"
	MsgPhasesUpdated          = "Phases updated"
	MsgPhaseNotFound          = "Phase not found"
	MsgPhaseWeeksInvalid      = "end_week must not be before start_week"
	MsgPhaseInUse             = "A phase that still has demand cannot be removed; move or delete its demand first"
	MsgPhaseNotInProject      = "The phase does not belong to this project"
	MsgDemandsUpdated         = "Demand updated"
	MsgDemandNotInProject     = "The demand does not belong to this project"
	MsgExpertiseUnknown       = "Unknown expertise"
	MsgPlanRetrieved          = "Plan retrieved"
	MsgPlanLoadFailed         = "Could not load the plan"
	MsgAssignmentsListed      = "Assignments retrieved"
	MsgWeekRangeInvalid       = "to must not be before from"
)

// Request limit messages.
const (
	MsgRateLimitExceeded       = "Rate limit exceeded"
	MsgClientIdentifierMissing = "Could not identify the client"
	MsgPayloadTooLarge         = "Request body is too large"
)
