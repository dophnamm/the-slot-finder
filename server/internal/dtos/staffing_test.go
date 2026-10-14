package dtos

import (
	"encoding/json"
	"testing"
	"time"

	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDraftResponse_TeamIDs(t *testing.T) {
	start := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	project := func(id, programID, demandID string) models.Project {
		return models.Project{BaseModel: models.BaseModel{ID: id}, ProgramID: programID, StartEarliest: start,
			Deadline: start.AddDate(0, 0, 14), Demands: []models.Demand{{BaseModel: models.BaseModel{ID: demandID}}}}
	}
	projects := []models.Project{
		project("p-acme", "acme", "d-acme"),
		project("p-beta", "beta", "d-beta"),
		project("p-gamma", "gamma", "d-gamma"),
	}
	teams := map[string]map[string]bool{"acme": {"bao": true, "alex": true}, "beta": {}}

	got := NewDraftResponse(models.Plan{}, projects, nil, 1, nil, nil, teams)

	require.Len(t, got.Demands, 3)
	assert.Equal(t, []string{"alex", "bao"}, got.Demands[0].TeamIDs, "sorted team of the demand's program")
	assert.Equal(t, []string{}, got.Demands[1].TeamIDs, "an empty team lets nobody staff the demand")
	assert.Nil(t, got.Demands[2].TeamIDs, "a program without a team puts no limit")

	body, err := json.Marshal(got.Demands[2])
	require.NoError(t, err)
	assert.Contains(t, string(body), `"team_ids":null`)
}

func TestNewDraftResponse_ProgramLeadID(t *testing.T) {
	lead := "lee"
	projects := []models.Project{
		{BaseModel: models.BaseModel{ID: "p-acme"}, Program: &models.Program{Name: "Acme", LeadID: &lead}},
		{BaseModel: models.BaseModel{ID: "p-beta"}, Program: &models.Program{Name: "Beta"}},
	}

	got := NewDraftResponse(models.Plan{}, projects, nil, 1, nil, nil, nil)

	require.Len(t, got.Projects, 2)
	require.NotNil(t, got.Projects[0].ProgramLeadID)
	assert.Equal(t, lead, *got.Projects[0].ProgramLeadID, "the client shows draft actions to the program's manager")
	assert.Nil(t, got.Projects[1].ProgramLeadID)
}
