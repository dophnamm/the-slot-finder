package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"strconv"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/repositories"
	"the-slot-finder/server/internal/request"
	"the-slot-finder/server/internal/services"
)

// actualFactors are the demo's actual hours as a share of planned hours, per project. Data
// Migration runs over plan and Checkout Revamp under it; Patient App is left for a live import.
var actualFactors = map[string]float64{"Data Migration": 1.15, "Checkout Revamp": 0.9}

// actualHourStep rounds seeded actual hours to half hours, as people write them.
const actualHourStep = 0.5

// seedActuals fills the timesheet template for the demo projects and imports it through the
// timesheet service, so the seed passes the same checks as the app. Rows that already hold
// actual hours are left alone, so running the seed again adds nothing.
func seedActuals(database *db.PostgresDB, cfg *config.Config) (int, error) {
	clk, err := clock.ProvideClock(cfg)
	if err != nil {
		return 0, err
	}
	people := repositories.ProvidePersonRepository(database)
	tx := db.ProvideTxManager(database)
	finder := services.ProvidePersonFinder(services.ProvidePeopleService(people, tx, cfg, services.ProvideEventRecorder(repositories.ProvideChangeEventRepository(database), clk)))
	timesheets := services.ProvideTimesheetService(repositories.ProvideActualHoursRepository(database), people,
		repositories.ProvideProjectRepository(database), repositories.ProvidePlanRepository(database),
		services.ProvideAuthzService(finder), tx, clk, cfg)
	ctx := request.NewPrincipalContext(context.Background(), seedPlanner)

	template, err := timesheets.Template(ctx, services.TemplateInput{})
	if err != nil {
		return 0, fmt.Errorf("timesheet template: %w", err)
	}
	filled, rows, err := fillActuals(template)
	if err != nil || rows == 0 {
		return 0, err
	}
	result, err := timesheets.Import(ctx, services.ImportInput{File: bytes.NewReader(filled), Size: int64(len(filled))})
	if err != nil {
		return 0, fmt.Errorf("timesheet import: %w", err)
	}
	return result.Created, nil
}

// fillActuals writes actual hours into the template rows of the demo projects that have none
// yet, and returns the new file and how many rows it filled.
func fillActuals(template []byte) ([]byte, int, error) {
	records, err := csv.NewReader(bytes.NewReader(template)).ReadAll()
	if err != nil {
		return nil, 0, fmt.Errorf("read timesheet template: %w", err)
	}
	col := map[string]int{}
	for i, name := range records[0] {
		col[name] = i
	}
	filled := 0
	for _, rec := range records[1:] {
		factor, ok := actualFactors[rec[col[constants.TimesheetColumnProject]]]
		if !ok || rec[col[constants.TimesheetColumnActualHours]] != "" {
			continue
		}
		planned, err := strconv.ParseFloat(rec[col[constants.TimesheetColumnPlannedHours]], 64)
		if err != nil || planned <= 0 {
			continue
		}
		hours := math.Round(planned*factor/actualHourStep) * actualHourStep
		rec[col[constants.TimesheetColumnActualHours]] = strconv.FormatFloat(hours, 'f', -1, 64)
		filled++
	}
	var out bytes.Buffer
	w := csv.NewWriter(&out)
	if err := w.WriteAll(records); err != nil {
		return nil, 0, fmt.Errorf("write timesheet: %w", err)
	}
	return out.Bytes(), filled, nil
}
