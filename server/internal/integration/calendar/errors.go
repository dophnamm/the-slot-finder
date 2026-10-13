package calendar

import "errors"

// Errors a fetch or a parse can return. They never hold the link, which is a secret.
var (
	ErrHostNotAllowed = errors.New("calendar link host is not allowed")
	ErrUnreachable    = errors.New("calendar link could not be fetched")
	ErrNotICS         = errors.New("calendar link is not an ICS calendar")
	ErrTooLarge       = errors.New("calendar is larger than the limit")
)
