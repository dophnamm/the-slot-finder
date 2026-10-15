package services

import (
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// inWeek48 is a Monday of 2026-W48: the weeks W46 and W47 of the QA project are history.
var inWeek48 = time.Date(2026, time.November, 23, 9, 0, 0, 0, time.UTC)

func (m staffingMocks) serviceAt(now time.Time) StaffingService {
	finder := leadPeople{}
	return ProvideStaffingService(m.plans, m.projects, m.people, m.absences, m.audit, m.events, finder, ProvideAuthzService(finder),
		passthroughTx{}, clock.NewFixed(now), testPlanningConfig(), nil)
}

func hoursInWeeks(a models.Assignment) map[string]float64 {
	out := map[string]float64{}
	for _, alloc := range a.Allocations {
		out[alloc.Week.String()] += alloc.Hours
	}
	return out
}

func TestStaffingService_Generate_PastWeeks(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	pastWeeks := []string{"2026-W46", "2026-W47"}

	tests := []struct {
		name     string
		baseline []models.Allocation
		// wantPast is the hours the draft must hold in the past weeks.
		wantPast map[string]float64
		wantSum  float64
	}{
		{name: "no history: nothing is planned in past weeks", wantPast: map[string]float64{}, wantSum: 60},
		{name: "past hours stay as they were and count towards the demand",
			baseline: []models.Allocation{{Week: mustWeek(t, "2026-W46"), Hours: 20}},
			wantPast: map[string]float64{"2026-W46": 20}, wantSum: 60},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newStaffingMocks()
			m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil)
			baseline := publishedBaseline(2)
			if tt.baseline != nil {
				baseline.Assignments = []models.Assignment{{BaseModel: models.BaseModel{ID: draftAssignmentID}, PersonID: qaPersonID,
					DemandID: demandID, Kind: constants.AssignmentKindPrimary, Allocations: tt.baseline}}
			}
			m.plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
			m.people.On("ListActive", mock.Anything).Return([]models.Person{qaPerson()}, nil)
			m.absences.On("ListOverlapping", mock.Anything, mock.Anything, mock.Anything).Return([]models.Absence{}, nil)
			var saved *models.Plan
			m.plans.On("CreateWithAssignments", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				saved = args.Get(1).(*models.Plan)
				m.plans.On("FindWithAssignments", mock.Anything, saved.ID).Return(saved, nil)
			}).Return(nil)

			_, err := m.serviceAt(inWeek48).Generate(planner, GenerateDraftInput{ProjectIDs: []string{projectID}})
			require.NoError(t, err)
			require.Len(t, saved.Assignments, 1)
			weeks := hoursInWeeks(saved.Assignments[0])
			for _, week := range pastWeeks {
				assert.InDelta(t, tt.wantPast[week], weeks[week], hoursEpsilon, week)
			}
			assert.InDelta(t, tt.wantSum, saved.Assignments[0].TotalHours(), hoursEpsilon)
			assert.Empty(t, saved.Unmet)
		})
	}
}
