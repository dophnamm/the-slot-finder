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
	"the-slot-finder/server/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSchedule records the input and returns one week without projects and one holiday.
type fakeSchedule struct {
	got *services.ScheduleInput
	err error
}

func (f *fakeSchedule) Get(_ context.Context, in services.ScheduleInput) (*services.PersonSchedule, error) {
	f.got = &in
	if f.err != nil {
		return nil, f.err
	}
	return &services.PersonSchedule{
		WorkingWeekdays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Sunday},
		Weeks:           []services.ScheduleWeek{{Week: in.From, Capacity: 36, Band: constants.BandUnderused}},
		Days: []services.ScheduleDay{{Date: time.Date(2026, time.October, 14, 0, 0, 0, 0, time.UTC),
			Holidays: []string{"0192f3a4-0000-7000-8000-0000000000f1"}}},
	}, nil
}

func TestScheduleHandler_Get(t *testing.T) {
	path := "/people/" + testPersonID + "/schedule"

	t.Run("maps the schedule with ISO weekdays and empty lists", func(t *testing.T) {
		schedule := &fakeSchedule{}
		h := ProvideScheduleHandler(schedule, newTestValidator(t))
		rec, envelope := serve(t, constants.PathPersonSchedule, http.MethodGet, path+"?from=2026-W42&to=2026-W43", "", h.Get)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		require.NotNil(t, schedule.got)
		assert.Equal(t, testPersonID, schedule.got.PersonID)
		assert.Equal(t, "2026-W42", schedule.got.From.String())
		assert.Equal(t, "2026-W43", schedule.got.To.String())

		var body dtos.ScheduleResponse
		require.NoError(t, json.Unmarshal(envelope.Data, &body))
		assert.Equal(t, []int{1, 2, 3, 4, 5, 7}, body.WorkingWeekdays, "Sunday is 7 in ISO 8601")
		require.Len(t, body.Weeks, 1)
		assert.Equal(t, "2026-W42", body.Weeks[0].Week)
		require.Len(t, body.Days, 1)
		assert.Equal(t, "2026-10-14", body.Days[0].Date)
		assert.Equal(t, "0192f3a4-0000-7000-8000-0000000000f1", body.Days[0].Holidays[0].ID)
		raw := string(envelope.Data)
		assert.Contains(t, raw, `"projects":[]`, "an empty list is [], never null")
		assert.Contains(t, raw, `"absences":[]`)
	})

	t.Run("refusals", func(t *testing.T) {
		tests := []struct {
			name       string
			target     string
			serviceErr error
			wantStatus int
			wantField  string
		}{
			{name: "missing from", target: path + "?to=2026-W43", wantStatus: http.StatusBadRequest, wantField: "from"},
			{name: "not an ISO week", target: path + "?from=2026-W42&to=2026-43", wantStatus: http.StatusBadRequest, wantField: "to"},
			{name: "not a uuid", target: "/people/abc/schedule?from=2026-W42&to=2026-W43", wantStatus: http.StatusBadRequest},
			{name: "someone else's schedule", target: path + "?from=2026-W42&to=2026-W43", wantStatus: http.StatusForbidden,
				serviceErr: errors.ForbiddenError(constants.MsgInsufficientPermissions, nil)},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h := ProvideScheduleHandler(&fakeSchedule{err: tt.serviceErr}, newTestValidator(t))
				rec, envelope := serve(t, constants.PathPersonSchedule, http.MethodGet, tt.target, "", h.Get)
				require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
				if tt.wantField != "" {
					var fields map[string]string
					require.NoError(t, json.Unmarshal(envelope.Data, &fields))
					assert.Contains(t, fields, validationKeyPrefix+tt.wantField)
				}
			})
		}
	})
}
