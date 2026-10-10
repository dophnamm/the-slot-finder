package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/request"
	"the-slot-finder/server/internal/services"
	"the-slot-finder/server/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPersonID = "0192f3a4-0000-7000-8000-0000000000aa"

// validationKeyPrefix starts each key of a validation error response.
const validationKeyPrefix = "validation_"

// fakePeople implements only what each test needs; other methods panic if called.
type fakePeople struct {
	services.PeopleService
	created *services.CreatePersonInput
	updated *services.UpdatePersonInput
	err     error
}

func (f *fakePeople) Create(_ context.Context, in services.CreatePersonInput) (*models.Person, error) {
	f.created = &in
	if f.err != nil {
		return nil, f.err
	}
	return &models.Person{BaseModel: models.BaseModel{ID: testPersonID}, Name: in.Name, Email: in.Email}, nil
}

func (f *fakePeople) Update(_ context.Context, id string, in services.UpdatePersonInput) (*models.Person, error) {
	f.updated = &in
	return &models.Person{BaseModel: models.BaseModel{ID: id}}, f.err
}

func (f *fakePeople) List(_ context.Context, in services.ListPeopleInput) (*dtos.DataResponse[models.Person], error) {
	return &dtos.DataResponse[models.Person]{
		Data:     []models.Person{{Name: "Alex"}},
		Pageable: &dtos.Pageable{Page: in.Page, PageSize: in.PageSize, Total: 1},
	}, nil
}

func newTestValidator(t *testing.T) *validator.Validate {
	t.Helper()
	v := validator.New()
	require.NoError(t, utils.RegisterValidations(v))
	return v
}

func plannerContext() context.Context {
	return request.NewPrincipalContext(context.Background(),
		request.Principal{Subject: "kc-planner", Role: constants.RolePlanner, Roles: []string{constants.RolePlanner}})
}

// serve runs one handler with the planner signed in and a route path such as /people/:id.
func serve(t *testing.T, route, method, target, body string, handler echo.HandlerFunc) (*httptest.ResponseRecorder, dtos.BaseResponse[json.RawMessage]) {
	t.Helper()
	e := echo.New()
	e.Add(method, route, handler)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(plannerContext())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	var envelope dtos.BaseResponse[json.RawMessage]
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope), rec.Body.String())
	return rec, envelope
}

const validPersonBody = `{"name":" Alex ","email":"Alex@Example.test","role":"specialist","weekly_hours":40,
"start_date":"2026-10-12","skills":[{"expertise_id":"0192f3a4-0000-7000-8000-000000000001","level":4,"primary":true}]}`

func TestPersonHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
		wantFields []string
	}{
		{name: "created", body: validPersonBody, wantStatus: http.StatusCreated},
		{name: "invalid fields", wantStatus: http.StatusBadRequest,
			body:       `{"name":"","email":"not-an-email","role":"boss","weekly_hours":0,"start_date":"12/10/2026","skills":[]}`,
			wantFields: []string{"name", "email", "role", "weekly_hours", "start_date", "skills"}},
		{name: "email taken", body: validPersonBody, wantStatus: http.StatusConflict,
			serviceErr: errors.NewAppError(constants.PersonEmailAlreadyExists, constants.MsgPersonEmailTaken, errors.ErrorTypeConflict, http.StatusConflict)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			people := &fakePeople{err: tt.serviceErr}
			h := ProvidePersonHandler(people, services.ProvideAuthzService(), newTestValidator(t))
			rec, envelope := serve(t, constants.PathPeople, http.MethodPost, constants.PathPeople, tt.body, h.Create)

			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantStatus == http.StatusCreated {
				require.NotNil(t, people.created)
				assert.Equal(t, "Alex", people.created.Name)
				assert.Equal(t, "alex@example.test", people.created.Email)
				require.NotNil(t, people.created.StartDate)
				assert.Equal(t, "2026-10-12", people.created.StartDate.Format(constants.DateLayout))
			}
			if len(tt.wantFields) > 0 {
				var fields map[string]string
				require.NoError(t, json.Unmarshal(envelope.Data, &fields))
				// The shared validation response prefixes each JSON field name.
				for _, f := range tt.wantFields {
					assert.Contains(t, fields, validationKeyPrefix+f)
				}
			}
		})
	}
}

func TestPersonHandler_Update(t *testing.T) {
	t.Run("clears end_date and keeps other fields", func(t *testing.T) {
		people := &fakePeople{}
		h := ProvidePersonHandler(people, services.ProvideAuthzService(), newTestValidator(t))
		rec, _ := serve(t, constants.PathPerson, http.MethodPatch, "/people/"+testPersonID,
			`{"weekly_hours":32,"end_date":null}`, h.Update)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.NotNil(t, people.updated)
		assert.InDelta(t, 32, *people.updated.WeeklyHours, 1e-9)
		assert.True(t, people.updated.EndDate.Set)
		assert.Nil(t, people.updated.EndDate.Value)
		assert.False(t, people.updated.StartDate.Set)
		assert.Nil(t, people.updated.Name)
	})

	t.Run("bad id", func(t *testing.T) {
		h := ProvidePersonHandler(&fakePeople{}, services.ProvideAuthzService(), newTestValidator(t))
		rec, _ := serve(t, constants.PathPerson, http.MethodPatch, "/people/not-a-uuid", `{}`, h.Update)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("bad date", func(t *testing.T) {
		h := ProvidePersonHandler(&fakePeople{}, services.ProvideAuthzService(), newTestValidator(t))
		rec, envelope := serve(t, constants.PathPerson, http.MethodPatch, "/people/"+testPersonID, `{"start_date":"tomorrow"}`, h.Update)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, string(envelope.Data), "start_date")
	})
}

func TestPersonHandler_List(t *testing.T) {
	h := ProvidePersonHandler(&fakePeople{}, services.ProvideAuthzService(), newTestValidator(t))
	rec, envelope := serve(t, constants.PathPeople, http.MethodGet, constants.PathPeople+"?page=2&page_size=5", "", h.List)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 2, envelope.Meta.Page)
	assert.Equal(t, 5, envelope.Meta.PageSize)
	var people []dtos.PersonResponse
	require.NoError(t, json.Unmarshal(envelope.Data, &people))
	require.Len(t, people, 1)
	assert.Equal(t, "Alex", people[0].Name)
	assert.NotNil(t, people[0].Skills) // an empty list is [], never null
}
