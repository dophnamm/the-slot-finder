package dtos

import (
	"testing"
	"time"

	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProjectResponse_ProgramLeadID(t *testing.T) {
	lead := "lee"
	now := isoweek.Of(time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC))

	got := NewProjectResponse(models.Project{Program: &models.Program{Name: "Acme", LeadID: &lead}}, now)
	require.NotNil(t, got.ProgramLeadID)
	assert.Equal(t, lead, *got.ProgramLeadID)

	assert.Nil(t, NewProjectResponse(models.Project{}, now).ProgramLeadID, "unknown without the program loaded")
}
