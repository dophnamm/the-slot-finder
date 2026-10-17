package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/integration/llm"
)

func declaredTools(role string) []string {
	var names []string
	for _, d := range aiToolDeclarations(role) {
		names = append(names, d.Name)
	}
	return names
}

func TestAIToolDeclarations(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		wantCreate bool
	}{
		{name: "planner has all three tools kinds", role: constants.RolePlanner, wantCreate: true},
		{name: "manager has no create_scenario", role: constants.RoleManager},
		{name: "unknown role has no create_scenario", role: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names := declaredTools(tt.role)
			assert.Contains(t, names, constants.AIToolFindProjects)
			assert.Contains(t, names, constants.AIToolFindPeople)
			assert.Equal(t, tt.wantCreate, assert.ObjectsAreEqual(true, containsName(names, constants.AIToolCreateScenario)))
		})
	}
}

func containsName(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}

var _ = llm.FunctionDeclaration{}
