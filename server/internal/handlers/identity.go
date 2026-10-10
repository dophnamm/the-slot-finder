package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/services"

	"github.com/labstack/echo/v5"
)

// IdentityHandler serves the signed-in user's own identity.
type IdentityHandler struct {
	BaseHandler
	authz  services.AuthzService
	people services.PeopleService
}

// ProvideIdentityHandler creates the identity handler.
func ProvideIdentityHandler(authz services.AuthzService, people services.PeopleService) *IdentityHandler {
	return &IdentityHandler{
		BaseHandler: *NewBaseHandler(),
		authz:       authz,
		people:      people,
	}
}

// Me godoc
// @Summary Current user
// @Description The signed-in user, read from the verified access token: id, username, name, email and role, plus person_id. On the first sign-in the person record is matched by email and linked to the account.
// @Tags Identity
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CurrentUserResponse}
// @Failure 401 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /me [get]
func (h *IdentityHandler) Me(c *echo.Context) error {
	ctx := c.Request().Context()
	principal, err := h.authz.Require(ctx, services.CanViewOwnProfile)
	if err != nil {
		return h.HandleError(c, err)
	}
	person, err := h.people.LinkPrincipal(ctx, principal)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCurrentUserRetrieved, dtos.NewCurrentUserResponse(principal, person), nil)
}
