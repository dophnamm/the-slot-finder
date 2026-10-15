package services

import (
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// The test clock is in 2026-W42: W41 is history, W42 and W43 are current and future.
func TestInsightsService_Heatmap_InactiveProjects(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	from, to := mustWeek(t, "2026-W41"), mustWeek(t, "2026-W43")
	tests := []struct {
		name   string
		status string
		want   [3]float64 // Erin's load in W41, W42, W43
	}{
		{name: "active project counts every week", status: constants.ProjectStatusActive, want: [3]float64{24, 24, 24}},
		{name: "planned project counts every week", status: constants.ProjectStatusPlanned, want: [3]float64{24, 24, 24}},
		{name: "cancelled project keeps history only", status: constants.ProjectStatusCancelled, want: [3]float64{24, 0, 0}},
		{name: "done project keeps history only", status: constants.ProjectStatusDone, want: [3]float64{24, 0, 0}},
		{name: "on hold project keeps history only", status: constants.ProjectStatusOnHold, want: [3]float64{24, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseline := publishedBaseline(1)
			baseline.Assignments = []models.Assignment{{PersonID: engPersonID,
				Demand: &models.Demand{ExpertiseID: expEngineerID, Expertise: engineerExp, Project: &models.Project{Status: tt.status}},
				Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W41"), Hours: 24}, {Week: mustWeek(t, "2026-W42"), Hours: 24},
					{Week: mustWeek(t, "2026-W43"), Hours: 24}}}}
			plans, people, absences := new(MockPlanRepository), new(MockPersonRepository), new(MockAbsenceRepository)
			plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
			people.On("ListActive", mock.Anything).Return([]models.Person{
				heatmapPerson(engPersonID, "Erin", skill(engineerExp, 4, true))}, nil)
			absences.On("ListOverlapping", mock.Anything, mock.Anything, mock.Anything).Return([]models.Absence{}, nil)

			heatmap, err := newInsightsService(plans, people, absences).Heatmap(planner, HeatmapInput{From: &from, To: &to})
			require.NoError(t, err)
			require.Len(t, heatmap.Rows, 1)
			for i, cell := range heatmap.Rows[0].Cells {
				assert.InDelta(t, tt.want[i], cell.Load, hoursEpsilon, cell.Week.String())
			}
		})
	}
}
