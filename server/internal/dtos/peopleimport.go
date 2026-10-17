package dtos

import "the-slot-finder/server/internal/models"

// PeopleImportResultResponse counts what a people import did.
type PeopleImportResultResponse struct {
	Created   int `json:"created" example:"12"`
	Updated   int `json:"updated" example:"3"`
	Unchanged int `json:"unchanged" example:"5"`
}

// NewPeopleImportResultResponse maps an import result.
func NewPeopleImportResultResponse(r models.PeopleImportResult) PeopleImportResultResponse {
	return PeopleImportResultResponse{Created: r.Created, Updated: r.Updated, Unchanged: r.Unchanged}
}
