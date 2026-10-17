package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/integration/llm"
	"the-slot-finder/server/internal/models"
)

// listProjects answers ListByStatus like the real service: only roles that read programs get through.
type listProjects struct {
	ProjectService
	authz    AuthzService
	projects []models.Project
	got      *ListProjectsInput
}

func (l listProjects) ListByStatus(ctx context.Context, in ListProjectsInput) (*dtos.DataResponse[models.Project], error) {
	if _, err := l.authz.ProgramReader(ctx); err != nil {
		return nil, err
	}
	*l.got = in
	return &dtos.DataResponse[models.Project]{Data: l.projects}, nil
}

type listPeople struct {
	PeopleService
	people []models.Person
	got    *ListPeopleInput
}

func (l listPeople) List(_ context.Context, in ListPeopleInput) (*dtos.DataResponse[models.Person], error) {
	*l.got = in
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
	data, err := svc.dispatch(ctx, llm.FunctionCall{Name: tool, Args: map[string]any{"query": query}}, "", new(int))
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
			var in ListProjectsInput
			svc := &aiService{authz: authz, projects: listProjects{authz: authz, projects: tt.projects, got: &in}}
			out, err := runFind(t, svc, asRole("kc-"+tt.role, tt.role), constants.AIToolFindProjects, tt.query)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, constants.NoLimit, in.PageSize)
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
			var in ListPeopleInput
			svc := &aiService{authz: ProvideAuthzService(leadPeople{}), people: listPeople{people: tt.people, got: &in}}
			out, err := runFind(t, svc, asRole("kc-"+tt.role, tt.role), constants.AIToolFindPeople, tt.query)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, constants.NoLimit, in.PageSize)
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

// fakeScenarios answers Create like the real service for roles: planners only.
type fakeScenarios struct {
	ScenarioService
	authz   AuthzService
	view    *ScenarioView
	err     error
	created []CreateScenarioInput
}

func (f *fakeScenarios) Create(ctx context.Context, in CreateScenarioInput) (*ScenarioView, error) {
	if _, err := f.authz.Require(ctx, CanManagePlans); err != nil {
		return nil, err
	}
	if f.err != nil {
		return nil, f.err
	}
	f.created = append(f.created, in)
	return f.view, nil
}

func scenarioView(changes ...models.ScenarioChange) *ScenarioView {
	name := "x"
	return &ScenarioView{Plan: &models.Plan{BaseModel: models.BaseModel{ID: "s1"}, Name: &name,
		Summary: &models.ScenarioSummary{PeopleChanged: 2, HoursMoved: 30, UnmetHours: 4}}, Changes: changes}
}

func shiftArgs() map[string]any {
	return map[string]any{"assumptions": []any{map[string]any{"type": constants.AssumptionShiftProject, "project_id": "p1", "weeks": 2.0}}}
}

func weeksArgs(weeks float64) map[string]any {
	return map[string]any{"assumptions": []any{map[string]any{"type": constants.AssumptionShiftProject, "project_id": "p1", "weeks": weeks}}}
}

func TestAIService_CreateScenario(t *testing.T) {
	changes := make([]models.ScenarioChange, 0, 7)
	for i := 0; i < 7; i++ {
		changes = append(changes, models.ScenarioChange{PersonName: fmt.Sprintf("P%d", i), Before: 10, After: 10 + float64(i),
			Demand: &models.Demand{Expertise: models.Expertise{Name: "QA"}, Project: &models.Project{Name: "Web"}}})
	}
	long := strings.Repeat("é", 70)
	tests := []struct {
		name    string
		role    string
		call    llm.FunctionCall
		calls   int
		err     error
		wantErr string
		check   func(t *testing.T, out map[string]any, f *fakeScenarios)
	}{
		{name: "valid shift saves and returns result fields", role: constants.RolePlanner,
			call: llm.FunctionCall{Name: constants.AIToolCreateScenario, Args: shiftArgs()},
			check: func(t *testing.T, out map[string]any, f *fakeScenarios) {
				assert.Equal(t, "s1", out["scenario_id"])
				assert.EqualValues(t, 2, out["people_changed"])
				assert.EqualValues(t, 30, out["hours_moved"])
				assert.EqualValues(t, 4, out["unmet_hours"])
				require.Len(t, f.created, 1)
				assert.Equal(t, AssumptionInput{Type: constants.AssumptionShiftProject, ProjectID: "p1", Weeks: 2}, f.created[0].Assumptions[0])
			}},
		{name: "valid absence saves", role: constants.RolePlanner,
			call: llm.FunctionCall{Name: constants.AIToolCreateScenario, Args: map[string]any{"assumptions": []any{map[string]any{
				"type": constants.AssumptionAddAbsence, "person_id": "u1", "from_date": "2026-10-19", "to_date": "2026-10-23", "hours_per_day": 8.0}}}},
			check: func(t *testing.T, _ map[string]any, f *fakeScenarios) {
				require.Len(t, f.created, 1)
				a := f.created[0].Assumptions[0]
				assert.Equal(t, "u1", a.PersonID)
				assert.Equal(t, 8.0, a.HoursPerDay)
				assert.Equal(t, "2026-10-23", a.ToDate.Format(constants.DateLayout))
			}},
		{name: "bad project id gives error and saves nothing", role: constants.RolePlanner,
			call:    llm.FunctionCall{Name: constants.AIToolCreateScenario, Args: shiftArgs()},
			err:     errors.NewAppError(constants.ProjectNotFound, constants.MsgProjectNotFound, errors.ErrorTypeNotFound, 404),
			wantErr: constants.MsgProjectNotFound},
		{name: "fractional weeks refused", role: constants.RolePlanner, wantErr: constants.MsgAssumptionWeeksInvalid,
			call: llm.FunctionCall{Name: constants.AIToolCreateScenario, Args: weeksArgs(2.7)}},
		{name: "huge weeks refused", role: constants.RolePlanner, wantErr: constants.MsgAssumptionWeeksInvalid,
			call: llm.FunctionCall{Name: constants.AIToolCreateScenario, Args: weeksArgs(-1e20)}},
		{name: "infinite weeks refused", role: constants.RolePlanner, wantErr: constants.MsgAssumptionWeeksInvalid,
			call: llm.FunctionCall{Name: constants.AIToolCreateScenario, Args: weeksArgs(math.Inf(1))}},
		{name: "manager refused with error", role: constants.RoleManager,
			call: llm.FunctionCall{Name: constants.AIToolCreateScenario, Args: shiftArgs()}, wantErr: constants.MsgInsufficientPermissions},
		{name: "second call in the same Ask refused", role: constants.RolePlanner, calls: 1,
			call: llm.FunctionCall{Name: constants.AIToolCreateScenario, Args: shiftArgs()}, wantErr: constants.MsgOneScenarioPerQuestion},
		{name: "name is the prefix plus the first 60 characters of the question", role: constants.RolePlanner,
			call: llm.FunctionCall{Name: constants.AIToolCreateScenario, Args: shiftArgs()},
			check: func(t *testing.T, _ map[string]any, f *fakeScenarios) {
				assert.Equal(t, aiScenarioNamePrefix+strings.Repeat("é", 60), f.created[0].Name)
			}},
		{name: "top_changes capped at 5, largest first", role: constants.RolePlanner,
			call: llm.FunctionCall{Name: constants.AIToolCreateScenario, Args: shiftArgs()},
			check: func(t *testing.T, out map[string]any, _ *fakeScenarios) {
				top := out["top_changes"].([]any)
				require.Len(t, top, constants.AIScenarioMaxChanges)
				assert.Equal(t, "P6", top[0].(map[string]any)["person"])
				assert.Equal(t, "Web - QA", top[0].(map[string]any)["demand"])
				assert.Equal(t, "P2", top[4].(map[string]any)["person"])
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authz := ProvideAuthzService(leadPeople{})
			f := &fakeScenarios{authz: authz, view: scenarioView(changes...), err: tt.err}
			svc := &aiService{authz: authz, scenarios: f, cfg: &config.Config{LLMMaxToolResultBytes: 100000, ScenarioMaxShiftWeeks: 12}}
			name := aiScenarioNamePrefix + truncateRunes(long, constants.AIScenarioNameMaxRunes)
			saved := 0
			ctx := asRole("kc-"+tt.role, tt.role)
			for i := 0; i < tt.calls; i++ { // earlier calls in the same Ask
				svc.runTool(ctx, tt.call, name, &saved)
			}
			part := svc.runTool(ctx, tt.call, name, &saved)
			resp := part.FunctionResponse.Response
			if tt.wantErr != "" {
				assert.Equal(t, tt.wantErr, resp["error"])
				assert.Len(t, f.created, min(tt.calls, 1))
				return
			}
			require.Nil(t, resp["error"])
			var out map[string]any
			require.NoError(t, json.Unmarshal([]byte(resp["result"].(string)), &out))
			tt.check(t, out, f)
		})
	}
}

// scriptedLLM answers each Generate call with the next scripted content and records the requests.
type scriptedLLM struct {
	script   []llm.Content
	requests []llm.Request
}

func (f *scriptedLLM) Generate(_ context.Context, req llm.Request) (*llm.Response, error) {
	f.requests = append(f.requests, req)
	if len(f.requests) > len(f.script) {
		return nil, fmt.Errorf("scriptedLLM has no reply for call %d", len(f.requests))
	}
	return &llm.Response{Content: f.script[len(f.requests)-1]}, nil
}

func callPart(name string, args map[string]any) llm.Part {
	return llm.Part{FunctionCall: &llm.FunctionCall{Name: name, Args: args}}
}

func textContent(text string) llm.Content { return llm.Content{Parts: []llm.Part{{Text: text}}} }

func TestAIService_AskWhatIf(t *testing.T) {
	const answer = "Saved the scenario; it is on the Scenarios page."
	question := "What if Checkout slips 3 weeks? " + strings.Repeat("é", 70)
	find := callPart(constants.AIToolFindProjects, map[string]any{"query": "Checkout"})
	create := callPart(constants.AIToolCreateScenario, shiftArgs())
	tests := []struct {
		name         string
		role         string
		maxBytes     int
		script       []llm.Content
		wantCreated  int
		wantRequests int
		check        func(t *testing.T, f *scriptedLLM, s *fakeScenarios)
	}{
		{name: "find_projects, then create_scenario, then final text", role: constants.RolePlanner, maxBytes: 100000,
			script:      []llm.Content{{Parts: []llm.Part{find}}, {Parts: []llm.Part{create}}, textContent(answer)},
			wantCreated: 1, wantRequests: 3,
			check: func(t *testing.T, f *scriptedLLM, s *fakeScenarios) {
				// The name comes from the question only, not from the page context.
				assert.Equal(t, aiScenarioNamePrefix+truncateRunes(question, constants.AIScenarioNameMaxRunes), s.created[0].Name)
			}},
		{name: "tool result cut still applies", role: constants.RolePlanner, maxBytes: 20,
			script:      []llm.Content{{Parts: []llm.Part{create}}, textContent(answer)},
			wantCreated: 1, wantRequests: 2,
			check: func(t *testing.T, f *scriptedLLM, _ *fakeScenarios) {
				last := f.requests[1].Contents[len(f.requests[1].Contents)-1]
				sent := last.Parts[0].FunctionResponse.Response["result"].(string)
				assert.True(t, strings.HasSuffix(sent, constants.LLMToolCutMarker))
				assert.Len(t, sent, 20+len(constants.LLMToolCutMarker))
			}},
		{name: "two create_scenario calls in one Ask", role: constants.RolePlanner, maxBytes: 100000,
			script:      []llm.Content{{Parts: []llm.Part{create, create}}, textContent(answer)},
			wantCreated: 1, wantRequests: 2,
			check: func(t *testing.T, f *scriptedLLM, _ *fakeScenarios) {
				last := f.requests[1].Contents[len(f.requests[1].Contents)-1]
				require.Len(t, last.Parts, 2)
				assert.NotNil(t, last.Parts[0].FunctionResponse.Response["result"])
				assert.Equal(t, constants.MsgOneScenarioPerQuestion, last.Parts[1].FunctionResponse.Response["error"])
			}},
		{name: "manager Ask has no create_scenario in the tools", role: constants.RoleManager, maxBytes: 100000,
			script:       []llm.Content{textContent(answer)},
			wantRequests: 1,
			check: func(t *testing.T, f *scriptedLLM, _ *fakeScenarios) {
				for _, d := range f.requests[0].Tools[0].FunctionDeclarations {
					assert.NotEqual(t, constants.AIToolCreateScenario, d.Name)
				}
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authz := ProvideAuthzService(leadPeople{})
			scenarios := &fakeScenarios{authz: authz, view: scenarioView()}
			fake := &scriptedLLM{script: tt.script}
			svc := &aiService{llm: fake, authz: authz, scenarios: scenarios, projects: listProjects{authz: authz,
				projects: namedProjects("Checkout"), got: new(ListProjectsInput)},
				cfg: &config.Config{LLMAPIKey: "k", LLMMaxToolRounds: 5, LLMMaxToolResultBytes: tt.maxBytes, ScenarioMaxShiftWeeks: 12}}
			got, err := svc.Ask(asRole("kc-"+tt.role, tt.role), AskInput{Question: question, PageContext: "program p1"})
			require.NoError(t, err)
			assert.Equal(t, answer, got)
			assert.Len(t, scenarios.created, tt.wantCreated)
			assert.Len(t, fake.requests, tt.wantRequests)
			tt.check(t, fake, scenarios)
		})
	}
}

func TestAISystemPrompt_WhatIfRules(t *testing.T) {
	for _, want := range []string{constants.AIToolFindProjects, constants.AIToolFindPeople, constants.AIToolCreateScenario,
		"what-if", "Scenarios page", "Never invent ids", "2026-W45", "YYYY-MM-DD"} {
		assert.Contains(t, aiSystemPrompt, want)
	}
}
