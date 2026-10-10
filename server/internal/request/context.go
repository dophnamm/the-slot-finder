package request

import (
	"context"
	"slices"

	"the-slot-finder/server/pkg/correlationid"
)

type ctxKey struct {
	name string
}

func (k ctxKey) String() string {
	return "slot-finder context value " + k.name
}

var ctxKeyLanguageCode = ctxKey{"language_code"}

// LanguageCodeFromContext retrieves the language code from the context.
func LanguageCodeFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ctxKeyLanguageCode).(string)
	return val, ok
}

// NewLanguageCodeContext creates a new context with the given language code.
func NewLanguageCodeContext(ctx context.Context, languageCode string) context.Context {
	return context.WithValue(ctx, ctxKeyLanguageCode, languageCode)
}

var ctxKeyRequestTimestamp = ctxKey{"request_timestamp"}

// RequestTimestampFromContext retrieves the request timestamp from the context.
// The timestamp is stored as Unix milliseconds.
func RequestTimestampFromContext(ctx context.Context) (int64, bool) {
	val, ok := ctx.Value(ctxKeyRequestTimestamp).(int64)
	return val, ok
}

// NewRequestTimestampContext creates a new context with the given request timestamp.
func NewRequestTimestampContext(ctx context.Context, requestTimestamp int64) context.Context {
	return context.WithValue(ctx, ctxKeyRequestTimestamp, requestTimestamp)
}

// CorrelationIDFromContext retrieves the correlation ID from the context.
func CorrelationIDFromContext(ctx context.Context) (string, bool) {
	return correlationid.FromContext(ctx)
}

// NewCorrelationIDContext creates a new context with the given correlation ID.
func NewCorrelationIDContext(ctx context.Context, correlationID string) context.Context {
	return correlationid.NewContext(ctx, correlationID)
}

var ctxKeyRequestURL = ctxKey{"request_url"}

// RequestURLFromContext retrieves the request URL from the context.
func RequestURLFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ctxKeyRequestURL).(string)
	return val, ok
}

// NewRequestURLContext creates a new context with the given request URL.
func NewRequestURLContext(ctx context.Context, requestURL string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestURL, requestURL)
}

// Principal is the signed-in user, taken from a verified access token.
// The zero value means "nobody", so code that reads it never needs a nil check.
type Principal struct {
	Subject  string
	Username string
	Name     string
	Email    string
	// Role is the most privileged of Roles that the application knows.
	Role  string
	Roles []string
}

// HasAnyRole reports whether the principal holds at least one of roles.
func (p Principal) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if slices.Contains(p.Roles, role) {
			return true
		}
	}
	return false
}

var ctxKeyPrincipal = ctxKey{"principal"}

// PrincipalFromContext retrieves the signed-in user. ok is false when there is none.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	if ctx == nil {
		return Principal{}, false
	}
	val, ok := ctx.Value(ctxKeyPrincipal).(Principal)
	return val, ok
}

// NewPrincipalContext creates a new context carrying the signed-in user.
func NewPrincipalContext(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, ctxKeyPrincipal, principal)
}
