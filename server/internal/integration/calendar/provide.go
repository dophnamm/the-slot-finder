package calendar

import "the-slot-finder/server/internal/config"

// ProvideFetcher builds the HTTP fetcher from the CALENDAR_* settings.
func ProvideFetcher(cfg *config.Config) Fetcher {
	return NewHTTPFetcher(FetchPolicy{AllowedHosts: cfg.CalendarAllowedHosts, Timeout: cfg.CalendarFetchTimeout,
		MaxBytes: cfg.CalendarMaxBytes})
}
