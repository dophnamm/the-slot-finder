package handlers

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/services"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeHolidayImport records what it is asked to import.
type fakeHolidayImport struct {
	in   *services.ImportHolidaysInput
	file []byte
}

func (f *fakeHolidayImport) Import(_ context.Context, in services.ImportHolidaysInput) (*services.HolidayImportResult, error) {
	f.in = &in
	if in.File != nil {
		f.file, _ = io.ReadAll(in.File)
	}
	return &services.HolidayImportResult{Added: 2, Skipped: 1}, nil
}

// postHolidayForm posts a multipart form with these fields and, when given, an ICS file.
func postHolidayForm(t *testing.T, h *HolidayImportHandler, fields map[string]string, file []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	for name, value := range fields {
		require.NoError(t, form.WriteField(name, value))
	}
	if file != nil {
		part, err := form.CreateFormFile(constants.HolidayImportFileField, "holidays.ics")
		require.NoError(t, err)
		_, err = part.Write(file)
		require.NoError(t, err)
	}
	require.NoError(t, form.Close())
	e := echo.New()
	e.POST(constants.PathHolidayImport, h.Import)
	req := httptest.NewRequest(http.MethodPost, constants.PathHolidayImport, &body)
	req.Header.Set(echo.HeaderContentType, form.FormDataContentType())
	req = req.WithContext(plannerContext())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestHolidayImportHandler_Import(t *testing.T) {
	fields := map[string]string{"from_date": "2026-01-01", "to_date": "2027-12-31", "hours_per_day": "8"}

	t.Run("a file with its dates and hours", func(t *testing.T) {
		service := &fakeHolidayImport{}
		rec := postHolidayForm(t, ProvideHolidayImportHandler(service, newTestValidator(t)), fields, []byte("BEGIN:VCALENDAR"))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.NotNil(t, service.in)
		assert.Equal(t, "2026-01-01", service.in.From.Format(constants.DateLayout))
		assert.Equal(t, "2027-12-31", service.in.To.Format(constants.DateLayout))
		assert.InDelta(t, 8, service.in.HoursPerDay, 1e-9)
		assert.Equal(t, "BEGIN:VCALENDAR", string(service.file))
		assert.Contains(t, rec.Body.String(), `"added":2`)
	})

	t.Run("a link, trimmed, never echoed", func(t *testing.T) {
		service := &fakeHolidayImport{}
		withLink := map[string]string{"url": "  " + calendarLink + "  "}
		for k, v := range fields {
			withLink[k] = v
		}
		rec := postHolidayForm(t, ProvideHolidayImportHandler(service, newTestValidator(t)), withLink, nil)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, calendarLink, service.in.URL)
		assert.Nil(t, service.in.File)
		assert.NotContains(t, rec.Body.String(), "private-secret")
	})

	t.Run("dates are required", func(t *testing.T) {
		service := &fakeHolidayImport{}
		rec := postHolidayForm(t, ProvideHolidayImportHandler(service, newTestValidator(t)), map[string]string{"hours_per_day": "8"}, nil)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Nil(t, service.in)
	})
}
