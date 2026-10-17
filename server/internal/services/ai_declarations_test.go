package services

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"the-slot-finder/server/internal/constants"
)

func TestAIToolDeclarations(t *testing.T) {
	read := []string{constants.AIToolHeatmap, constants.AIToolCapacityForecast, constants.AIToolRisks, constants.AIToolOpenDrafts,
		constants.AIToolProject, constants.AIToolFindProjects, constants.AIToolFindPeople, constants.AIToolProgramOverview}
	tests := []struct {
		name string
		role string
		want []string
	}{
		{name: "planner has every tool, including create_scenario", role: constants.RolePlanner, want: append(slices.Clone(read), constants.AIToolCreateScenario)},
		{name: "manager has no create_scenario", role: constants.RoleManager, want: read},
		{name: "unknown role has no create_scenario", role: "", want: read},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var names []string
			for _, d := range aiToolDeclarations(tt.role) {
				names = append(names, d.Name)
			}
			assert.ElementsMatch(t, tt.want, names)
		})
	}
}
