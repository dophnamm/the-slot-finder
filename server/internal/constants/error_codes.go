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
