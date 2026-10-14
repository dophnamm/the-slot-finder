package dtos

import (
	"strings"

	"the-slot-finder/server/internal/models"
)

// HeatmapRequest is the query of GET /insights/heatmap. Weeks default to the current week
// and the HEATMAP_DEFAULT_WEEKS after it.
type HeatmapRequest struct {
	From      string `json:"from" query:"from" validate:"omitempty,isoweek"`
	To        string `json:"to" query:"to" validate:"omitempty,isoweek"`
	Expertise string `json:"expertise" query:"expertise" validate:"omitempty,uuid"`
}

// Normalize trims the values.
func (r *HeatmapRequest) Normalize() {
	r.From = strings.TrimSpace(r.From)
	r.To = strings.TrimSpace(r.To)
	r.Expertise = strings.TrimSpace(r.Expertise)
}

// ExpertiseHoursResponse is the part of a week's load on one expertise.
type ExpertiseHoursResponse struct {
	ExpertiseID string  `json:"expertise_id"`
	Expertise   string  `json:"expertise" example:"QA"`
	Hours       float64 `json:"hours" example:"8"`
	// Primary is true when it is the person's primary skill.
	Primary bool `json:"primary"`
}

// HeatmapCellResponse is one person in one week.
type HeatmapCellResponse struct {
	Week     string  `json:"week" example:"2026-W42"`
	Capacity float64 `json:"capacity" example:"36"`
	Load     float64 `json:"load" example:"32"`
	// Utilization is load ÷ capacity; null when capacity is 0.
	Utilization *float64 `json:"utilization" example:"0.89"`
	Band        string   `json:"band" enums:"unavailable,underused,healthy,near_full,overloaded"`
	// Tentative marks planned (not yet approved) absence in the week.
	Tentative   bool                     `json:"tentative"`
	ByExpertise []ExpertiseHoursResponse `json:"by_expertise"`
	// ActualHours is what the person worked that week from imported timesheets; null when
	// nothing was imported. Load, utilization and band stay planned.
	ActualHours *float64 `json:"actual_hours" example:"30"`
}

// HeatmapRowResponse is one person's weeks.
type HeatmapRowResponse struct {
	Person             PersonRef             `json:"person"`
	PrimaryExpertiseID string                `json:"primary_expertise_id"`
	PrimaryExpertise   string                `json:"primary_expertise" example:"Engineer"`
	Cells              []HeatmapCellResponse `json:"cells"`
}

// HeatmapResponse is the person × week matrix, rows grouped by primary expertise.
type HeatmapResponse struct {
	Weeks []string             `json:"weeks" example:"2026-W42,2026-W43"`
	Rows  []HeatmapRowResponse `json:"rows"`
}

// NewHeatmapResponse maps a heatmap.
func NewHeatmapResponse(h models.Heatmap) HeatmapResponse {
	weeks := make([]string, len(h.Weeks))
	for i, w := range h.Weeks {
		weeks[i] = w.String()
	}
	rows := make([]HeatmapRowResponse, 0, len(h.Rows))
	for _, r := range h.Rows {
		cells := make([]HeatmapCellResponse, 0, len(r.Cells))
		for _, c := range r.Cells {
			split := make([]ExpertiseHoursResponse, 0, len(c.ByExpertise))
			for _, e := range c.ByExpertise {
				split = append(split, ExpertiseHoursResponse{ExpertiseID: e.Expertise.ID, Expertise: e.Expertise.Name, Hours: e.Hours, Primary: e.Primary})
			}
			cells = append(cells, HeatmapCellResponse{Week: c.Week.String(), Capacity: c.Capacity, Load: c.Load,
				Utilization: c.Utilization, Band: c.Band, Tentative: c.Tentative, ByExpertise: split, ActualHours: c.ActualHours})
		}
		rows = append(rows, HeatmapRowResponse{Person: PersonRef{ID: r.Person.ID, Name: r.Person.Name},
			PrimaryExpertiseID: r.PrimaryExpertise.ID, PrimaryExpertise: r.PrimaryExpertise.Name, Cells: cells})
	}
	return HeatmapResponse{Weeks: weeks, Rows: rows}
}
