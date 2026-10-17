package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
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

const peopleCSV = "name,email,role,weekly_hours,non_project_hours,start_date,end_date,primary_expertise,primary_level,other_skills\n"

// fakePeopleImports implements services.PeopleImportService for handler tests.
type fakePeopleImports struct {
	template []byte
	uploaded string
	result   models.PeopleImportResult
	err      error
}

func (f *fakePeopleImports) Template(_ context.Context) ([]byte, error) { return f.template, f.err }

func (f *fakePeopleImports) Import(_ context.Context, in services.ImportInput) (*models.PeopleImportResult, error) {
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

func TestPeopleImportHandler_Template(t *testing.T) {
	fake := &fakePeopleImports{template: []byte(peopleCSV)}
	rec := serveTimesheet(t, http.MethodGet, constants.PathPeopleImportTemplate, nil, "", ProvidePeopleImportHandler(fake).Template)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, constants.CSVContentType, rec.Header().Get(echo.HeaderContentType))
	assert.Equal(t, `attachment; filename="people.csv"`, rec.Header().Get(echo.HeaderContentDisposition))
	assert.Equal(t, peopleCSV, rec.Body.String())
}

func TestPeopleImportHandler_Import(t *testing.T) {
	t.Run("imports the uploaded file", func(t *testing.T) {
		fake := &fakePeopleImports{result: models.PeopleImportResult{Created: 1, Updated: 2, Unchanged: 3}}
		body, contentType := multipartBody(t, constants.PeopleImportFormField, constants.PeopleImportFileName, peopleCSV)
		rec := serveTimesheet(t, http.MethodPost, constants.PathPeopleImport, body, contentType, ProvidePeopleImportHandler(fake).Import)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, peopleCSV, fake.uploaded)
		var envelope dtos.BaseResponse[dtos.PeopleImportResultResponse]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
		assert.Equal(t, dtos.PeopleImportResultResponse{Created: 1, Updated: 2, Unchanged: 3}, envelope.Data)
	})

	t.Run("no file", func(t *testing.T) {
		rec := serveTimesheet(t, http.MethodPost, constants.PathPeopleImport, bytes.NewBufferString("{}"), echo.MIMEApplicationJSON,
			ProvidePeopleImportHandler(&fakePeopleImports{}).Import)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		var envelope dtos.BaseResponse[json.RawMessage]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
		assert.Equal(t, constants.ValidationError, envelope.Meta.ErrorCode)
	})

	t.Run("row problems come back in the data", func(t *testing.T) {
		problems := []services.RowProblem{{Line: 3, Column: constants.PeopleImportColumnEmail, Code: constants.PeopleImportEmailInvalid}}
		fake := &fakePeopleImports{err: errors.NewAppError(constants.PeopleImportInvalid, constants.MsgPeopleImportInvalid,
			errors.ErrorTypeValidation, http.StatusBadRequest).WithContext(constants.PeopleImportProblemsKey, problems)}
		body, contentType := multipartBody(t, constants.PeopleImportFormField, constants.PeopleImportFileName, peopleCSV)
		rec := serveTimesheet(t, http.MethodPost, constants.PathPeopleImport, body, contentType, ProvidePeopleImportHandler(fake).Import)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		var envelope dtos.BaseResponse[map[string][]services.RowProblem]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
		assert.Equal(t, constants.PeopleImportInvalid, envelope.Meta.ErrorCode)
		assert.Equal(t, problems, envelope.Data[constants.PeopleImportProblemsKey])
	})
}
