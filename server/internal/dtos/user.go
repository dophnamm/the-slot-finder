package dtos

import "the-slot-finder/server/internal/constants"

// UserRequest holds the identity fields synced to Keycloak.
type UserRequest struct {
	Email      string `json:"email,omitempty" example:"john.doe@example.com" validate:"omitempty,email"`
	FirstName  string `json:"first_name,omitempty" example:"John" validate:"omitempty,min=2,max=100"`
	LastName   string `json:"last_name,omitempty" example:"Doe" validate:"omitempty,min=2,max=100"`
	KeycloakID string `json:"keycloak_id,omitempty" example:"123" validate:"omitempty,min=2,max=100"`
}

// CreateUserRequest is the payload for creating a Keycloak user.
type CreateUserRequest struct {
	UserRequest
}

// UpdateUserRequest is the payload for updating a Keycloak user.
type UpdateUserRequest struct {
	UserRequest
	Status constants.UserStatus `json:"status" example:"active" enums:"active,inactive"`
}
