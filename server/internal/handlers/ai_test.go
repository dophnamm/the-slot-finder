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
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAI answers SummarizeDraft with a fixed summary or error and records the draft id.
type fakeAI struct {
	services.AIService
	summary   *models.DraftSummary
	err       error
	gotID     string
	intake    *services.IntakeResult
	gotIntake services.IntakeInput
}

func (f *fakeAI) SummarizeDraft(_ context.Context, draftID string) (*models.DraftSummary, error) {
	f.gotID = draftID
	return f.summary, f.err
}

func TestAIHandler_SummarizeDraft(t *testing.T) {
	const draftID = "0192f3a4-0000-7000-8000-0000000000aa"
	tests := []struct {
		name       string
		id         string
		ai         *fakeAI
		wantStatus int
		wantCode   string
	}{
		{name: "summarizes", id: draftID, ai: &fakeAI{summary: &models.DraftSummary{Summary: "Done.", Source: constants.SummarySourceAI}},
			wantStatus: http.StatusOK, wantCode: constants.Success},
		{name: "bad id", id: "nope", ai: &fakeAI{}, wantStatus: http.StatusBadRequest, wantCode: constants.ValidationError},
		{name: "missing draft", id: draftID, ai: &fakeAI{err: errors.NotFoundError("draft", nil)},
			wantStatus: http.StatusNotFound, wantCode: constants.UserNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ProvideAIHandler(tt.ai, nil)
			e := echo.New()
			e.POST(constants.PathStaffingDraftSummary, h.SummarizeDraft)
			req := httptest.NewRequest(http.MethodPost, "/staffing-drafts/"+tt.id+"/summary", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())

			var body dtos.BaseResponse[dtos.DraftSummaryResponse]
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, tt.wantCode, body.Meta.ErrorCode)
			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, dtos.DraftSummaryResponse{Summary: "Done.", Source: constants.SummarySourceAI}, body.Data)
				assert.Equal(t, draftID, tt.ai.gotID)
			}
		})
	}
}

func (f *fakeAI) ExtractIntake(_ context.Context, in services.IntakeInput) (*services.IntakeResult, error) {
	f.gotIntake = in
	return f.intake, f.err
}

func TestNewIntakeResponse(t *testing.T) {
	level := 2
	points := 12.0
	start := "2026-11-02"
	full := services.IntakeResult{
		Suggestion: services.Suggestion{
			Name: "Checkout Revamp", StartEarliest: &start, ScopePoints: &points,
			Demands: []services.SuggestedDemand{{ExpertiseID: "e1", ExpertiseName: "QA", TotalHours: 120, MinLevel: &level, Quote: "120 hours of QA"}},
		},
		Warnings: []string{"w"},
	}
	full.Masked.Emails, full.Masked.Customers, full.Masked.Phones = 1, 2, 3
	raw, err := json.Marshal(newIntakeResponse(&full))
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"Checkout Revamp","start_earliest":"2026-11-02","deadline":null,"scope_points":12,
		"demands":[{"expertise_id":"e1","expertise_name":"QA","total_hours":120,"min_level":2,"quote":"120 hours of QA"}],
		"warnings":["w"],"masked":{"emails":1,"customers":2,"phones":3}}`, string(raw))

	raw, err = json.Marshal(newIntakeResponse(&services.IntakeResult{}))
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"","start_earliest":null,"deadline":null,"scope_points":null,"demands":[],"warnings":[],
		"masked":{"emails":0,"customers":0,"phones":0}}`, string(raw))
}

func TestAIHandler_ExtractIntake(t *testing.T) {
	result := &services.IntakeResult{Suggestion: services.Suggestion{Name: "Checkout"}}
	result.Masked.Emails = 2
	tests := []struct {
		name       string
		body       string
		ai         *fakeAI
		wantStatus int
		wantCode   string
	}{
		{name: "suggests", body: `{"text":" a brief ","mask_terms":["Acme"]}`, ai: &fakeAI{intake: result},
			wantStatus: http.StatusOK, wantCode: constants.Success},
		{name: "empty text", body: `{"text":" "}`, ai: &fakeAI{}, wantStatus: http.StatusBadRequest, wantCode: constants.ValidationError},
		{name: "service error", body: `{"text":"x"}`, ai: &fakeAI{err: errors.ForbiddenError("no", nil)},
			wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ProvideAIHandler(tt.ai, validator.New())
			e := echo.New()
			e.POST(constants.PathAIIntakeExtractions, h.ExtractIntake)
			req := httptest.NewRequest(http.MethodPost, constants.PathAIIntakeExtractions, strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantStatus != http.StatusOK {
				return
			}
			var body dtos.BaseResponse[dtos.IntakeResponse]
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, 2, body.Data.Masked.Emails)
			assert.Equal(t, services.IntakeInput{Text: "a brief", MaskTerms: []string{"Acme"}}, tt.ai.gotIntake)
		})
	}
}
