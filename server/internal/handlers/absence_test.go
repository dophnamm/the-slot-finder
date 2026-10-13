package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAbsenceID = "0192f3a4-0000-7000-8000-0000000000c1"

// fakeAvailability implements only Approve; other methods panic if called.
type fakeAvailability struct {
	services.AvailabilityService
	approved string
	err      error
}

func (f *fakeAvailability) Approve(_ context.Context, id string) (*models.Absence, error) {
	f.approved = id
	if f.err != nil {
		return nil, f.err
	}
	person := testPersonID
	day := time.Date(2026, time.October, 22, 0, 0, 0, 0, time.UTC)
	return &models.Absence{BaseModel: models.BaseModel{ID: id}, PersonID: &person, FromDate: day, ToDate: day, HoursPerDay: 8,
		Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusApproved}, nil
}

func TestAbsenceHandler_Approve(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serviceErr error
		wantStatus int
	}{
		{name: "approved", target: "/absences/" + testAbsenceID + "/approve", wantStatus: http.StatusOK},
		{name: "not a uuid", target: "/absences/abc/approve", wantStatus: http.StatusBadRequest},
		{name: "already approved", target: "/absences/" + testAbsenceID + "/approve", wantStatus: http.StatusConflict,
			serviceErr: errors.NewAppError(constants.AbsenceNotPlanned, constants.MsgAbsenceNotPlanned, errors.ErrorTypeConflict, http.StatusConflict)},
		{name: "unknown absence", target: "/absences/" + testAbsenceID + "/approve", wantStatus: http.StatusNotFound,
			serviceErr: errors.NewAppError(constants.AbsenceNotFound, constants.MsgAbsenceNotFound, errors.ErrorTypeNotFound, http.StatusNotFound)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			availability := &fakeAvailability{err: tt.serviceErr}
			h := ProvideAbsenceHandler(availability, newTestValidator(t))
			rec, envelope := serve(t, constants.PathAbsenceApprove, http.MethodPost, tt.target, "", h.Approve)
			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, testAbsenceID, availability.approved)
				var absence dtos.AbsenceResponse
				require.NoError(t, json.Unmarshal(envelope.Data, &absence))
				assert.Equal(t, constants.AbsenceStatusApproved, absence.Status)
			}
			if tt.wantStatus == http.StatusConflict {
				assert.Equal(t, constants.AbsenceNotPlanned, envelope.Meta.ErrorCode)
			}
		})
	}
}
