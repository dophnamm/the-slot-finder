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
	authz services.AuthzService
}

// ProvideIdentityHandler creates the identity handler.
func ProvideIdentityHandler(authz services.AuthzService) *IdentityHandler {
	return &IdentityHandler{
		BaseHandler: *NewBaseHandler(),
		authz:       authz,
	}
}

// Me godoc
// @Summary Current user
// @Description The signed-in user, read from the verified access token: id, username, name, email and role.
// @Tags Identity
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CurrentUserResponse}
// @Failure 401 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /me [get]
func (h *IdentityHandler) Me(c *echo.Context) error {
	principal, err := h.authz.Require(c.Request().Context(), services.CanViewOwnProfile)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCurrentUserRetrieved, dtos.NewCurrentUserResponse(principal), nil)
}
