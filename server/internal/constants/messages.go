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

// Request limit messages.
const (
	MsgRateLimitExceeded       = "Rate limit exceeded"
	MsgClientIdentifierMissing = "Could not identify the client"
	MsgPayloadTooLarge         = "Request body is too large"
)
