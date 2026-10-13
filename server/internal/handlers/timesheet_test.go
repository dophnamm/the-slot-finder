package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/services"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const timesheetCSV = "person_email,project_id,week,actual_hours\nada@example.test,0192f3a4-0000-7000-8000-00000000e101,2026-W41,8\n"

// fakeTimesheets implements services.TimesheetService for handler tests.
type fakeTimesheets struct {
	template   []byte
	templateIn services.TemplateInput
	uploaded   string
	result     models.ImportResult
	err        error
}

func (f *fakeTimesheets) Template(_ context.Context, in services.TemplateInput) ([]byte, error) {
	f.templateIn = in
	return f.template, f.err
}

func (f *fakeTimesheets) Import(_ context.Context, in services.ImportInput) (*models.ImportResult, error) {
	data, err := io.ReadAll(in.File)
	if err != nil {
		return nil, err
	}
	f.uploaded = string(data)
	if f.err != nil {
		return nil, f.err
	}
	return &f.result, nil
}

// serveTimesheet runs one timesheet handler with the planner signed in.
func serveTimesheet(t *testing.T, method, target string, body io.Reader, contentType string, handler echo.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	path, _, _ := strings.Cut(target, "?")
	e.Add(method, path, handler)
	req := httptest.NewRequest(method, target, body)
	if contentType != "" {
		req.Header.Set(echo.HeaderContentType, contentType)
	}
	req = req.WithContext(plannerContext())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// multipartBody is a form holding one file.
func multipartBody(t *testing.T, field, name, content string) (io.Reader, string) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile(field, name)
	require.NoError(t, err)
	_, err = part.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return &body, w.FormDataContentType()
}

func TestTimesheetHandler_Template(t *testing.T) {
	t.Run("downloads the CSV for the chosen weeks", func(t *testing.T) {
		fake := &fakeTimesheets{template: []byte(timesheetCSV)}
		h := ProvideTimesheetHandler(fake, newTestValidator(t))
		rec := serveTimesheet(t, http.MethodGet, constants.PathTimesheetTemplate+"?from=2026-W41&to=2026-W42", nil, "", h.Template)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, constants.TimesheetContentType, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, `attachment; filename="timesheet.csv"`, rec.Header().Get(echo.HeaderContentDisposition))
		assert.Equal(t, timesheetCSV, rec.Body.String())
		assert.Equal(t, "2026-W41", fake.templateIn.From.String())
		assert.Equal(t, "2026-W42", fake.templateIn.To.String())
	})

	t.Run("without weeks the service picks the defaults", func(t *testing.T) {
		fake := &fakeTimesheets{template: []byte(timesheetCSV)}
		rec := serveTimesheet(t, http.MethodGet, constants.PathTimesheetTemplate, nil, "",
			ProvideTimesheetHandler(fake, newTestValidator(t)).Template)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.True(t, fake.templateIn.From.IsZero())
		assert.True(t, fake.templateIn.To.IsZero())
	})

	t.Run("a week that does not exist", func(t *testing.T) {
		rec := serveTimesheet(t, http.MethodGet, constants.PathTimesheetTemplate+"?from=2026-W99", nil, "",
			ProvideTimesheetHandler(&fakeTimesheets{}, newTestValidator(t)).Template)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestTimesheetHandler_Import(t *testing.T) {
	t.Run("imports the uploaded file", func(t *testing.T) {
		fake := &fakeTimesheets{result: models.ImportResult{Created: 1, Updated: 2, Skipped: 3}}
		body, contentType := multipartBody(t, constants.TimesheetFormField, constants.TimesheetFileName, timesheetCSV)
		rec := serveTimesheet(t, http.MethodPost, constants.PathTimesheetImport, body, contentType,
			ProvideTimesheetHandler(fake, newTestValidator(t)).Import)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, timesheetCSV, fake.uploaded)
		var envelope dtos.BaseResponse[dtos.ImportResultResponse]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
		assert.Equal(t, dtos.ImportResultResponse{Created: 1, Updated: 2, Skipped: 3}, envelope.Data)
	})

	t.Run("no file", func(t *testing.T) {
		rec := serveTimesheet(t, http.MethodPost, constants.PathTimesheetImport, bytes.NewBufferString("{}"), echo.MIMEApplicationJSON,
			ProvideTimesheetHandler(&fakeTimesheets{}, newTestValidator(t)).Import)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		var envelope dtos.BaseResponse[json.RawMessage]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
		assert.Equal(t, constants.ValidationError, envelope.Meta.ErrorCode)
	})

	t.Run("row problems come back in the data", func(t *testing.T) {
		problems := []services.RowProblem{{Line: 5, Column: constants.TimesheetColumnPersonEmail, Code: constants.TimesheetPersonUnknown}}
		fake := &fakeTimesheets{err: errors.NewAppError(constants.TimesheetInvalid, constants.MsgTimesheetInvalid,
			errors.ErrorTypeValidation, http.StatusBadRequest).WithContext(constants.TimesheetProblemsKey, problems)}
		body, contentType := multipartBody(t, constants.TimesheetFormField, constants.TimesheetFileName, timesheetCSV)
		rec := serveTimesheet(t, http.MethodPost, constants.PathTimesheetImport, body, contentType,
			ProvideTimesheetHandler(fake, newTestValidator(t)).Import)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		var envelope dtos.BaseResponse[map[string][]services.RowProblem]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
		assert.Equal(t, constants.TimesheetInvalid, envelope.Meta.ErrorCode)
		assert.Equal(t, problems, envelope.Data[constants.TimesheetProblemsKey])
	})
}
