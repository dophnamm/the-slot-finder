//go:build integration

package services

import (
	"net/http"
	"testing"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeopleImport_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	qa := models.Expertise{BaseModel: models.NewBaseModel(), Name: "QA"}
	require.NoError(t, pdb.Create(&qa).Error)
	planner := asRole("kc-planner", constants.RolePlanner)
	cfg := peopleImportConfig()
	cfg.PeopleImportMaxRows = 10
	repo := repositories.ProvidePersonRepository(pdb)
	tx := db.ProvideTxManager(pdb)
	events := ProvideEventRecorder(repositories.ProvideChangeEventRepository(pdb), clock.NewFixed(tsNow))
	svc := ProvidePeopleImportService(ProvidePeopleService(repo, tx, cfg, events), repo, ProvideAuthzService(leadPeople{}), tx, cfg)
	row := func(name, email, level string) string {
		return name + "," + email + ",specialist,40,4,,,QA," + level + ",\n"
	}
	good := piHeader + row("Quinn", "quinn@example.test", "3")
	people := func() int64 {
		var n int64
		require.NoError(t, pdb.Model(&models.Person{}).Count(&n).Error)
		return n
	}
	eventCount := func(eventType string) int64 {
		var n int64
		q := pdb.Model(&models.ChangeEvent{})
		if eventType != "" {
			q = q.Where("type = ?", eventType)
		}
		require.NoError(t, q.Count(&n).Error)
		return n
	}

	_, err := svc.Import(planner, importFile(good+row("Bad", "not-an-email", "3")))
	requireAppError(t, err, constants.PeopleImportInvalid, http.StatusBadRequest)
	assert.Zero(t, people(), "one bad row saves nothing")

	result, err := svc.Import(planner, importFile(good))
	require.NoError(t, err)
	assert.Equal(t, models.PeopleImportResult{Created: 1}, *result)

	before := eventCount("")
	result, err = svc.Import(planner, importFile(good))
	require.NoError(t, err)
	assert.Equal(t, models.PeopleImportResult{Unchanged: 1}, *result)
	assert.Equal(t, before, eventCount(""), "an unchanged file records no event")

	skillEvents := eventCount(constants.EventSkillChanged)
	result, err = svc.Import(planner, importFile(piHeader+row("Quinn", "quinn@example.test", "4")))
	require.NoError(t, err)
	assert.Equal(t, models.PeopleImportResult{Updated: 1}, *result)
	assert.Equal(t, skillEvents+1, eventCount(constants.EventSkillChanged), "a changed level records one SkillChanged")
}
