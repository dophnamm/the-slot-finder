package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/integration/llm"
	"the-slot-finder/server/internal/models"
)

// listProjects answers ListByStatus like the real service: only roles that read programs get through.
type listProjects struct {
	ProjectService
	authz    AuthzService
	projects []models.Project
}

func (l listProjects) ListByStatus(ctx context.Context, _ ListProjectsInput) (*dtos.DataResponse[models.Project], error) {
	if _, err := l.authz.ProgramReader(ctx); err != nil {
		return nil, err
	}
	return &dtos.DataResponse[models.Project]{Data: l.projects}, nil
}

type listPeople struct {
	PeopleService
	people []models.Person
}

func (l listPeople) List(context.Context, ListPeopleInput) (*dtos.DataResponse[models.Person], error) {
	return &dtos.DataResponse[models.Person]{Data: l.people}, nil
}

func namedProjects(names ...string) []models.Project {
	out := make([]models.Project, 0, len(names))
	for i, n := range names {
		out = append(out, models.Project{BaseModel: models.BaseModel{ID: fmt.Sprintf("p%d", i)}, Name: n,
			Deadline: time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC), Program: &models.Program{Name: "Acme"}})
	}
	return out
}

func namedPeople(names ...string) []models.Person {
	out := make([]models.Person, 0, len(names))
	for i, n := range names {
		out = append(out, models.Person{BaseModel: models.BaseModel{ID: fmt.Sprintf("u%d", i)}, Name: n,
			Skills: []models.PersonSkill{
				{IsPrimary: false, Expertise: models.Expertise{Name: "QA"}},
				{IsPrimary: true, Expertise: models.Expertise{Name: "Engineer"}},
			}})
	}
	return out
}

func runFind(t *testing.T, svc *aiService, ctx context.Context, tool, query string) (map[string]any, error) {
	t.Helper()
	data, err := svc.dispatch(ctx, llm.FunctionCall{Name: tool, Args: map[string]any{"query": query}})
	if err != nil {
		return nil, err
	}
	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))
	return out, nil
}

func TestAIService_FindProjects(t *testing.T) {
	many := namedProjects("Mobile 1", "Mobile 2", "Mobile 3", "Mobile 4", "Mobile 5", "Mobile 6", "Mobile 7")
	tests := []struct {
		name     string
		role     string
		projects []models.Project
		query    string
		want     []string
		wantErr  bool
	}{
		{name: "name match, case-insensitive", role: constants.RolePlanner, projects: namedProjects("Mobile App", "Web"),
			query: "mOBILE", want: []string{"Mobile App"}},
		{name: "no match gives an empty list, not an error", role: constants.RolePlanner, projects: namedProjects("Web"),
			query: "zzz"},
		{name: "at most 5 matches", role: constants.RolePlanner, projects: many, query: "mobile",
			want: []string{"Mobile 1", "Mobile 2", "Mobile 3", "Mobile 4", "Mobile 5"}},
		{name: "manager allowed", role: constants.RoleManager, projects: namedProjects("Mobile App"), query: "mobile",
			want: []string{"Mobile App"}},
		{name: "specialist refused", role: constants.RoleSpecialist, projects: namedProjects("Mobile App"),
			query: "mobile", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authz := ProvideAuthzService(leadPeople{})
			svc := &aiService{authz: authz, projects: listProjects{authz: authz, projects: tt.projects}}
			out, err := runFind(t, svc, asRole("kc-"+tt.role, tt.role), constants.AIToolFindProjects, tt.query)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			items := out["projects"].([]any)
			got := []string{}
			for _, it := range items {
				m := it.(map[string]any)
				got = append(got, m["name"].(string))
				assert.NotEmpty(t, m["id"])
				assert.Equal(t, "Acme", m["program"])
				assert.Equal(t, "2026-12-18", m["deadline"])
			}
			assert.Equal(t, append([]string{}, tt.want...), got)
		})
	}
}

func TestAIService_FindPeople(t *testing.T) {
	many := namedPeople("Sam 1", "Sam 2", "Sam 3", "Sam 4", "Sam 5", "Sam 6")
	tests := []struct {
		name    string
		role    string
		people  []models.Person
		query   string
		want    []string
		wantErr bool
	}{
		{name: "name match", role: constants.RolePlanner, people: namedPeople("Ada Lovelace", "Bob"), query: "LOVE",
			want: []string{"Ada Lovelace"}},
		{name: "no match gives an empty list", role: constants.RolePlanner, people: namedPeople("Bob"), query: "zzz"},
		{name: "at most 5 matches", role: constants.RolePlanner, people: many, query: "sam",
			want: []string{"Sam 1", "Sam 2", "Sam 3", "Sam 4", "Sam 5"}},
		{name: "manager allowed", role: constants.RoleManager, people: namedPeople("Ada"), query: "ada", want: []string{"Ada"}},
		{name: "specialist refused", role: constants.RoleSpecialist, people: namedPeople("Ada"), query: "ada", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &aiService{authz: ProvideAuthzService(leadPeople{}), people: listPeople{people: tt.people}}
			out, err := runFind(t, svc, asRole("kc-"+tt.role, tt.role), constants.AIToolFindPeople, tt.query)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			got := []string{}
			for _, it := range out["people"].([]any) {
				m := it.(map[string]any)
				got = append(got, m["name"].(string))
				assert.NotEmpty(t, m["id"])
				assert.Equal(t, "Engineer", m["primary_expertise"])
			}
			assert.Equal(t, append([]string{}, tt.want...), got)
		})
	}
}
