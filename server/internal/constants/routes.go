package constants

// HTTP route paths. Feature groups add their own path constants here, e.g. PathPeople = "/people".
const (
	PathAPIBase        = "/api/v1"
	PathSwagger        = "/swagger/*"
	PathHealth         = "/"
	PathHealthDatabase = "/health/database"
	PathHealthMetrics  = "/health/metrics"
)
