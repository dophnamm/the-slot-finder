package services

import (
	"context"
	"net/http"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/request"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	managerPersonID  = "0192f3a4-0000-7000-8000-0000000000c1"
	teamMemberID     = "0192f3a4-0000-7000-8000-0000000000c2"
	outsiderPersonID = "0192f3a4-0000-7000-8000-0000000000c3"
)

// selfFinder maps the test subjects to their person records.
type selfFinder map[string]string

func (f selfFinder) FindForPrincipal(_ context.Context, principal request.Principal) (*models.Person, error) {
	if id, ok := f[principal.Subject]; ok {
		return &models.Person{BaseModel: models.BaseModel{ID: id}}, nil
	}
	return nil, nil
}

// ledTeams returns the team of the programs each lead leads.
type ledTeams map[string][]string

func (t ledTeams) PersonIDsLedBy(_ context.Context, leadID string) ([]string, error) {
	return t[leadID], nil
}

func newVisibilityAuthz() AuthzService {
	return ProvideTeamAuthzService(
		selfFinder{"kc-manager": managerPersonID, "kc-specialist": selfPersonID},
		ledTeams{managerPersonID: {teamMemberID}},
	)
}

func TestAuthzService_VisiblePeople(t *testing.T) {
	authz := newVisibilityAuthz()

	planner, err := authz.VisiblePeople(asRole("kc-planner", constants.RolePlanner))
	require.NoError(t, err)
	assert.True(t, planner.All)
	assert.True(t, planner.Allows(outsiderPersonID))

	lead, err := authz.VisiblePeople(asRole("kc-manager", constants.RoleManager))
	require.NoError(t, err)
	assert.False(t, lead.All)
	assert.True(t, lead.Allows(teamMemberID), "someone in a team of a program they lead")
	assert.True(t, lead.Allows(managerPersonID), "themselves")
	assert.False(t, lead.Allows(outsiderPersonID))
	assert.Equal(t, []string{managerPersonID, teamMemberID}, lead.IDs())

	specialist, err := authz.VisiblePeople(asRole("kc-specialist", constants.RoleSpecialist))
	require.NoError(t, err)
	assert.Equal(t, []string{selfPersonID}, specialist.IDs())
}

func TestAuthzService_RequireVisiblePerson(t *testing.T) {
	authz := newVisibilityAuthz()
	lead := asRole("kc-manager", constants.RoleManager)

	_, err := authz.RequireVisiblePerson(lead, teamMemberID)
	require.NoError(t, err)
	_, err = authz.RequireVisiblePerson(lead, outsiderPersonID)
	requireStatus(t, err, http.StatusForbidden)
	_, err = authz.RequireVisiblePerson(asRole("kc-planner", constants.RolePlanner), outsiderPersonID)
	require.NoError(t, err)
}
