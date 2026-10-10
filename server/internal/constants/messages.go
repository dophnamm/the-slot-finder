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
)

// Request limit messages.
const (
	MsgRateLimitExceeded       = "Rate limit exceeded"
	MsgClientIdentifierMissing = "Could not identify the client"
	MsgPayloadTooLarge         = "Request body is too large"
)
