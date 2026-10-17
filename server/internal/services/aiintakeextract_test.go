package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/integration/llm"
	"the-slot-finder/server/internal/models"
)

type intakePrograms struct {
	ProgramService
	programs []models.Program
	got      *ListProgramsInput
}

func (p intakePrograms) List(_ context.Context, in ListProgramsInput) (*dtos.DataResponse[models.Program], error) {
	*p.got = in
	return &dtos.DataResponse[models.Program]{Data: p.programs}, nil
}

type intakeExpertises struct {
	PeopleService
}

func (intakeExpertises) ListExpertises(context.Context) ([]models.Expertise, error) {
	return []models.Expertise{{BaseModel: models.BaseModel{ID: "e-qa"}, Name: "QA"}, {BaseModel: models.BaseModel{ID: "e-eng"}, Name: "Engineer"}}, nil
}

// failingLLM answers every call with err.
type failingLLM struct {
	err   error
	calls int
}

func (f *failingLLM) Generate(context.Context, llm.Request) (*llm.Response, error) {
	f.calls++
	return nil, f.err
}

const (
	intakeBrief = "Acme Corp wants a checkout revamp. Contact ana.lee@acme.com or +84 912 345 678, Bao Tran leads it. " +
		"We expect about 120 hours of QA."
)

func validCall() llm.Content {
	return llm.Content{Parts: []llm.Part{callPart(constants.AIToolSuggestProject, map[string]any{
		"name": "Checkout Revamp", "start_earliest": nil, "deadline": "2026-12-18", "scope_points": 12,
		"demands": []any{map[string]any{"expertise_name": "QA", "total_hours": 120, "min_level": 2, "quote": "about 120 hours of QA"}},
	})}}
}

func TestAIService_ExtractIntake(t *testing.T) {
	programs := []models.Program{{Client: "Acme Corp", LeadID: intakeLead(leadPersonID)}, {Client: "Globex", LeadID: intakeLead(leadPersonID)}}
	tests := []struct {
		name      string
		role      string
		subject   string
		key       string
		programs  []models.Program
		script    []llm.Content
		wantCode  string
		wantCalls int
		check     func(t *testing.T, res *IntakeResult)
	}{
		{name: "valid suggestion", role: constants.RolePlanner, key: "k", programs: programs, script: []llm.Content{validCall()}, wantCalls: 1,
			check: func(t *testing.T, res *IntakeResult) {
				assert.Equal(t, "Checkout Revamp", res.Suggestion.Name)
				require.Len(t, res.Suggestion.Demands, 1)
				assert.Equal(t, "e-qa", res.Suggestion.Demands[0].ExpertiseID)
				assert.Equal(t, maskCounts{Emails: 1, Customers: 2, Phones: 1}, res.Masked)
				assert.Empty(t, res.Warnings)
			}},
		{name: "manager who leads a program", role: constants.RoleManager, subject: "kc-manager", key: "k", programs: programs,
			script: []llm.Content{validCall()}, wantCalls: 1},
		{name: "no function call, then success", role: constants.RolePlanner, key: "k", programs: programs,
			script: []llm.Content{textContent("sure"), validCall()}, wantCalls: 2},
		{name: "no function call twice", role: constants.RolePlanner, key: "k", programs: programs,
			script: []llm.Content{textContent("a"), textContent("b")}, wantCalls: 2, wantCode: constants.AIUpstreamFailed},
		{name: "wrong function name twice", role: constants.RolePlanner, key: "k", programs: programs,
			script: []llm.Content{{Parts: []llm.Part{callPart("other", nil)}}, {Parts: []llm.Part{callPart("other", nil)}}}, wantCalls: 2,
			wantCode: constants.AIUpstreamFailed},
		{name: "invalid arguments twice", role: constants.RolePlanner, key: "k", programs: programs,
			script: []llm.Content{{Parts: []llm.Part{callPart(constants.AIToolSuggestProject, map[string]any{"name": 5})}},
				{Parts: []llm.Part{callPart(constants.AIToolSuggestProject, map[string]any{"name": 5})}}},
			wantCalls: 2, wantCode: constants.AIUpstreamFailed},
		{name: "not configured", role: constants.RolePlanner, programs: programs, wantCalls: 0, wantCode: constants.AINotConfigured},
		{name: "manager without a program of their own refused", role: constants.RoleManager, subject: "kc-manager", key: "k",
			programs: []models.Program{{Client: "Acme Corp", LeadID: intakeLead(otherPersonID2)}}, wantCode: constants.Forbidden},
		{name: "specialist refused", role: constants.RoleSpecialist, key: "k", programs: programs, wantCode: constants.Forbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authz := ProvideAuthzService(leadPeople{})
			fake := &scriptedLLM{script: tt.script}
			var in ListProgramsInput
			svc := &aiService{llm: fake, authz: authz, people: intakeExpertises{}, programs: intakePrograms{programs: tt.programs, got: &in},
				cfg: &config.Config{LLMAPIKey: tt.key}}
			subject := tt.subject
			if subject == "" {
				subject = "kc-" + tt.role
			}
			res, err := svc.ExtractIntake(asRole(subject, tt.role),
				IntakeInput{Text: intakeBrief, MaskTerms: []string{"Bao Tran"}})
			assert.Len(t, fake.requests, tt.wantCalls)
			if tt.wantCode != "" {
				var appErr *errors.AppError
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.wantCode, appErr.Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, constants.NoLimit, in.PageSize)
			if tt.check != nil {
				tt.check(t, res)
			}
		})
	}
}

func TestAIService_ExtractIntake_requestHidesMaskedText(t *testing.T) {
	fake := &scriptedLLM{script: []llm.Content{validCall()}}
	svc := &aiService{llm: fake, authz: ProvideAuthzService(leadPeople{}), people: intakeExpertises{},
		programs: intakePrograms{programs: []models.Program{{Client: "Acme Corp"}}, got: new(ListProgramsInput)},
		cfg:      &config.Config{LLMAPIKey: "k"}}
	_, err := svc.ExtractIntake(asRole("kc-planner", constants.RolePlanner), IntakeInput{Text: intakeBrief, MaskTerms: []string{"Bao Tran"}})
	require.NoError(t, err)
	require.Len(t, fake.requests, 1)
	body, err := json.Marshal(fake.requests[0])
	require.NoError(t, err)
	for _, secret := range []string{"ana.lee@acme.com", "acme.com", "Acme", "Bao Tran", "912 345 678"} {
		assert.NotContains(t, string(body), secret)
	}
	assert.Contains(t, string(body), "about 120 hours of QA")
	assert.Contains(t, string(body), "e-qa")
	require.Len(t, fake.requests[0].Tools, 1)
	require.Len(t, fake.requests[0].Tools[0].FunctionDeclarations, 1)
	assert.Equal(t, constants.AIToolSuggestProject, fake.requests[0].Tools[0].FunctionDeclarations[0].Name)
}

func TestAIService_ExtractIntake_providerErrors(t *testing.T) {
	tests := []struct {
		err      error
		wantCode string
		status   int
	}{
		{err: &llm.RateLimitedError{}, wantCode: constants.AIRateLimited, status: 429},
		{err: fmt.Errorf("boom"), wantCode: constants.AIUpstreamFailed, status: 502},
	}
	for _, tt := range tests {
		t.Run(tt.wantCode, func(t *testing.T) {
			fake := &failingLLM{err: tt.err}
			svc := &aiService{llm: fake, authz: ProvideAuthzService(leadPeople{}), people: intakeExpertises{},
				programs: intakePrograms{got: new(ListProgramsInput)}, cfg: &config.Config{LLMAPIKey: "k"}}
			_, err := svc.ExtractIntake(asRole("kc-planner", constants.RolePlanner), IntakeInput{Text: intakeBrief})
			var appErr *errors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, tt.wantCode, appErr.Code)
			assert.Equal(t, tt.status, appErr.HTTPStatus)
			assert.Equal(t, 1, fake.calls)
		})
	}
}

func intakeLead(id string) *string { return &id }
