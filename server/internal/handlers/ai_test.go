package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// fakeAI answers SummarizeDraft with a fixed summary or error and records the draft id.
type fakeAI struct {
	services.AIService
	summary *models.DraftSummary
	err     error
	gotID   string
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
