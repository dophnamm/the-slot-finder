//go:build integration

package services

import (
	"context"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Two approvals of the same planned absence that both read "planned" must not both win:
// the write itself only turns a still-planned absence into an approved one.
func TestAbsenceRepository_ApprovePlanned_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	ctx := context.Background()
	ada := models.Person{BaseModel: models.NewBaseModel(), Name: "Ada", Email: "ada@example.test", Role: constants.RoleSpecialist,
		WeeklyHours: 40, NonProjectHours: 4, Active: true}
	require.NoError(t, pdb.Create(&ada).Error)
	leave := models.Absence{BaseModel: models.NewBaseModel(), PersonID: &ada.ID, FromDate: day(time.October, 22),
		ToDate: day(time.October, 22), HoursPerDay: 8, Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusPlanned}
	holiday := models.Absence{BaseModel: models.NewBaseModel(), FromDate: day(time.October, 23), ToDate: day(time.October, 23),
		HoursPerDay: 8, Type: constants.AbsenceTypeHoliday, Status: constants.AbsenceStatusPlanned}
	require.NoError(t, pdb.Create(&leave).Error)
	require.NoError(t, pdb.Create(&holiday).Error)
	repo := repositories.ProvideAbsenceRepository(pdb)

	first, err := repo.ApprovePlanned(ctx, leave.ID)
	require.NoError(t, err)
	assert.True(t, first, "the first approval wins")
	second, err := repo.ApprovePlanned(ctx, leave.ID)
	require.NoError(t, err)
	assert.False(t, second, "the second finds it approved already")

	stored, err := repo.FindOneByID(ctx, leave.ID)
	require.NoError(t, err)
	assert.Equal(t, constants.AbsenceStatusApproved, stored.Status)

	changed, err := repo.ApprovePlanned(ctx, holiday.ID)
	require.NoError(t, err)
	assert.False(t, changed, "a holiday is never approved this way")
}

// Calendar leave deleted by hand is remembered by its event UID; leave the sync removed is not.
func TestAbsenceRepository_DismissCalendarLeave_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	ctx := context.Background()
	ada := models.Person{BaseModel: models.NewBaseModel(), Name: "Ada", Email: "ada@example.test", Role: constants.RoleSpecialist,
		WeeklyHours: 40, Active: true}
	require.NoError(t, pdb.Create(&ada).Error)
	repo := repositories.ProvideAbsenceRepository(pdb)
	leave := func(uid string) models.Absence {
		return models.Absence{BaseModel: models.NewBaseModel(), PersonID: &ada.ID, FromDate: day(time.October, 22), ToDate: day(time.October, 22),
			HoursPerDay: 8, Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusPlanned, Source: constants.AbsenceSourceCalendar, ExternalUID: &uid}
	}
	byHand, bySync := leave("by-hand"), leave("by-sync")
	require.NoError(t, pdb.Create(&byHand).Error)
	require.NoError(t, pdb.Create(&bySync).Error)

	require.NoError(t, repo.DismissCalendarLeave(ctx, &byHand))
	require.NoError(t, repo.Delete(ctx, &bySync))

	uids, err := repo.ListDismissedCalendarUIDs(ctx, ada.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"by-hand"}, uids)
	live, err := repo.ListCalendarLeave(ctx, ada.ID)
	require.NoError(t, err)
	assert.Empty(t, live)
}
