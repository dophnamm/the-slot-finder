package constants

// Custom validation tags, registered in utils.RegisterValidations.
const (
	// ValidationTagRole accepts only the values in Roles.
	ValidationTagRole = "role"
	// ValidationTagDate accepts a calendar date in DateLayout, e.g. 2026-10-12.
	ValidationTagDate = "isodate"
)
