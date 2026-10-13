package dtos

import "the-slot-finder/server/internal/models"

// CreateAbsenceRequest is the body of POST /people/{id}/absences. Dates use the 2026-10-12
// format and are inclusive. status is honoured for planners; others always report planned.
type CreateAbsenceRequest struct {
	FromDate    string  `json:"from_date" validate:"required,isodate" example:"2026-10-19"`
	ToDate      string  `json:"to_date" validate:"required,isodate" example:"2026-10-23"`
	HoursPerDay float64 `json:"hours_per_day" validate:"gt=0,lte=24" example:"8"`
	Type        string  `json:"type" validate:"required,absence_type" enums:"leave,training"`
	Status      string  `json:"status" validate:"omitempty,absence_status" enums:"planned,approved"`
}

// CreateHolidayRequest is the body of POST /holidays.
type CreateHolidayRequest struct {
	FromDate    string  `json:"from_date" validate:"required,isodate" example:"2026-12-24"`
	ToDate      string  `json:"to_date" validate:"required,isodate" example:"2026-12-25"`
	HoursPerDay float64 `json:"hours_per_day" validate:"gt=0,lte=24" example:"8"`
}

// AbsenceResponse is an absence or a holiday (person_id null).
type AbsenceResponse struct {
	ID          string  `json:"id"`
	PersonID    *string `json:"person_id"`
	FromDate    string  `json:"from_date" example:"2026-10-19"`
	ToDate      string  `json:"to_date" example:"2026-10-23"`
	HoursPerDay float64 `json:"hours_per_day" example:"8"`
	Type        string  `json:"type" example:"leave"`
	Status      string  `json:"status" example:"approved"`
	Source      string  `json:"source" example:"manual" enums:"manual,calendar"`
}

// NewAbsenceResponse maps an absence model.
func NewAbsenceResponse(a models.Absence) AbsenceResponse {
	return AbsenceResponse{
		ID:          a.ID,
		PersonID:    a.PersonID,
		FromDate:    *FormatDate(&a.FromDate),
		ToDate:      *FormatDate(&a.ToDate),
		HoursPerDay: a.HoursPerDay,
		Type:        a.Type,
		Status:      a.Status,
		Source:      a.Source,
	}
}

// NewAbsenceResponses maps a list; it never returns nil.
func NewAbsenceResponses(absences []models.Absence) []AbsenceResponse {
	out := make([]AbsenceResponse, 0, len(absences))
	for _, a := range absences {
		out = append(out, NewAbsenceResponse(a))
	}
	return out
}
