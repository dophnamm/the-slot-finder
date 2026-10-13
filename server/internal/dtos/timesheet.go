package dtos

import (
	"strings"

	"the-slot-finder/server/internal/models"
)

// TimesheetTemplateRequest is the query of GET /timesheets/template. Both weeks are optional:
// the range defaults to the TIMESHEET_TEMPLATE_WEEKS up to the current week.
type TimesheetTemplateRequest struct {
	From string `json:"from" query:"from" validate:"omitempty,isoweek"`
	To   string `json:"to" query:"to" validate:"omitempty,isoweek"`
}

// Normalize trims the weeks.
func (r *TimesheetTemplateRequest) Normalize() {
	r.From = strings.TrimSpace(r.From)
	r.To = strings.TrimSpace(r.To)
}

// ImportResultResponse counts what a timesheet import did.
type ImportResultResponse struct {
	Created int `json:"created" example:"12"`
	Updated int `json:"updated" example:"3"`
	Skipped int `json:"skipped" example:"5"`
}

// NewImportResultResponse maps an import result.
func NewImportResultResponse(r models.ImportResult) ImportResultResponse {
	return ImportResultResponse{Created: r.Created, Updated: r.Updated, Skipped: r.Skipped}
}
