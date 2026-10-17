package handlers

import (
	"fmt"
	"net/http"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/labstack/echo/v5"
)

// PeopleImportHandler serves the people template download and the import of people from a CSV.
type PeopleImportHandler struct {
	BaseHandler
	imports services.PeopleImportService
}

// ProvidePeopleImportHandler creates the people import handler.
func ProvidePeopleImportHandler(imports services.PeopleImportService) *PeopleImportHandler {
	return &PeopleImportHandler{BaseHandler: *NewBaseHandler(), imports: imports}
}

// Template godoc
// @Summary People import template
// @Description A CSV with the header of the people file and one example row whose primary expertise is the first expertise in the catalogue.
// @Tags People
// @Produce text/csv
// @Security BearerAuth
// @Success 200 {file} file
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /people/import-template [get]
func (h *PeopleImportHandler) Template(c *echo.Context) error {
	data, err := h.imports.Template(c.Request().Context())
	if err != nil {
		return h.HandleError(c, err)
	}
	c.Response().Header().Set(echo.HeaderContentDisposition,
		fmt.Sprintf(constants.ContentDispositionAttachmentFormat, constants.PeopleImportFileName))
	return c.Blob(http.StatusOK, constants.CSVContentType, data)
}

// Import godoc
// @Summary Import people
// @Description Upload a filled-in template as the multipart field "file". Every row is checked first: any problem saves nothing and returns 400 PEOPLE_IMPORT_INVALID with data.problems [{line, column, code}]. A person is matched by email: a new one is created, an existing one is updated, and a row that changes nothing is counted as unchanged. A file over PEOPLE_IMPORT_MAX_BYTES or PEOPLE_IMPORT_MAX_ROWS gets 413 PEOPLE_IMPORT_TOO_LARGE.
// @Tags People
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Filled-in people template (CSV)"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.PeopleImportResultResponse}
// @Failure 400 {object} object{meta=dtos.Meta,data=object{problems=[]services.RowProblem}}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 413 {object} object{meta=dtos.Meta}
// @Failure 429 {object} object{meta=dtos.Meta,data=dtos.RateLimitResponse}
// @Router /people/import [post]
func (h *PeopleImportHandler) Import(c *echo.Context) error {
	header, err := c.FormFile(constants.PeopleImportFormField)
	if err != nil {
		return h.ValidationErrorResponse(c, constants.MsgPeopleImportFileMissing,
			map[string]string{constants.PeopleImportFormField: constants.MsgPeopleImportFileMissing})
	}
	file, err := header.Open()
	if err != nil {
		return h.HandleError(c, errors.ValidationError(constants.MsgPeopleImportUnreadable, err))
	}
	defer func() { _ = file.Close() }()
	result, err := h.imports.Import(c.Request().Context(), services.ImportInput{File: file, Size: header.Size})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgPeopleImported, dtos.NewPeopleImportResultResponse(*result), nil)
}
