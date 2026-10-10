package dtos

import "the-slot-finder/server/internal/request"

// CurrentUserResponse is the signed-in user, as returned by GET /me.
type CurrentUserResponse struct {
	ID       string `json:"id" example:"5b0c2b7e-4c1e-4a43-9b8e-0b1f2f6f8a10"`
	Username string `json:"username" example:"planner"`
	Name     string `json:"name" example:"Pat Planner"`
	Email    string `json:"email" example:"planner@slot-finder.local"`
	Role     string `json:"role" example:"planner" enums:"planner,lead,specialist"`
}

// NewCurrentUserResponse maps the signed-in user to the response.
func NewCurrentUserResponse(principal request.Principal) CurrentUserResponse {
	return CurrentUserResponse{
		ID:       principal.Subject,
		Username: principal.Username,
		Name:     principal.Name,
		Email:    principal.Email,
		Role:     principal.Role,
	}
}
