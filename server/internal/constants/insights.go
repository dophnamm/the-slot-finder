package constants

// Utilization bands of a heatmap cell. The limits are settings; overloaded is over 100 %.
const (
	BandUnavailable = "unavailable"
	BandUnderused   = "underused"
	BandHealthy     = "healthy"
	BandNearFull    = "near_full"
	BandOverloaded  = "overloaded"
)
