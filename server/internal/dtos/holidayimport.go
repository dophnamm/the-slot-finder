package dtos

import "strings"

// ImportHolidaysRequest is the form of POST /holidays/import next to the optional ICS file: a
// link instead of the file, the dates to keep, and the hours each holiday takes off.
type ImportHolidaysRequest struct {
	URL         string  `form:"url" validate:"omitempty,url,max=2048"`
	FromDate    string  `form:"from_date" validate:"required,isodate" example:"2026-01-01"`
	ToDate      string  `form:"to_date" validate:"required,isodate" example:"2027-12-31"`
	HoursPerDay float64 `form:"hours_per_day" validate:"gt=0,lte=24" example:"8"`
}

// Normalize trims the link.
func (r *ImportHolidaysRequest) Normalize() { r.URL = strings.TrimSpace(r.URL) }

// HolidayImportResponse counts the holidays added and the ones that were there already.
type HolidayImportResponse struct {
	Added   int `json:"added" example:"11"`
	Skipped int `json:"skipped" example:"2"`
}
