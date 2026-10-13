# Timesheet Import Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A planner downloads a CSV template built from the baseline, fills in the hours people actually worked, and imports it; programs and projects then show Actual hours and Burn %, and the member calendar shows "Worked N h".

**Architecture:** A new `actual_hours` table (person × project × ISO week) owned by a new **timesheets** module: one `timesheet.go` file in each layer (`models`, `repositories`, `services`, `dtos`, `handlers`). The service parses the CSV with `encoding/csv`, checks every row before saving anything, and upserts in one `TxManager.Do`. The pure `internal/metrics` package gains Actual hours and Burn %; the insights service feeds it from the new repository. The client gets blob download and multipart upload through `shared/api/http.ts`, a `features/timesheet-import` dialog on the People page, and a Burn bar that falls back to Planned burn.

**Tech Stack:** Go (Echo v5, Uber FX, GORM, Atlas, testify, testcontainers), React + TypeScript (TanStack Query, Base UI via shadcn, Vitest + Testing Library).

**Spec:** [`docs/features/timesheet-import.md`](../features/timesheet-import.md). Read it before starting. The metric formulas are in [root README §7](../../README.md#7-metrics); the API rows are in [server/README §3](../../server/README.md#3-api-restjson).

**Builds on:** [member-calendar](../features/member-calendar.md) is merged first. This plan relies on exactly these names from it:
- server: `models.ScheduleWeek{Week isoweek.Week; Capacity, Load float64; Utilization *float64; Band string; Tentative bool; Projects []ScheduleProject}` in `internal/models/schedule.go`, aliased in services (`type ScheduleWeek = models.ScheduleWeek`) because `dtos` cannot import `services`. Also `ScheduleService.Get(ctx, ScheduleInput{PersonID; From, To}) (*PersonSchedule, error)`, `ProvideScheduleService(plans, people, absences, authz, cfg)`, and `dtos.ScheduleWeekResponse` / `NewScheduleResponse` in `internal/dtos/schedule.go`;
- client: `ScheduleWeek` and the people hooks in `@/shared/people`, the utilization bands in `@/shared/utilization`, and `features/person-calendar` (`PersonCalendar`), whose week column is `components/week-row` (`WeekRow`).

## Global Constraints

- Branch `feat/timesheet-import` from `main`. One PR for the feature; merge only when the user says "merge", with **rebase merge**, keeping the branch, never `--admin`, only after CI passes.
- **Server and client never share a commit.** Small logical commits. Every commit passes its own side's checks:
  - server (`cd server`): `gofmt -l .` prints nothing, then `go build ./... && go vet ./... && go test ./...` and `make lint`
  - client (`cd client`): `npm run format`, then `npx tsc -b && npm run lint && npm test && npm run build`
- Every commit message ends with a blank line and `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- **Never hardcode.** Paths, roles, messages, error codes, CSV column names, settings and file names each have one named definition (`internal/constants`, `internal/config`; on the client `src/resources`, `src/utils`, component `constants.ts`).
- All code comments, docs and UI text are in **English**. Client text goes through `t()` with keys in `src/shared/i18n/language-key.ts`, sorted case-insensitively; use `’` (not `'`) inside keys.
- Load the repo skills before touching an area: `structuring-server-modules`, `writing-go-server-code`, `testing-go-server-code` (server); `structuring-client-components`, `organizing-client-shared-code`, `styling-client-components`, `translating-client-text` (client).
- Routes: `GET /api/v1/timesheets/template?from=&to=` (`text/csv`, a file download) and `POST /api/v1/timesheets/import` (`multipart/form-data`, field `file`). Roles: **planner** only (`TimesheetRoles = []string{RolePlanner}`); a lead or specialist gets `403`.
- Settings (in `internal/config/config.go`, `cmd/server/.env.example`, server/README §8): `TIMESHEET_TEMPLATE_WEEKS` = 4, `TIMESHEET_MAX_WEEKS` = 12, `TIMESHEET_MAX_WEEK_HOURS` = 80, `TIMESHEET_MAX_BYTES` = 1048576 (1 MB), `TIMESHEET_MAX_ROWS` = 5000.
- Template columns, in this order: `person_email`, `person_name`, `program`, `project`, `project_id`, `week`, `planned_hours`, `actual_hours`. An import needs `person_email`, `project_id`, `week`, `actual_hours` in its header, in any order, and ignores the other columns.
- Any failed row: `400` with `error_code` `TIMESHEET_INVALID` and `data.problems = [{ line, column, code }]`; nothing is saved. An empty `actual_hours` skips the row; `0` saves zero hours. Importing again overwrites the same person, project and week.
- An import records **no change event** and never triggers a re-plan.
- Metrics: `actual_hours` = Σ actual hours (JSON `actual_hours`), `burn_percent` = actual ÷ budget × 100 (JSON `burn_percent`). Both are `null` until the project (or, for a program, any of its projects) has actual hours. Planned burn is unchanged.
- Swagger (`server/docs`) and migrations (`cmd/migrations/sql`) are generated, never edited by hand.

## Review Focus

The spec does not mention these inputs, and no happy-path test would catch them. Each one has a test in the task named after it.

1. **A file saved by Excel**: it starts with a UTF-8 byte order mark and uses CRLF line endings. Expected: the header is still recognised and the rows import. → Task 4.
2. **A row shorter than the header**: some tools drop trailing empty cells, so `...,2026-W41,20` has no `actual_hours` cell at all. Expected: the cell counts as empty and the row is skipped. No panic and no `500`. → Task 4.
3. **The same person, project and week twice in one file.** Expected: a `ROW_DUPLICATE` problem. It must never reach Postgres, where `ON CONFLICT ... cannot affect row a second time` would be a `500`. → Task 4.
4. **Email in another case or with spaces** (`" Alex@Example.TEST "`). Expected: it matches the stored lower-case email. → Task 5.
5. **A name that starts with `=`, `+`, `-` or `@`** (a project called `=HYPERLINK(...)`) in the template. Expected: the cell is written with a leading `'`, so a spreadsheet shows it as text instead of running it as a formula. → Task 6.

## File Structure

**Server** (`server/`)

| File | Change | Responsibility |
|---|---|---|
| `internal/metrics/metrics.go` (+ `_test.go`) | modify | `Input.ActualHours`, `Progress.ActualHours`, `Progress.BurnPercent`, `SumOptional` |
| `internal/dtos/metrics.go` | modify | `actual_hours`, `burn_percent` on `ProgressResponse` |
| `internal/models/timesheet.go` | create | `ActualHours` GORM model, table `actual_hours` |
| `cmd/migrations/sql/<timestamp>_actual_hours.sql`, `atlas.sum` | generate | the table and its unique index |
| `internal/repositories/timesheet.go` | create | `ActualHoursRepository`: `Upsert`, `SumByProject`, `ByPersonWeeks`, `ListInRange` |
| `internal/repositories/person.go` | modify | `ListByEmails` |
| `internal/repositories/project.go` | modify | `ListByIDs` |
| `internal/constants/timesheet.go` | create | CSV columns, row problem codes, file name and content type |
| `internal/constants/{routes,roles,error_codes,messages}.go` | modify | paths, `TimesheetRoles`, error codes, messages |
| `internal/config/config.go` (+ `_test.go`), `cmd/server/.env.example` | modify | the five `TIMESHEET_*` settings |
| `internal/services/authz.go` | modify | `CanImportTimesheets` |
| `internal/services/timesheet.go` (+ `_test.go`, `_integration_test.go`) | create | CSV parsing, `TimesheetService.Import` and `Template` |
| `internal/services/insights.go` (+ `metrics_test.go`, `insights_test.go`, `members_test.go`) | modify | read actual hours into project and program metrics |
| `internal/models/schedule.go`, `internal/services/schedule.go` (+ `_test.go`) | modify | `ScheduleWeek.ActualHours`, read through `ByPersonWeeks` |
| `internal/dtos/timesheet.go` | create | `TimesheetTemplateRequest`, `ImportResultResponse` |
| `internal/dtos/schedule.go` | modify | `actual_hours` on `ScheduleWeekResponse` |
| `internal/handlers/timesheet.go` (+ `_test.go`) | create | CSV download and multipart upload |
| `cmd/server/routes/router.go` (+ `_test.go`), `cmd/server/main.go` | modify | routes with roles, FX providers |
| `docs/` (Swagger) | generate | `make swagger-load` |

**Client** (`client/src/`)

| File | Change | Responsibility |
|---|---|---|
| `shared/api/blob.ts` (+ `blob.test.ts`) | create | read a JSON error envelope that arrived as a `Blob` |
| `shared/api/http.ts` | modify | pass `Blob` responses through, unwrap blob errors |
| `utils/functions/file.ts` (+ `file.test.ts`), `utils/functions/index.ts` | create / modify | `saveFile(blob, name)` |
| `shared/programs/types.ts`, `shared/programs/utils.ts` (+ `utils.test.ts`), `shared/programs/index.ts` | modify | `actual_hours`, `burn_percent`, `burnMetric()` |
| `features/program-overview/components/metrics-card/component.tsx` | modify | Burn bar, Actual figure |
| `features/program-projects/components/project-row/component.tsx` | modify | Burn bar |
| `features/risks/component.test.tsx` | modify | fixture gains the two new fields |
| `resources/endpoints/timesheets.ts`, `resources/endpoints/index.ts` | create / modify | `TIMESHEET_ENDPOINTS` |
| `resources/rbac/permissions.ts` | modify | `'timesheets:import'` (planner) |
| `features/timesheet-import/` | create | `TimesheetImport`: button, dialog, problem table, API and hooks |
| `features/people/{types.ts,component.tsx}` | modify | `actions` slot next to **Add person** |
| `pages/people/component.tsx` | modify | puts `TimesheetImport` in that slot for planners |
| `shared/people/types.ts` | modify | `ScheduleWeek.actual_hours` |
| `features/person-calendar/components/week-row/components/worked-hours/` | create | "Worked N h" under a week's load |
| `features/person-calendar/components/week-row/component.tsx`, `features/person-calendar/component.test.tsx` | modify | render it; fixtures gain `actual_hours` |
| `shared/i18n/language-key.ts` | modify | new keys |

**Docs:** `docs/features/timesheet-import.md` (status Done, tasks ticked) and `docs/features/README.md` (row 12 Done) in the last task.

---

## Server

Work in `server/`. Before Task 1: `git checkout main && git pull --rebase && git checkout -b feat/timesheet-import`.

### Task 1: Actual hours and Burn % in the metric formulas

**Files:**
- Modify: `server/internal/metrics/metrics.go`
- Modify: `server/internal/metrics/metrics_test.go`
- Modify: `server/internal/dtos/metrics.go`
- Create: `server/internal/dtos/metrics_test.go`
- Generate: `server/docs/*` (Swagger)

**Interfaces:**
- Produces: `metrics.Input.ActualHours *float64`; `metrics.Progress.ActualHours *float64`, `metrics.Progress.BurnPercent *float64`; `metrics.SumOptional(a, b *float64) *float64`; `dtos.ProgressResponse.ActualHours` (`actual_hours`) and `.BurnPercent` (`burn_percent`).

- [ ] **Step 1: Write the failing tests** — append to `server/internal/metrics/metrics_test.go`:

```go
func TestCompute_ActualHours(t *testing.T) {
	base := Input{BudgetHours: 400, PlannedHours: 320, PlannedToDateHours: 120}

	t.Run("no actual hours leaves actual and burn unknown", func(t *testing.T) {
		p := Compute(base, settings)
		assert.Nil(t, p.ActualHours)
		assert.Nil(t, p.BurnPercent)
		assert.InDelta(t, 30, *p.PlannedBurnPercent, 1e-9, "planned burn is unchanged")
	})

	t.Run("actual hours give burn against the budget", func(t *testing.T) {
		in := base
		actual := 100.0
		in.ActualHours = &actual
		p := Compute(in, settings)
		require.NotNil(t, p.ActualHours)
		assert.InDelta(t, 100, *p.ActualHours, 1e-9)
		require.NotNil(t, p.BurnPercent)
		assert.InDelta(t, 25, *p.BurnPercent, 1e-9, "100 of 400 h")
	})

	t.Run("zero actual hours are known, so burn is 0 %", func(t *testing.T) {
		in := base
		zero := 0.0
		in.ActualHours = &zero
		p := Compute(in, settings)
		require.NotNil(t, p.BurnPercent)
		assert.InDelta(t, 0, *p.BurnPercent, 1e-9)
	})

	t.Run("actual hours without a budget have no burn", func(t *testing.T) {
		in := base
		in.BudgetHours = 0
		actual := 8.0
		in.ActualHours = &actual
		p := Compute(in, settings)
		assert.InDelta(t, 8, *p.ActualHours, 1e-9)
		assert.Nil(t, p.BurnPercent)
	})
}

func TestSumOptional(t *testing.T) {
	five, three := 5.0, 3.0
	assert.Nil(t, SumOptional(nil, nil))
	assert.InDelta(t, 5, *SumOptional(&five, nil), 1e-9)
	assert.InDelta(t, 3, *SumOptional(nil, &three), 1e-9)
	assert.InDelta(t, 8, *SumOptional(&five, &three), 1e-9)
}
```

Create `server/internal/dtos/metrics_test.go`:

```go
package dtos

import (
	"encoding/json"
	"testing"

	"the-slot-finder/server/internal/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProgressResponse_ActualHours(t *testing.T) {
	actual, burn := 100.0, 25.0
	got := NewProgressResponse(metrics.Progress{ActualHours: &actual, BurnPercent: &burn})
	require.NotNil(t, got.ActualHours)
	assert.InDelta(t, 100, *got.ActualHours, 1e-9)
	require.NotNil(t, got.BurnPercent)
	assert.InDelta(t, 25, *got.BurnPercent, 1e-9)

	body, err := json.Marshal(NewProgressResponse(metrics.Progress{}))
	require.NoError(t, err)
	assert.Contains(t, string(body), `"actual_hours":null`)
	assert.Contains(t, string(body), `"burn_percent":null`)
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./internal/metrics ./internal/dtos -run 'ActualHours|SumOptional' -v`
Expected: build failure, `in.ActualHours undefined (type Input has no field or method ActualHours)`.

- [ ] **Step 3: Implement** — in `server/internal/metrics/metrics.go`:

Add to `Input`, after `PlannedToDateHours`:

```go
	// ActualHours is the sum of the hours imported from timesheets; nil while none exist.
	ActualHours *float64
```

Add to `Progress`, after `PlannedToDateHours`:

```go
	// ActualHours and BurnPercent (actual ÷ budget) are nil until actual hours exist.
	ActualHours *float64
	BurnPercent *float64
```

In `Compute`, right after the `p := Progress{...}` literal:

```go
	if in.ActualHours != nil {
		actual := *in.ActualHours
		p.ActualHours = &actual
		p.BurnPercent = Percent(actual, in.BudgetHours)
	}
```

Add after `Percent`:

```go
// SumOptional adds two hours that may be unknown (nil). The sum is nil only when both are,
// so a program has actual hours as soon as one of its projects has.
func SumOptional(a, b *float64) *float64 {
	if a == nil && b == nil {
		return nil
	}
	var sum float64
	if a != nil {
		sum += *a
	}
	if b != nil {
		sum += *b
	}
	return &sum
}
```

In `server/internal/dtos/metrics.go`, add to `ProgressResponse` after `PlannedToDateHours`:

```go
	// ActualHours and BurnPercent are null until timesheets recorded actual hours.
	ActualHours *float64 `json:"actual_hours" example:"100"`
	BurnPercent *float64 `json:"burn_percent" example:"25"`
```

and in `NewProgressResponse` add `ActualHours: p.ActualHours, BurnPercent: p.BurnPercent,` to the literal (after `PlannedToDateHours: p.PlannedToDateHours,`).

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/metrics ./internal/dtos -v -run 'Compute|SumOptional|ProgressResponse'`
Expected: PASS.

- [ ] **Step 5: Regenerate Swagger and check the side**

Run: `make swagger-load` (without `swag` on the PATH: `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d ./cmd/server,./internal/handlers,./internal/middlewares,./internal/services,./internal/repositories,./internal/models,./internal/utils,./internal/config,./internal/constants,./internal/dtos,./internal/logger,./internal/db --output ./docs`), then `gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`.
Expected: `gofmt` prints nothing; tests PASS; lint `0 issues`. `git diff --stat docs` shows `actual_hours` and `burn_percent` added to `dtos.ProgressResponse`.

- [ ] **Step 6: Commit**

```bash
git add internal/metrics internal/dtos/metrics.go internal/dtos/metrics_test.go docs
git commit -m "feat(server): actual hours and burn in the metric formulas

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: The `actual_hours` table and its repository

**Files:**
- Create: `server/internal/models/timesheet.go`
- Generate: `server/cmd/migrations/sql/<timestamp>_actual_hours.sql`, `server/cmd/migrations/sql/atlas.sum`
- Create: `server/internal/repositories/timesheet.go`
- Modify: `server/cmd/server/main.go` (provide the repository)
- Create: `server/internal/services/timesheet_integration_test.go`

**Interfaces:**
- Produces:
  - `models.ActualHours{BaseModel; PersonID, ProjectID string; Week isoweek.Week; Hours float64; UpdatedBy string; Person *Person; Project *Project}`, table `actual_hours`, unique on `person_id, project_id, week_start`. `UpdatedBy` holds the planner's **Keycloak subject** (like `AuditEvent.ActorSubject`), because a planner may have no person record.
  - `repositories.ActualHoursRepository` with `Upsert(ctx, rows []models.ActualHours) (created, updated int, err error)`, `SumByProject(ctx, projectIDs []string) (map[string]float64, error)`, `ByPersonWeeks(ctx, personID string, from, to isoweek.Week) (map[isoweek.Week]float64, error)`, `ListInRange(ctx, from, to isoweek.Week) ([]models.ActualHours, error)`; `repositories.ProvideActualHoursRepository(*db.PostgresDB)`.
  - Test helper `timesheetFixture(t, pdb) (models.Person, models.Project)` (integration build tag).

- [ ] **Step 1: Write the model** — create `server/internal/models/timesheet.go`:

```go
package models

import "the-slot-finder/server/internal/isoweek"

// ActualHours is the hours a person really worked on a project in one ISO week, imported
// from a timesheet. It feeds Actual hours and Burn %; planning and capacity never read it.
type ActualHours struct {
	BaseModel
	PersonID  string       `gorm:"column:person_id;type:uuid;not null;uniqueIndex:idx_actual_hours_person_project_week"`
	ProjectID string       `gorm:"column:project_id;type:uuid;not null;uniqueIndex:idx_actual_hours_person_project_week;index"`
	Week      isoweek.Week `gorm:"column:week_start;not null;uniqueIndex:idx_actual_hours_person_project_week;index"`
	Hours     float64      `gorm:"column:hours;type:numeric(6,2);not null"`
	// UpdatedBy is the Keycloak subject of the planner whose import set the value last: a
	// subject, not a person id, because a planner may have no person record.
	UpdatedBy string `gorm:"column:updated_by;not null"`

	Person  *Person  `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
	Project *Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// TableName keeps the table name "actual_hours" as it reads, without another plural "s".
func (ActualHours) TableName() string { return "actual_hours" }
```

- [ ] **Step 2: Generate the migration**

Run (needs Docker running and the Atlas CLI, `brew install ariga/tap/atlas`): `make migrate-generate name=actual_hours`
Expected: a new `cmd/migrations/sql/<timestamp>_actual_hours.sql` holding `CREATE TABLE "public"."actual_hours"` with `"week_start" date NOT NULL`, `"hours" numeric(6,2) NOT NULL`, `"updated_by" text NOT NULL`, two `ON DELETE CASCADE` foreign keys, and `CREATE UNIQUE INDEX "idx_actual_hours_person_project_week" ON "public"."actual_hours" ("person_id", "project_id", "week_start")`; `atlas.sum` updated. Do not edit either file by hand. Then `make migrate-up` against the local database.

- [ ] **Step 3: Write the failing integration test** — create `server/internal/services/timesheet_integration_test.go`:

```go
//go:build integration

package services

import (
	"context"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// timesheetFixture adds Ada and the Checkout project of the Acme program.
func timesheetFixture(t *testing.T, pdb *db.PostgresDB) (models.Person, models.Project) {
	t.Helper()
	ada := models.Person{BaseModel: models.NewBaseModel(), Name: "Ada", Email: "ada@example.test", Role: constants.RoleSpecialist,
		WeeklyHours: 40, NonProjectHours: 4, Active: true}
	require.NoError(t, pdb.Create(&ada).Error)
	program := models.Program{BaseModel: models.NewBaseModel(), Name: "Acme", Client: "Acme", Status: constants.ProgramStatusOpen,
		StartDate: day(time.October, 5), EndDate: day(time.December, 31), Objectives: []string{}}
	require.NoError(t, pdb.Create(&program).Error)
	project := models.Project{BaseModel: models.NewBaseModel(), ProgramID: program.ID, Name: "Checkout", Priority: 1,
		Status: constants.ProjectStatusActive, StartEarliest: day(time.October, 5), Deadline: day(time.December, 11)}
	require.NoError(t, pdb.Create(&project).Error)
	return ada, project
}

func TestActualHoursRepository_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	ctx := context.Background()
	ada, project := timesheetFixture(t, pdb)
	repo := repositories.ProvideActualHoursRepository(pdb)
	w41, w42 := mustWeek(t, "2026-W41"), mustWeek(t, "2026-W42")

	created, updated, err := repo.Upsert(ctx, []models.ActualHours{
		{PersonID: ada.ID, ProjectID: project.ID, Week: w41, Hours: 30, UpdatedBy: "kc-planner"},
		{PersonID: ada.ID, ProjectID: project.ID, Week: w42, Hours: 0, UpdatedBy: "kc-planner"},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, created)
	assert.Zero(t, updated)

	created, updated, err = repo.Upsert(ctx, []models.ActualHours{
		{PersonID: ada.ID, ProjectID: project.ID, Week: w41, Hours: 32.5, UpdatedBy: "kc-planner"},
	})
	require.NoError(t, err)
	assert.Zero(t, created)
	assert.Equal(t, 1, updated, "the same person, project and week is replaced")

	sums, err := repo.SumByProject(ctx, []string{project.ID, uuid.NewString()})
	require.NoError(t, err)
	assert.Len(t, sums, 1, "a project without rows is missing from the map")
	assert.InDelta(t, 32.5, sums[project.ID], hoursEpsilon)

	weeks, err := repo.ByPersonWeeks(ctx, ada.ID, w41, w42)
	require.NoError(t, err)
	assert.InDelta(t, 32.5, weeks[w41], hoursEpsilon)
	zero, recorded := weeks[w42]
	assert.True(t, recorded, "zero hours are recorded, not missing")
	assert.Zero(t, zero)

	rows, err := repo.ListInRange(ctx, w41, w41)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "Ada", rows[0].Person.Name)
	assert.Equal(t, "Acme", rows[0].Project.Program.Name)
}
```

- [ ] **Step 4: Run it to verify it fails**

Run: `go test -tags=integration ./internal/services -run TestActualHoursRepository_Postgres -v`
Expected: build failure, `undefined: repositories.ProvideActualHoursRepository`.

- [ ] **Step 5: Implement the repository** — create `server/internal/repositories/timesheet.go`:

```go
package repositories

import (
	"context"

	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"gorm.io/gorm/clause"
)

// actualHoursBatchSize keeps one upsert statement far below Postgres' 65,535 parameters
// (9 columns per row).
const actualHoursBatchSize = 1000

// Relations of an actual hours row: the person, and the project with its program.
const (
	actualHoursPersonRel      = "Person"
	actualHoursProjectRel     = "Project"
	actualHoursProgramPreload = "Project.Program"
)

// ActualHoursRepository stores the hours people really worked, per person, project and ISO
// week. It belongs to the timesheets module.
type ActualHoursRepository interface {
	// Upsert saves each row, replacing the hours of a person, project and week that is
	// already stored. It returns how many rows were new and how many replaced a value. The
	// rows must not repeat a person, project and week.
	Upsert(ctx context.Context, rows []models.ActualHours) (created, updated int, err error)
	// SumByProject returns the actual hours of each project that has any; a project without
	// rows is missing from the map.
	SumByProject(ctx context.Context, projectIDs []string) (map[string]float64, error)
	// ByPersonWeeks returns one person's actual hours per week over all projects, from
	// through to; a week without rows is missing from the map.
	ByPersonWeeks(ctx context.Context, personID string, from, to isoweek.Week) (map[isoweek.Week]float64, error)
	// ListInRange returns every row from through to, with the person and the project and its
	// program.
	ListInRange(ctx context.Context, from, to isoweek.Week) ([]models.ActualHours, error)
}

type actualHoursRepository struct {
	db *db.PostgresDB
}

// ProvideActualHoursRepository creates the actual hours repository.
func ProvideActualHoursRepository(db *db.PostgresDB) ActualHoursRepository {
	return &actualHoursRepository{db: db}
}

func (r *actualHoursRepository) Upsert(ctx context.Context, rows []models.ActualHours) (int, int, error) {
	if len(rows) == 0 {
		return 0, 0, nil
	}
	keys := make([][]any, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, []any{row.PersonID, row.ProjectID, row.Week})
	}
	var existing int64
	if err := r.db.From(ctx).Model(&models.ActualHours{}).
		Where("(person_id, project_id, week_start) IN ?", keys).Count(&existing).Error; err != nil {
		return 0, 0, err
	}
	err := r.db.From(ctx).Omit(actualHoursPersonRel, actualHoursProjectRel).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "person_id"}, {Name: "project_id"}, {Name: "week_start"}},
		DoUpdates: clause.AssignmentColumns([]string{"hours", "updated_by", "updated_at"}),
	}).CreateInBatches(&rows, actualHoursBatchSize).Error
	if err != nil {
		return 0, 0, err
	}
	return len(rows) - int(existing), int(existing), nil
}

// projectHours is one row of SumByProject.
type projectHours struct {
	ProjectID string
	Hours     float64
}

func (r *actualHoursRepository) SumByProject(ctx context.Context, projectIDs []string) (map[string]float64, error) {
	out := make(map[string]float64, len(projectIDs))
	if len(projectIDs) == 0 {
		return out, nil
	}
	var sums []projectHours
	err := r.db.From(ctx).Model(&models.ActualHours{}).Select("project_id, SUM(hours) AS hours").
		Where("project_id IN ?", projectIDs).Group("project_id").Scan(&sums).Error
	if err != nil {
		return nil, err
	}
	for _, s := range sums {
		out[s.ProjectID] = s.Hours
	}
	return out, nil
}

// weekHours is one row of ByPersonWeeks.
type weekHours struct {
	Week  isoweek.Week `gorm:"column:week_start"`
	Hours float64
}

func (r *actualHoursRepository) ByPersonWeeks(ctx context.Context, personID string, from, to isoweek.Week) (map[isoweek.Week]float64, error) {
	var sums []weekHours
	err := r.db.From(ctx).Model(&models.ActualHours{}).Select("week_start, SUM(hours) AS hours").
		Where("person_id = ? AND week_start BETWEEN ? AND ?", personID, from, to).Group("week_start").Scan(&sums).Error
	if err != nil {
		return nil, err
	}
	out := make(map[isoweek.Week]float64, len(sums))
	for _, s := range sums {
		out[s.Week] = s.Hours
	}
	return out, nil
}

func (r *actualHoursRepository) ListInRange(ctx context.Context, from, to isoweek.Week) ([]models.ActualHours, error) {
	var rows []models.ActualHours
	err := r.db.From(ctx).Preload(actualHoursPersonRel).Preload(actualHoursProgramPreload).
		Where("week_start BETWEEN ? AND ?", from, to).
		Order("week_start ASC").Order("person_id ASC").Order("project_id ASC").Find(&rows).Error
	return rows, err
}
```

In `server/cmd/server/main.go`, add `repositories.ProvideActualHoursRepository,` to `fx.Provide` after `repositories.ProvideChangeEventRepository,`.

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test -tags=integration ./internal/services -run TestActualHoursRepository_Postgres -v`, then `gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`
Expected: PASS; `cmd/server` `TestAppOptions_Validate` (FX `ValidateApp`) still passes; lint `0 issues`.

- [ ] **Step 7: Commit**

```bash
git add internal/models/timesheet.go internal/repositories/timesheet.go cmd/migrations/sql cmd/server/main.go internal/services/timesheet_integration_test.go
git commit -m "feat(server): actual hours table and repository

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: Programs and projects report actual hours and burn

**Files:**
- Modify: `server/internal/services/insights.go` (struct, `ProvideInsightsService`, `metricsData`, `projectInput`, `ProgramOverview`, `ProjectMetrics`, `Risks`)
- Create: `server/internal/services/timesheet_test.go` (the `MockActualHoursRepository`)
- Modify: `server/internal/services/metrics_test.go`, `server/internal/services/insights_test.go`, `server/internal/services/members_test.go`

**Interfaces:**
- Consumes: `repositories.ActualHoursRepository.SumByProject` (Task 2); `metrics.Input.ActualHours`, `metrics.SumOptional` (Task 1).
- Produces: `ProvideInsightsService(plans, people, absences, programs, projects, actuals repositories.ActualHoursRepository, authz, clk, cfg)` — `actuals` goes right after `projects`. If member-calendar added parameters, keep them where they are and insert `actuals` after `projects`. Test helpers `MockActualHoursRepository` and `newMetricsServiceWithActuals(t, actual map[string]float64) InsightsService`.

- [ ] **Step 1: Write the mock and the failing test** — create `server/internal/services/timesheet_test.go`:

```go
package services

import (
	"context"

	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockActualHoursRepository is a testify mock of repositories.ActualHoursRepository.
type MockActualHoursRepository struct {
	mock.Mock
}

func (m *MockActualHoursRepository) Upsert(ctx context.Context, rows []models.ActualHours) (int, int, error) {
	args := m.Called(ctx, rows)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockActualHoursRepository) SumByProject(ctx context.Context, projectIDs []string) (map[string]float64, error) {
	args := m.Called(ctx, projectIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockActualHoursRepository) ByPersonWeeks(ctx context.Context, personID string, from, to isoweek.Week) (map[isoweek.Week]float64, error) {
	args := m.Called(ctx, personID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[isoweek.Week]float64), args.Error(1)
}

func (m *MockActualHoursRepository) ListInRange(ctx context.Context, from, to isoweek.Week) ([]models.ActualHours, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).([]models.ActualHours), args.Error(1)
}
```

In `server/internal/services/metrics_test.go`, replace `newMetricsService` with:

```go
func newMetricsService(t *testing.T) InsightsService {
	return newMetricsServiceWithActuals(t, map[string]float64{})
}

// newMetricsServiceWithActuals is newMetricsService with these actual hours per project id.
func newMetricsServiceWithActuals(t *testing.T, actual map[string]float64) InsightsService {
	plans, programs, projects, people, absences := metricsFixture(t)
	actuals := new(MockActualHoursRepository)
	actuals.On("SumByProject", mock.Anything, mock.Anything).Return(actual, nil)
	cfg := heatmapConfig()
	cfg.ScopeLagThreshold, cfg.AlertHorizonWeeks = 20, 4
	return ProvideInsightsService(plans, people, absences, programs, projects, actuals, ProvideAuthzService(leadPeople{}),
		clock.NewFixed(time.Date(2026, time.November, 15, 9, 0, 0, 0, time.UTC)), cfg)
}
```

and append:

```go
func TestInsightsService_ActualHours(t *testing.T) {
	lead := asRole("kc-lead", constants.RoleLead)

	t.Run("without actual hours, actual and burn stay unknown", func(t *testing.T) {
		overview, err := newMetricsService(t).ProgramOverview(lead, metricsProgramID)
		require.NoError(t, err)
		assert.Nil(t, overview.Metrics.ActualHours)
		assert.Nil(t, overview.Metrics.BurnPercent)
	})

	t.Run("a project's actual hours give its burn and the program's", func(t *testing.T) {
		svc := newMetricsServiceWithActuals(t, map[string]float64{checkoutID: 100})
		byID, err := svc.ProjectMetrics(lead, metricsProgramID)
		require.NoError(t, err)
		require.NotNil(t, byID[checkoutID].BurnPercent)
		assert.InDelta(t, 25, *byID[checkoutID].BurnPercent, 1e-9, "100 of Checkout's 400 h")
		assert.Nil(t, byID[loyaltyID].ActualHours, "Loyalty has no actual hours")

		overview, err := svc.ProgramOverview(lead, metricsProgramID)
		require.NoError(t, err)
		require.NotNil(t, overview.Metrics.ActualHours)
		assert.InDelta(t, 100, *overview.Metrics.ActualHours, 1e-9)
		assert.InDelta(t, 25, *overview.Metrics.BurnPercent, 1e-9, "100 of the program's 400 h")
	})
}
```

In `insights_test.go` (`newInsightsService`) and the two `ProvideInsightsService` calls in `members_test.go`, add a `nil` for `actuals` right after the `projects` argument, e.g. `ProvideInsightsService(plans, people, absences, nil, nil, nil, ProvideAuthzService(finder), ...)`. Heatmap and members never read actual hours.

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/services -run 'TestInsightsService' -v`
Expected: build failure, `too many arguments in call to ProvideInsightsService`.

- [ ] **Step 3: Implement** — in `server/internal/services/insights.go`:

Add the field `actuals repositories.ActualHoursRepository` to `insightsService` after `projects`, the parameter `actuals repositories.ActualHoursRepository` to `ProvideInsightsService` after `projects`, and `actuals: actuals,` to the literal it returns.

Replace `metricsData` with:

```go
// metricsData loads the projects (one program, or all), the baseline's planned hours and the
// actual hours of each project that has any.
func (s *insightsService) metricsData(ctx context.Context, programID, op string) ([]models.Project, map[string]plannedHours, map[string]float64, time.Time, error) {
	projects, err := s.projects.ListForMetrics(ctx, programID)
	if err != nil {
		return nil, nil, nil, time.Time{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceMetrics)
	}
	baseline, err := findBaseline(ctx, s.plans, op, true)
	if err != nil {
		return nil, nil, nil, time.Time{}, err
	}
	ids := make([]string, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
	}
	actual, err := s.actuals.SumByProject(ctx, ids)
	if err != nil {
		return nil, nil, nil, time.Time{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceMetrics)
	}
	today := s.today()
	return projects, plannedByDemand(baseline.Assignments, isoweek.Of(today)), actual, today, nil
}
```

Change `projectInput` to take the actual hours and set them (signature and the lines before `return in`):

```go
// projectInput is a project's metric input from its demand, the baseline and its actual hours.
func (s *insightsService) projectInput(p models.Project, planned map[string]plannedHours, actual map[string]float64, today time.Time) metrics.Input {
	// ... the existing body up to the demand loop stays as it is ...
	if hours, ok := actual[p.ID]; ok {
		in.ActualHours = &hours
	}
	return in
}
```

Update the three callers:
- `ProgramOverview`: `projects, planned, actual, today, err := s.metricsData(ctx, programID, opProgramOverview)`; inside the loop `pin := s.projectInput(p, planned, actual, today)` and, after `in.PlannedToDateHours += pin.PlannedToDateHours`, add `in.ActualHours = metrics.SumOptional(in.ActualHours, pin.ActualHours)`.
- `ProjectMetrics`: `projects, planned, actual, today, err := ...` and `metrics.Compute(s.projectInput(p, planned, actual, today), s.settings())`.
- `Risks`: the same two changes, so a risk row carries the same burn as the projects tab.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/services -run 'TestInsightsService|TestProgramMembers' -v`, then `gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`
Expected: PASS, including the existing overview, project metrics and risks tests, unchanged.

- [ ] **Step 5: Commit**

```bash
git add internal/services/insights.go internal/services/timesheet_test.go internal/services/metrics_test.go internal/services/insights_test.go internal/services/members_test.go
git commit -m "feat(server): programs and projects report actual hours and burn

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 4: Read and check a timesheet CSV

**Files:**
- Create: `server/internal/constants/timesheet.go`
- Create: `server/internal/services/timesheet.go` (parsing only in this task)
- Modify: `server/internal/services/timesheet_test.go`

**Interfaces:**
- Produces:
  - `services.RowProblem{Line int; Column, Code string}` (JSON `line`, `column`, `code`)
  - `parseTimesheet(data []byte, limits timesheetLimits) ([]timesheetRow, []RowProblem, error)`, where `timesheetLimits{Current isoweek.Week; MaxHours float64; MaxRows int}`; `timesheetRow{Line int; Email, ProjectID string; Week isoweek.Week; Hours *float64}` holds the email in lower case and the project id as a canonical UUID; `errTooManyRows`
  - every constant in `constants/timesheet.go`
  - test helpers `tsProjectID`, `tsHeader`, `tsLine(week, actual)`, `tsLimits(t)`, `hoursOf(v)`

- [ ] **Step 1: Add the constants** — create `server/internal/constants/timesheet.go`:

```go
package constants

// Columns of the timesheet CSV (docs/features/timesheet-import.md). The template writes them in
// this order; an import reads them by name, in any order, and ignores the others.
const (
	TimesheetColumnPersonEmail  = "person_email"
	TimesheetColumnPersonName   = "person_name"
	TimesheetColumnProgram      = "program"
	TimesheetColumnProject      = "project"
	TimesheetColumnProjectID    = "project_id"
	TimesheetColumnWeek         = "week"
	TimesheetColumnPlannedHours = "planned_hours"
	TimesheetColumnActualHours  = "actual_hours"
)

// TimesheetColumns is the template's header, in order.
var TimesheetColumns = []string{TimesheetColumnPersonEmail, TimesheetColumnPersonName, TimesheetColumnProgram,
	TimesheetColumnProject, TimesheetColumnProjectID, TimesheetColumnWeek, TimesheetColumnPlannedHours,
	TimesheetColumnActualHours}

// TimesheetRequiredColumns must be in the header of an imported file.
var TimesheetRequiredColumns = []string{TimesheetColumnPersonEmail, TimesheetColumnProjectID, TimesheetColumnWeek,
	TimesheetColumnActualHours}

// Problem codes of a timesheet import, each reported as { line, column, code }.
const (
	TimesheetColumnMissing  = "COLUMN_MISSING"
	TimesheetCSVMalformed   = "CSV_MALFORMED"
	TimesheetPersonUnknown  = "PERSON_UNKNOWN"
	TimesheetProjectUnknown = "PROJECT_UNKNOWN"
	TimesheetWeekInvalid    = "WEEK_INVALID"
	TimesheetWeekFuture     = "WEEK_IN_FUTURE"
	TimesheetHoursInvalid   = "HOURS_INVALID"
	TimesheetRowDuplicate   = "ROW_DUPLICATE"
)

// TimesheetProblemsKey holds the problems in the data of a TIMESHEET_INVALID response.
const TimesheetProblemsKey = "problems"

// The timesheet file: the multipart field of an upload, and the download's type and name.
const (
	TimesheetFormField                 = "file"
	TimesheetContentType               = "text/csv; charset=utf-8"
	TimesheetFileName                  = "timesheet.csv"
	ContentDispositionAttachmentFormat = "attachment; filename=%q"
)

// A template cell that starts with one of TimesheetFormulaPrefixes would run as a formula in a
// spreadsheet; it is written with TimesheetTextPrefix in front so it shows as text.
const (
	TimesheetFormulaPrefixes = "=+-@\t\r"
	TimesheetTextPrefix      = "'"
)
```

- [ ] **Step 2: Write the failing tests** — add to `server/internal/services/timesheet_test.go` (merge the imports: `strings`, `testing`, `the-slot-finder/server/internal/constants`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`):

```go
const (
	tsProjectID = "0192f3a4-0000-7000-8000-00000000e101"
	tsHeader    = "person_email,person_name,program,project,project_id,week,planned_hours,actual_hours\n"
)

// tsLine is one template line for Ada on the test project.
func tsLine(week, actual string) string {
	return "ada@example.test,Ada,Acme,Checkout," + tsProjectID + "," + week + ",20," + actual + "\n"
}

// tsLimits allow weeks up to 2026-W42, 80 hours a row and 3 rows.
func tsLimits(t *testing.T) timesheetLimits {
	t.Helper()
	return timesheetLimits{Current: mustWeek(t, "2026-W42"), MaxHours: 80, MaxRows: 3}
}

func hoursOf(v float64) *float64 { return &v }

func TestParseTimesheet_Rows(t *testing.T) {
	w40, w41, w42 := mustWeek(t, "2026-W40"), mustWeek(t, "2026-W41"), mustWeek(t, "2026-W42")
	tests := []struct {
		name string
		file string
		want []timesheetRow
	}{
		{
			name: "a filled-in row, an empty one and zero hours",
			file: tsHeader + tsLine("2026-W40", "18.5") + tsLine("2026-W41", "") + tsLine("2026-W42", "0"),
			want: []timesheetRow{
				{Line: 2, Email: "ada@example.test", ProjectID: tsProjectID, Week: w40, Hours: hoursOf(18.5)},
				{Line: 3, Email: "ada@example.test", ProjectID: tsProjectID, Week: w41},
				{Line: 4, Email: "ada@example.test", ProjectID: tsProjectID, Week: w42, Hours: hoursOf(0)},
			},
		},
		{
			name: "saved by Excel: byte order mark, CRLF, other column order, case and spaces",
			file: "﻿Week, Actual_Hours ,PROJECT_ID,person_email\r\n2026-W41,7.5, " + strings.ToUpper(tsProjectID) +
				" , Ada@Example.TEST \r\n",
			want: []timesheetRow{{Line: 2, Email: "ada@example.test", ProjectID: tsProjectID, Week: w41, Hours: hoursOf(7.5)}},
		},
		{
			name: "a row shorter than the header has an empty actual_hours",
			file: tsHeader + "ada@example.test,Ada,Acme,Checkout," + tsProjectID + ",2026-W41,20\n",
			want: []timesheetRow{{Line: 2, Email: "ada@example.test", ProjectID: tsProjectID, Week: w41}},
		},
		{name: "a header only", file: tsHeader},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, problems, err := parseTimesheet([]byte(tt.file), tsLimits(t))
			require.NoError(t, err)
			assert.Empty(t, problems)
			assert.Equal(t, tt.want, rows)
		})
	}
}

func TestParseTimesheet_Problems(t *testing.T) {
	missing := func(columns ...string) []RowProblem {
		out := make([]RowProblem, 0, len(columns))
		for _, column := range columns {
			out = append(out, RowProblem{Line: 1, Column: column, Code: constants.TimesheetColumnMissing})
		}
		return out
	}
	hours := func(line int) RowProblem {
		return RowProblem{Line: line, Column: constants.TimesheetColumnActualHours, Code: constants.TimesheetHoursInvalid}
	}
	tests := []struct {
		name string
		file string
		want []RowProblem
	}{
		{name: "an empty file", file: "", want: missing(constants.TimesheetRequiredColumns...)},
		{
			name: "a semicolon-separated file",
			file: "person_email;project_id;week;actual_hours\n",
			want: missing(constants.TimesheetRequiredColumns...),
		},
		{
			name: "a header without actual_hours",
			file: "person_email,project_id,week\n",
			want: missing(constants.TimesheetColumnActualHours),
		},
		{
			name: "every check of one row",
			file: tsHeader + ",Ada,Acme,Checkout,not-a-uuid,2026-W99,20,abc\n",
			want: []RowProblem{
				{Line: 2, Column: constants.TimesheetColumnPersonEmail, Code: constants.TimesheetPersonUnknown},
				{Line: 2, Column: constants.TimesheetColumnProjectID, Code: constants.TimesheetProjectUnknown},
				{Line: 2, Column: constants.TimesheetColumnWeek, Code: constants.TimesheetWeekInvalid},
				hours(2),
			},
		},
		{
			name: "a future week, too many hours and a comma decimal",
			file: tsHeader + tsLine("2026-W43", "8") + tsLine("2026-W41", "80.5") + tsLine("2026-W40", `"7,5"`),
			want: []RowProblem{
				{Line: 2, Column: constants.TimesheetColumnWeek, Code: constants.TimesheetWeekFuture},
				hours(3),
				hours(4),
			},
		},
		{
			name: "negative and not-a-number hours",
			file: tsHeader + tsLine("2026-W41", "-1") + tsLine("2026-W40", "NaN"),
			want: []RowProblem{hours(2), hours(3)},
		},
		{
			name: "the same person, project and week twice, in another case",
			file: tsHeader + tsLine("2026-W41", "8") + strings.Replace(tsLine("2026-W41", ""), "ada@", "ADA@", 1),
			want: []RowProblem{{Line: 3, Code: constants.TimesheetRowDuplicate}},
		},
		{
			name: "a broken quote",
			file: tsHeader + "\"ada@example.test,Ada\n",
			want: []RowProblem{{Line: 2, Code: constants.TimesheetCSVMalformed}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, problems, err := parseTimesheet([]byte(tt.file), tsLimits(t))
			require.NoError(t, err)
			assert.Equal(t, tt.want, problems)
		})
	}
}

func TestParseTimesheet_TooManyRows(t *testing.T) {
	file := tsHeader + tsLine("2026-W38", "1") + tsLine("2026-W39", "1") + tsLine("2026-W40", "1") + tsLine("2026-W41", "1")
	_, _, err := parseTimesheet([]byte(file), tsLimits(t))
	assert.ErrorIs(t, err, errTooManyRows)
}
```

- [ ] **Step 3: Run them to verify they fail**

Run: `go test ./internal/services -run TestParseTimesheet -v`
Expected: build failure, `undefined: parseTimesheet`.

- [ ] **Step 4: Implement** — create `server/internal/services/timesheet.go`:

```go
package services

import (
	"bytes"
	"encoding/csv"
	stderrors "errors"
	"io"
	"math"
	"strconv"
	"strings"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/isoweek"

	"github.com/google/uuid"
)

// utf8BOM starts a CSV file that Excel saves as UTF-8; it is dropped before reading.
const utf8BOM = "﻿"

// headerLine is the file line of the CSV header.
const headerLine = 1

// errTooManyRows stops reading a timesheet with more than TIMESHEET_MAX_ROWS rows.
var errTooManyRows = stderrors.New("timesheet has too many rows")

// RowProblem is one problem in an imported timesheet: the file line, the column (empty for
// the whole row) and a code such as PERSON_UNKNOWN. TIMESHEET_INVALID lists them in its data.
type RowProblem struct {
	Line   int    `json:"line"`
	Column string `json:"column"`
	Code   string `json:"code"`
}

// timesheetRow is one line of an imported timesheet: the email in lower case, the project id
// as a canonical UUID, and Hours nil when actual_hours is empty.
type timesheetRow struct {
	Line      int
	Email     string
	ProjectID string
	Week      isoweek.Week
	Hours     *float64
}

// timesheetKey is the person, project and week a row is about; a file may name each once.
type timesheetKey struct {
	email, projectID string
	week             isoweek.Week
}

// timesheetLimits are the checks besides the format: no week after Current, at most
// MaxHours in one row, and at most MaxRows rows.
type timesheetLimits struct {
	Current  isoweek.Week
	MaxHours float64
	MaxRows  int
}

// parseTimesheet reads a timesheet CSV and checks its header and the format and limits of
// each row. It returns every row it read and every problem it found; the caller saves nothing
// when there is any. Excel's byte order mark is dropped, header names match in any order and
// case, and the cells a short row lacks count as empty. errTooManyRows means the file holds
// more than limits.MaxRows rows.
func parseTimesheet(data []byte, limits timesheetLimits) ([]timesheetRow, []RowProblem, error) {
	reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(data, []byte(utf8BOM))))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if stderrors.Is(err, io.EOF) {
		return nil, missingColumns(map[string]int{}), nil
	}
	if err != nil {
		return nil, []RowProblem{malformedProblem(err)}, nil
	}
	columns := headerIndex(header)
	if problems := missingColumns(columns); len(problems) > 0 {
		return nil, problems, nil
	}
	var rows []timesheetRow
	var problems []RowProblem
	seen := map[timesheetKey]bool{}
	for {
		record, err := reader.Read()
		if stderrors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			// The reader cannot tell where a broken record ends, so the rest is not read.
			problems = append(problems, malformedProblem(err))
			break
		}
		if len(rows) == limits.MaxRows {
			return nil, nil, errTooManyRows
		}
		line, _ := reader.FieldPos(0)
		row, rowProblems := parseRow(record, columns, line, limits)
		if len(rowProblems) == 0 {
			key := timesheetKey{email: row.Email, projectID: row.ProjectID, week: row.Week}
			if seen[key] {
				rowProblems = append(rowProblems, RowProblem{Line: line, Code: constants.TimesheetRowDuplicate})
			}
			seen[key] = true
		}
		rows = append(rows, row)
		problems = append(problems, rowProblems...)
	}
	return rows, problems, nil
}

// parseRow reads one record by column name and checks its format and limits.
func parseRow(record []string, columns map[string]int, line int, limits timesheetLimits) (timesheetRow, []RowProblem) {
	cell := func(name string) string {
		if i := columns[name]; i < len(record) {
			return strings.TrimSpace(record[i])
		}
		return ""
	}
	var problems []RowProblem
	problem := func(column, code string) {
		problems = append(problems, RowProblem{Line: line, Column: column, Code: code})
	}
	row := timesheetRow{Line: line, Email: strings.ToLower(cell(constants.TimesheetColumnPersonEmail))}
	if row.Email == "" {
		problem(constants.TimesheetColumnPersonEmail, constants.TimesheetPersonUnknown)
	}
	if projectID, err := uuid.Parse(cell(constants.TimesheetColumnProjectID)); err != nil {
		problem(constants.TimesheetColumnProjectID, constants.TimesheetProjectUnknown)
	} else {
		row.ProjectID = projectID.String()
	}
	week, err := isoweek.Parse(cell(constants.TimesheetColumnWeek))
	switch {
	case err != nil:
		problem(constants.TimesheetColumnWeek, constants.TimesheetWeekInvalid)
	case week.After(limits.Current):
		problem(constants.TimesheetColumnWeek, constants.TimesheetWeekFuture)
	default:
		row.Week = week
	}
	if text := cell(constants.TimesheetColumnActualHours); text != "" {
		hours, err := strconv.ParseFloat(text, 64)
		if err != nil || math.IsNaN(hours) || hours < 0 || hours > limits.MaxHours {
			problem(constants.TimesheetColumnActualHours, constants.TimesheetHoursInvalid)
		} else {
			row.Hours = &hours
		}
	}
	return row, problems
}

// headerIndex maps each column name, trimmed and in lower case, to its position; the first of
// two equal names wins.
func headerIndex(header []string) map[string]int {
	out := make(map[string]int, len(header))
	for i, name := range header {
		key := strings.ToLower(strings.TrimSpace(name))
		if _, ok := out[key]; !ok {
			out[key] = i
		}
	}
	return out
}

// missingColumns reports each required column the header lacks.
func missingColumns(columns map[string]int) []RowProblem {
	var problems []RowProblem
	for _, name := range constants.TimesheetRequiredColumns {
		if _, ok := columns[name]; !ok {
			problems = append(problems, RowProblem{Line: headerLine, Column: name, Code: constants.TimesheetColumnMissing})
		}
	}
	return problems
}

// malformedProblem is a CSV syntax error at the line its record starts on.
func malformedProblem(err error) RowProblem {
	problem := RowProblem{Code: constants.TimesheetCSVMalformed}
	var parseErr *csv.ParseError
	if stderrors.As(err, &parseErr) {
		problem.Line = parseErr.StartLine
	}
	return problem
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/services -run TestParseTimesheet -v`, then `gofmt -l . && go build ./... && go vet ./... && make lint`
Expected: PASS; lint `0 issues` (the unexported helpers are used by the tests, and Task 5 uses them in the service).

- [ ] **Step 6: Commit**

```bash
git add internal/constants/timesheet.go internal/services/timesheet.go internal/services/timesheet_test.go
git commit -m "feat(server): read and check a timesheet CSV

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 5: Import actual hours, all or nothing

**Files:**
- Modify: `server/internal/constants/{routes.go,roles.go,error_codes.go,messages.go}`
- Modify: `server/internal/config/config.go`, `server/internal/config/config_test.go`, `server/cmd/server/.env.example`
- Modify: `server/internal/services/authz.go`
- Modify: `server/internal/models/timesheet.go` (`ImportResult`)
- Modify: `server/internal/repositories/person.go`, `server/internal/repositories/project.go`
- Modify: `server/internal/services/person_test.go`, `server/internal/services/project_test.go` (mocks)
- Modify: `server/internal/services/timesheet.go`, `server/internal/services/timesheet_test.go`, `server/internal/services/timesheet_integration_test.go`
- Modify: `server/cmd/server/main.go`

**Interfaces:**
- Consumes: `parseTimesheet`, `RowProblem` (Task 4); `ActualHoursRepository.Upsert` (Task 2).
- Produces:
  - `services.TimesheetService` with `Import(ctx, ImportInput{File io.Reader; Size int64}) (*models.ImportResult, error)` and `services.ProvideTimesheetService(actuals, people, projects, authz, tx, clk, cfg)`
  - `models.ImportResult{Created, Updated, Skipped int}`. It lives in `models`, not `services`, so `dtos` can map it: `services` already imports `dtos`, so `dtos` cannot import `services`.
  - `PersonRepository.ListByEmails(ctx, emails []string) ([]models.Person, error)` and `ProjectRepository.ListByIDs(ctx, ids []string) ([]models.Project, error)`
  - `CanImportTimesheets`, `constants.TimesheetRoles`, `PathTimesheetTemplate`, `PathTimesheetImport`
  - error codes `TimesheetInvalid`, `TimesheetTooLarge`, `TimesheetWeekInFuture`, `TimesheetFileMissing`
  - config fields `TimesheetTemplateWeeks`, `TimesheetMaxWeeks`, `TimesheetMaxWeekHours`, `TimesheetMaxBytes`, `TimesheetMaxRows`
  - test helpers `timesheetConfig()`, `tsNow`, `tsPersonID`, `newTimesheetService(...)`, `importFile(file)`

- [ ] **Step 1: Constants, settings and the permission**

`internal/constants/routes.go`, after the scenario paths:

```go
	PathTimesheetTemplate = "/timesheets/template"
	PathTimesheetImport   = "/timesheets/import"
```

`internal/constants/roles.go`, in the role lists:

```go
	// TimesheetRoles may download the timesheet template and import actual hours.
	TimesheetRoles = []string{RolePlanner}
```

`internal/constants/error_codes.go`, after `AssumptionInvalid`:

```go

	// Timesheet errors
	TimesheetInvalid      = "TIMESHEET_INVALID"
	TimesheetTooLarge     = "TIMESHEET_TOO_LARGE"
	TimesheetWeekInFuture = "TIMESHEET_WEEK_IN_FUTURE"
	TimesheetFileMissing  = "TIMESHEET_FILE_MISSING"
```

`internal/constants/messages.go`, at the end of the block:

```go
	MsgTimesheetImported       = "Timesheet imported"
	MsgTimesheetInvalid        = "Some rows of the timesheet have problems; nothing was saved"
	MsgTimesheetTooLargeFormat = "A timesheet may hold at most %d bytes and %d rows"
	MsgTimesheetUnreadable     = "Could not read the timesheet file"
	MsgTimesheetFileMissing    = "Choose a CSV file to import"
	MsgTimesheetSaveFailed     = "Could not save the actual hours"
	MsgTimesheetWeekInFuture   = "The template cannot cover weeks after the current one"
	MsgTimesheetTemplateFailed = "Could not build the timesheet template"
```

`internal/config/config.go`: add to `Config`, before the program metrics block:

```go
	// Timesheet import (docs/features/timesheet-import.md): the template's default and largest
	// range in weeks, the most actual hours one row may hold, and the largest file.
	TimesheetTemplateWeeks int
	TimesheetMaxWeeks      int
	TimesheetMaxWeekHours  float64
	TimesheetMaxBytes      int64
	TimesheetMaxRows       int
```

and to `Load()`, after `ScenarioMaxAssumptions`:

```go

		TimesheetTemplateWeeks: getEnvAsInt("TIMESHEET_TEMPLATE_WEEKS", 4),
		TimesheetMaxWeeks:      getEnvAsInt("TIMESHEET_MAX_WEEKS", 12),
		TimesheetMaxWeekHours:  getEnvAsFloat("TIMESHEET_MAX_WEEK_HOURS", 80),
		TimesheetMaxBytes:      int64(getEnvAsInt("TIMESHEET_MAX_BYTES", 1<<20)),
		TimesheetMaxRows:       getEnvAsInt("TIMESHEET_MAX_ROWS", 5000),
```

`cmd/server/.env.example`, after the scenario settings:

```
# Timesheet import: weeks in the template by default and at most, the most actual hours one
# row may hold, and the largest file in bytes and rows
TIMESHEET_TEMPLATE_WEEKS=4
TIMESHEET_MAX_WEEKS=12
TIMESHEET_MAX_WEEK_HOURS=80
TIMESHEET_MAX_BYTES=1048576
TIMESHEET_MAX_ROWS=5000
```

Append to `internal/config/config_test.go`:

```go
func TestLoad_TimesheetSettings(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 4, cfg.TimesheetTemplateWeeks)
	assert.Equal(t, 12, cfg.TimesheetMaxWeeks)
	assert.InDelta(t, 80, cfg.TimesheetMaxWeekHours, 1e-9)
	assert.Equal(t, int64(1<<20), cfg.TimesheetMaxBytes)
	assert.Equal(t, 5000, cfg.TimesheetMaxRows)

	t.Setenv("TIMESHEET_MAX_ROWS", "200")
	cfg, err = Load()
	require.NoError(t, err)
	assert.Equal(t, 200, cfg.TimesheetMaxRows)
}
```

`internal/services/authz.go`, after `CanViewEvents`:

```go
// CanImportTimesheets allows planners to download the timesheet template and import actual hours.
func CanImportTimesheets(principal request.Principal) bool {
	return principal.HasAnyRole(constants.TimesheetRoles...)
}
```

Append to `internal/models/timesheet.go`:

```go
// ImportResult counts what a timesheet import did: rows saved as new, rows that replaced a
// value, and rows skipped because actual_hours was empty. It is not a table.
type ImportResult struct {
	Created int
	Updated int
	Skipped int
}
```

Run `make migrate-generate name=check_no_change`. Expected: `The migration directory is synced with the desired state, no changes to be made` and no new file: `ImportResult` has no GORM tags, like `Heatmap`. If a file appears, delete it and report.

- [ ] **Step 2: Repository lookups and their mocks**

`internal/repositories/person.go`: add to `PersonRepository`

```go
	// ListByEmails returns the people with these emails, active or not. Emails are stored in
	// lower case, so callers pass them in lower case.
	ListByEmails(ctx context.Context, emails []string) ([]models.Person, error)
```

and the method:

```go
func (r *personRepository) ListByEmails(ctx context.Context, emails []string) ([]models.Person, error) {
	var people []models.Person
	if len(emails) == 0 {
		return people, nil
	}
	err := r.db.From(ctx).Where("email IN ?", emails).Order("id ASC").Find(&people).Error
	return people, err
}
```

`internal/repositories/project.go`: add to `ProjectRepository`

```go
	// ListByIDs returns the projects with these ids, cancelled ones included.
	ListByIDs(ctx context.Context, ids []string) ([]models.Project, error)
```

and the method:

```go
func (r *projectRepository) ListByIDs(ctx context.Context, ids []string) ([]models.Project, error) {
	var projects []models.Project
	if len(ids) == 0 {
		return projects, nil
	}
	err := r.db.From(ctx).Where("id IN ?", ids).Order("id ASC").Find(&projects).Error
	return projects, err
}
```

Add to `MockPersonRepository` (`internal/services/person_test.go`):

```go
func (m *MockPersonRepository) ListByEmails(ctx context.Context, emails []string) ([]models.Person, error) {
	args := m.Called(ctx, emails)
	return args.Get(0).([]models.Person), args.Error(1)
}
```

Add to `MockProjectRepository` (`internal/services/project_test.go`):

```go
func (m *MockProjectRepository) ListByIDs(ctx context.Context, ids []string) ([]models.Project, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.Project), args.Error(1)
}
```

Run: `go build ./... && go vet ./...`. Expected: no output.

- [ ] **Step 3: Write the failing unit tests** — append to `server/internal/services/timesheet_test.go` (merge the imports: `net/http`, `time`, `the-slot-finder/server/internal/clock`, `the-slot-finder/server/internal/config`, `the-slot-finder/server/internal/errors`):

```go
const tsPersonID = "0192f3a4-0000-7000-8000-00000000e201"

// tsNow is a Wednesday in 2026-W42.
var tsNow = time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC)

func timesheetConfig() *config.Config {
	return &config.Config{TimesheetTemplateWeeks: 4, TimesheetMaxWeeks: 12, TimesheetMaxWeekHours: 80,
		TimesheetMaxBytes: 1 << 10, TimesheetMaxRows: 3}
}

func newTimesheetService(actuals *MockActualHoursRepository, people *MockPersonRepository, projects *MockProjectRepository) TimesheetService {
	return ProvideTimesheetService(actuals, people, projects, ProvideAuthzService(leadPeople{}), passthroughTx{},
		clock.NewFixed(tsNow), timesheetConfig())
}

func importFile(file string) ImportInput {
	return ImportInput{File: strings.NewReader(file), Size: int64(len(file))}
}

func TestTimesheetService_Import(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	ada := models.Person{BaseModel: models.BaseModel{ID: tsPersonID}, Email: "ada@example.test"}
	project := models.Project{BaseModel: models.BaseModel{ID: tsProjectID}}
	known := func() (*MockActualHoursRepository, *MockPersonRepository, *MockProjectRepository) {
		actuals, people, projects := new(MockActualHoursRepository), new(MockPersonRepository), new(MockProjectRepository)
		people.On("ListByEmails", mock.Anything, []string{"ada@example.test"}).Return([]models.Person{ada}, nil)
		projects.On("ListByIDs", mock.Anything, []string{tsProjectID}).Return([]models.Project{project}, nil)
		return actuals, people, projects
	}

	t.Run("saves filled-in rows and skips empty ones", func(t *testing.T) {
		actuals, people, projects := known()
		actuals.On("Upsert", mock.Anything, []models.ActualHours{{PersonID: tsPersonID, ProjectID: tsProjectID,
			Week: mustWeek(t, "2026-W41"), Hours: 18.5, UpdatedBy: "kc-planner"}}).Return(1, 0, nil)

		result, err := newTimesheetService(actuals, people, projects).Import(planner,
			importFile(tsHeader+tsLine("2026-W41", "18.5")+tsLine("2026-W42", "")))
		require.NoError(t, err)
		assert.Equal(t, models.ImportResult{Created: 1, Skipped: 1}, *result)
		actuals.AssertExpectations(t)
	})

	t.Run("an email in another case and with spaces finds the person", func(t *testing.T) {
		actuals, people, projects := known()
		actuals.On("Upsert", mock.Anything, mock.Anything).Return(0, 1, nil)
		file := tsHeader + strings.Replace(tsLine("2026-W41", "8"), "ada@example.test", " Ada@Example.TEST ", 1)

		result, err := newTimesheetService(actuals, people, projects).Import(planner, importFile(file))
		require.NoError(t, err)
		assert.Equal(t, models.ImportResult{Updated: 1}, *result)
		people.AssertExpectations(t)
	})

	t.Run("an unknown person or project saves nothing", func(t *testing.T) {
		actuals, people, projects := new(MockActualHoursRepository), new(MockPersonRepository), new(MockProjectRepository)
		people.On("ListByEmails", mock.Anything, mock.Anything).Return([]models.Person{}, nil)
		projects.On("ListByIDs", mock.Anything, mock.Anything).Return([]models.Project{}, nil)

		_, err := newTimesheetService(actuals, people, projects).Import(planner, importFile(tsHeader+tsLine("2026-W41", "8")))
		requireAppError(t, err, constants.TimesheetInvalid, http.StatusBadRequest)
		assert.Equal(t, []RowProblem{
			{Line: 2, Column: constants.TimesheetColumnPersonEmail, Code: constants.TimesheetPersonUnknown},
			{Line: 2, Column: constants.TimesheetColumnProjectID, Code: constants.TimesheetProjectUnknown},
		}, errors.GetAppError(err).Context[constants.TimesheetProblemsKey])
		actuals.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything)
	})

	t.Run("problems from the file and the lookups come in line order", func(t *testing.T) {
		actuals, people, projects := new(MockActualHoursRepository), new(MockPersonRepository), new(MockProjectRepository)
		people.On("ListByEmails", mock.Anything, mock.Anything).Return([]models.Person{}, nil)
		projects.On("ListByIDs", mock.Anything, mock.Anything).Return([]models.Project{project}, nil)
		file := tsHeader + tsLine("2026-W43", "8") + strings.Replace(tsLine("2026-W41", "8"), "ada@", "bo@", 1)

		_, err := newTimesheetService(actuals, people, projects).Import(planner, importFile(file))
		requireAppError(t, err, constants.TimesheetInvalid, http.StatusBadRequest)
		assert.Equal(t, []RowProblem{
			{Line: 2, Column: constants.TimesheetColumnWeek, Code: constants.TimesheetWeekFuture},
			{Line: 2, Column: constants.TimesheetColumnPersonEmail, Code: constants.TimesheetPersonUnknown},
			{Line: 3, Column: constants.TimesheetColumnPersonEmail, Code: constants.TimesheetPersonUnknown},
		}, errors.GetAppError(err).Context[constants.TimesheetProblemsKey])
	})

	t.Run("a file over the size limit, declared or not", func(t *testing.T) {
		svc := newTimesheetService(nil, nil, nil)
		big := tsHeader + strings.Repeat(" ", 1<<10)
		_, err := svc.Import(planner, importFile(big))
		requireAppError(t, err, constants.TimesheetTooLarge, http.StatusRequestEntityTooLarge)
		_, err = svc.Import(planner, ImportInput{File: strings.NewReader(big)})
		requireAppError(t, err, constants.TimesheetTooLarge, http.StatusRequestEntityTooLarge)
	})

	t.Run("more rows than allowed", func(t *testing.T) {
		file := tsHeader + tsLine("2026-W38", "1") + tsLine("2026-W39", "1") + tsLine("2026-W40", "1") + tsLine("2026-W41", "1")
		_, err := newTimesheetService(nil, nil, nil).Import(planner, importFile(file))
		requireAppError(t, err, constants.TimesheetTooLarge, http.StatusRequestEntityTooLarge)
	})

	t.Run("a database failure", func(t *testing.T) {
		actuals, people, projects := known()
		actuals.On("Upsert", mock.Anything, mock.Anything).Return(0, 0, assert.AnError)
		_, err := newTimesheetService(actuals, people, projects).Import(planner, importFile(tsHeader+tsLine("2026-W41", "8")))
		requireAppError(t, err, constants.DatabaseError, http.StatusInternalServerError)
	})

	t.Run("only planners import", func(t *testing.T) {
		_, err := newTimesheetService(nil, nil, nil).Import(asRole("kc-lead", constants.RoleLead), importFile(tsHeader))
		requireStatus(t, err, http.StatusForbidden)
	})
}
```

Append to `server/internal/services/timesheet_integration_test.go` (merge the imports: `net/http`, `strings`, `the-slot-finder/server/internal/clock`):

```go
func TestTimesheetImport_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	_, project := timesheetFixture(t, pdb)
	planner := asRole("kc-planner", constants.RolePlanner)
	svc := ProvideTimesheetService(repositories.ProvideActualHoursRepository(pdb), repositories.ProvidePersonRepository(pdb),
		repositories.ProvideProjectRepository(pdb), ProvideAuthzService(leadPeople{}), db.ProvideTxManager(pdb),
		clock.NewFixed(tsNow), timesheetConfig())
	good := "person_email,project_id,week,actual_hours\nada@example.test," + project.ID + ",2026-W41,30\n"
	bad := good + "nobody@example.test," + project.ID + ",2026-W42,8\n"
	count := func() int64 {
		var n int64
		require.NoError(t, pdb.Model(&models.ActualHours{}).Count(&n).Error)
		return n
	}

	_, err := svc.Import(planner, importFile(bad))
	requireAppError(t, err, constants.TimesheetInvalid, http.StatusBadRequest)
	assert.Zero(t, count(), "one bad row saves nothing")

	result, err := svc.Import(planner, importFile(good))
	require.NoError(t, err)
	assert.Equal(t, models.ImportResult{Created: 1}, *result)

	result, err = svc.Import(planner, importFile(good))
	require.NoError(t, err)
	assert.Equal(t, models.ImportResult{Updated: 1}, *result, "importing again overwrites")
	assert.Equal(t, int64(1), count())

	var events int64
	require.NoError(t, pdb.Model(&models.ChangeEvent{}).Count(&events).Error)
	assert.Zero(t, events, "an import records no change event")
}
```

- [ ] **Step 4: Run them to verify they fail**

Run: `go test ./internal/services -run 'TestTimesheetService_Import' -v`
Expected: build failure, `undefined: ProvideTimesheetService`.

- [ ] **Step 5: Implement** — add to `server/internal/services/timesheet.go` (merge the imports: `cmp`, `context`, `fmt`, `maps`, `net/http`, `slices`, `the-slot-finder/server/internal/clock`, `the-slot-finder/server/internal/config`, `the-slot-finder/server/internal/db`, `the-slot-finder/server/internal/errors`, `the-slot-finder/server/internal/models`, `the-slot-finder/server/internal/repositories`):

```go
const (
	opImportTimesheet = "import_timesheet"
	resourceTimesheet = "timesheet"
)

// ImportInput is an uploaded timesheet: its content and the size the client declared.
type ImportInput struct {
	File io.Reader
	Size int64
}

// TimesheetService builds the timesheet template and imports actual hours.
type TimesheetService interface {
	// Import checks every row of a filled-in template and saves the actual hours in one
	// transaction. Any problem saves nothing and returns TIMESHEET_INVALID with the problems;
	// a file over TIMESHEET_MAX_BYTES or TIMESHEET_MAX_ROWS returns TIMESHEET_TOO_LARGE.
	Import(ctx context.Context, in ImportInput) (*models.ImportResult, error)
}

type timesheetService struct {
	actuals  repositories.ActualHoursRepository
	people   repositories.PersonRepository
	projects repositories.ProjectRepository
	authz    AuthzService
	tx       db.TxManager
	clock    clock.Clock
	cfg      *config.Config
}

// ProvideTimesheetService creates the timesheet service.
func ProvideTimesheetService(actuals repositories.ActualHoursRepository, people repositories.PersonRepository,
	projects repositories.ProjectRepository, authz AuthzService, tx db.TxManager, clk clock.Clock, cfg *config.Config,
) TimesheetService {
	return &timesheetService{actuals: actuals, people: people, projects: projects, authz: authz, tx: tx, clock: clk, cfg: cfg}
}

func (s *timesheetService) Import(ctx context.Context, in ImportInput) (*models.ImportResult, error) {
	principal, err := s.authz.Require(ctx, CanImportTimesheets)
	if err != nil {
		return nil, err
	}
	data, err := s.readFile(in)
	if err != nil {
		return nil, err
	}
	rows, problems, err := parseTimesheet(data, timesheetLimits{Current: s.currentWeek(),
		MaxHours: s.cfg.TimesheetMaxWeekHours, MaxRows: s.cfg.TimesheetMaxRows})
	if err != nil {
		return nil, s.tooLarge(opImportTimesheet)
	}
	personIDs, projects, err := s.lookups(ctx, rows)
	if err != nil {
		return nil, err
	}
	problems = append(problems, unknownReferences(rows, personIDs, projects)...)
	if len(problems) > 0 {
		slices.SortStableFunc(problems, func(a, b RowProblem) int { return cmp.Compare(a.Line, b.Line) })
		return nil, errors.NewAppError(constants.TimesheetInvalid, constants.MsgTimesheetInvalid, errors.ErrorTypeValidation,
			http.StatusBadRequest).WithContext(constants.TimesheetProblemsKey, problems).
			WithOperation(opImportTimesheet).WithResource(resourceTimesheet)
	}
	saved := make([]models.ActualHours, 0, len(rows))
	for _, row := range rows {
		if row.Hours != nil {
			saved = append(saved, models.ActualHours{PersonID: personIDs[row.Email], ProjectID: row.ProjectID, Week: row.Week,
				Hours: *row.Hours, UpdatedBy: principal.Subject})
		}
	}
	result := &models.ImportResult{Skipped: len(rows) - len(saved)}
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		created, updated, err := s.actuals.Upsert(ctx, saved)
		result.Created, result.Updated = created, updated
		return err
	})
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgTimesheetSaveFailed, err).WithOperation(opImportTimesheet).WithResource(resourceTimesheet)
	}
	return result, nil
}

// readFile reads the upload, refusing more than TIMESHEET_MAX_BYTES whatever size the client
// declared.
func (s *timesheetService) readFile(in ImportInput) ([]byte, error) {
	if in.Size > s.cfg.TimesheetMaxBytes {
		return nil, s.tooLarge(opImportTimesheet)
	}
	data, err := io.ReadAll(io.LimitReader(in.File, s.cfg.TimesheetMaxBytes+1))
	if err != nil {
		return nil, errors.ValidationError(constants.MsgTimesheetUnreadable, err).WithOperation(opImportTimesheet).WithResource(resourceTimesheet)
	}
	if int64(len(data)) > s.cfg.TimesheetMaxBytes {
		return nil, s.tooLarge(opImportTimesheet)
	}
	return data, nil
}

// tooLarge is the 413 for a file over TIMESHEET_MAX_BYTES or TIMESHEET_MAX_ROWS.
func (s *timesheetService) tooLarge(op string) error {
	return errors.NewAppError(constants.TimesheetTooLarge,
		fmt.Sprintf(constants.MsgTimesheetTooLargeFormat, s.cfg.TimesheetMaxBytes, s.cfg.TimesheetMaxRows),
		errors.ErrorTypeValidation, http.StatusRequestEntityTooLarge).WithOperation(op).WithResource(resourceTimesheet)
}

// currentWeek is the organization's current ISO week.
func (s *timesheetService) currentWeek() isoweek.Week {
	return isoweek.Of(s.clock.Now().In(s.clock.Location()))
}

// lookups loads the people the rows name by email and the projects they name by id. It returns
// person ids by lower-case email, and the ids of the projects that exist.
func (s *timesheetService) lookups(ctx context.Context, rows []timesheetRow) (map[string]string, map[string]bool, error) {
	emails, projectIDs := map[string]bool{}, map[string]bool{}
	for _, row := range rows {
		if row.Email != "" {
			emails[row.Email] = true
		}
		if row.ProjectID != "" {
			projectIDs[row.ProjectID] = true
		}
	}
	people, err := s.people.ListByEmails(ctx, slices.Sorted(maps.Keys(emails)))
	if err != nil {
		return nil, nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opImportTimesheet).WithResource(resourceTimesheet)
	}
	projects, err := s.projects.ListByIDs(ctx, slices.Sorted(maps.Keys(projectIDs)))
	if err != nil {
		return nil, nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opImportTimesheet).WithResource(resourceTimesheet)
	}
	personIDs := make(map[string]string, len(people))
	for _, p := range people {
		personIDs[strings.ToLower(p.Email)] = p.ID
	}
	known := make(map[string]bool, len(projects))
	for _, p := range projects {
		known[p.ID] = true
	}
	return personIDs, known, nil
}

// unknownReferences reports rows whose email or project id matches nothing. Rows whose cell
// was already empty or malformed are reported by parseTimesheet instead.
func unknownReferences(rows []timesheetRow, personIDs map[string]string, projects map[string]bool) []RowProblem {
	var problems []RowProblem
	for _, row := range rows {
		if _, ok := personIDs[row.Email]; row.Email != "" && !ok {
			problems = append(problems, RowProblem{Line: row.Line, Column: constants.TimesheetColumnPersonEmail,
				Code: constants.TimesheetPersonUnknown})
		}
		if row.ProjectID != "" && !projects[row.ProjectID] {
			problems = append(problems, RowProblem{Line: row.Line, Column: constants.TimesheetColumnProjectID,
				Code: constants.TimesheetProjectUnknown})
		}
	}
	return problems
}
```

In `cmd/server/main.go`, add `services.ProvideTimesheetService,` to `fx.Provide` after `services.ProvideScenarioService,`.

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./internal/services ./internal/config -run 'TestTimesheetService_Import|TestParseTimesheet|TestLoad_TimesheetSettings' -v`, then `go test -tags=integration ./internal/services -run 'Timesheet|ActualHours' -v`, then `gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`
Expected: PASS everywhere; lint `0 issues`.

- [ ] **Step 7: Commit**

```bash
git add internal/constants internal/config cmd/server/.env.example cmd/server/main.go internal/models/timesheet.go \
  internal/repositories/person.go internal/repositories/project.go internal/services
git commit -m "feat(server): import actual hours from a timesheet, all or nothing

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 6: Build the CSV template

**Files:**
- Modify: `server/internal/services/timesheet.go`
- Modify: `server/internal/services/timesheet_test.go`
- Modify: `server/internal/services/timesheet_integration_test.go` (constructor call)

**Interfaces:**
- Consumes: `findBaseline(ctx, plans, op, true)` (`services/plan.go`); `ActualHoursRepository.ListInRange` (Task 2); `constants.TimesheetColumns`, `TimesheetFormulaPrefixes`, `TimesheetTextPrefix` (Task 4).
- Produces:
  - `TimesheetService.Template(ctx, TemplateInput{From, To isoweek.Week}) ([]byte, error)`. A zero week takes the default: `To` = the current week, `From` = `To` minus `TIMESHEET_TEMPLATE_WEEKS − 1` weeks.
  - `ProvideTimesheetService(actuals, people, projects, plans repositories.PlanRepository, authz, tx, clk, cfg)`: `plans` is new, after `projects`
  - `writeTemplate([]templateRow) ([]byte, error)`, `safeCell(string) string`

- [ ] **Step 1: Write the failing tests**

In `server/internal/services/timesheet_test.go`, change the body of `newTimesheetService` to pass `nil` for the new `plans` argument: `ProvideTimesheetService(actuals, people, projects, nil, ProvideAuthzService(leadPeople{}), passthroughTx{}, clock.NewFixed(tsNow), timesheetConfig())`. In `timesheet_integration_test.go`, add `repositories.ProvidePlanRepository(pdb),` after `repositories.ProvideProjectRepository(pdb),` in the `ProvideTimesheetService` call. Then append to `timesheet_test.go`:

```go
const (
	tsProgramID      = "0192f3a4-0000-7000-8000-00000000e301"
	tsOtherProjectID = "0192f3a4-0000-7000-8000-00000000e102"
)

func newTemplateService(plans *MockPlanRepository, actuals *MockActualHoursRepository) TimesheetService {
	return ProvideTimesheetService(actuals, nil, nil, plans, ProvideAuthzService(leadPeople{}), passthroughTx{},
		clock.NewFixed(tsNow), timesheetConfig())
}

func TestTimesheetService_Template(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	program := &models.Program{BaseModel: models.BaseModel{ID: tsProgramID}, Name: "Acme"}
	checkout := &models.Project{BaseModel: models.BaseModel{ID: tsProjectID}, Name: "Checkout", Program: program}
	loyalty := &models.Project{BaseModel: models.BaseModel{ID: tsOtherProjectID}, Name: "Loyalty", Program: program}
	ada := &models.Person{BaseModel: models.BaseModel{ID: tsPersonID}, Name: "Ada", Email: "ada@example.test"}
	baseline := publishedBaseline(1)
	baseline.Assignments = []models.Assignment{{PersonID: tsPersonID, Person: ada,
		Demand: &models.Demand{ProjectID: tsProjectID, Project: checkout},
		Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W38"), Hours: 40}, {Week: mustWeek(t, "2026-W41"), Hours: 20},
			{Week: mustWeek(t, "2026-W42"), Hours: 16}}}}
	recorded := []models.ActualHours{
		{PersonID: tsPersonID, ProjectID: tsProjectID, Week: mustWeek(t, "2026-W41"), Hours: 18, Person: ada, Project: checkout},
		{PersonID: tsPersonID, ProjectID: tsOtherProjectID, Week: mustWeek(t, "2026-W42"), Hours: 4.5, Person: ada, Project: loyalty},
	}
	fixture := func(from, to string) (*MockPlanRepository, *MockActualHoursRepository) {
		plans, actuals := new(MockPlanRepository), new(MockActualHoursRepository)
		plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
		actuals.On("ListInRange", mock.Anything, mustWeek(t, from), mustWeek(t, to)).Return(recorded, nil)
		return plans, actuals
	}

	t.Run("the last weeks up to this one, from the baseline and the recorded hours", func(t *testing.T) {
		plans, actuals := fixture("2026-W39", "2026-W42")
		data, err := newTemplateService(plans, actuals).Template(planner, TemplateInput{})
		require.NoError(t, err)
		assert.Equal(t, tsHeader+
			"ada@example.test,Ada,Acme,Checkout,"+tsProjectID+",2026-W41,20,18\n"+
			"ada@example.test,Ada,Acme,Checkout,"+tsProjectID+",2026-W42,16,\n"+
			"ada@example.test,Ada,Acme,Loyalty,"+tsOtherProjectID+",2026-W42,0,4.5\n", string(data),
			"W38 is before the range; Loyalty has actual hours without baseline hours")
		actuals.AssertExpectations(t)
	})

	t.Run("a chosen range", func(t *testing.T) {
		plans, actuals := fixture("2026-W41", "2026-W41")
		data, err := newTemplateService(plans, actuals).Template(planner,
			TemplateInput{From: mustWeek(t, "2026-W41"), To: mustWeek(t, "2026-W41")})
		require.NoError(t, err)
		assert.Equal(t, tsHeader+"ada@example.test,Ada,Acme,Checkout,"+tsProjectID+",2026-W41,20,18\n", string(data))
	})

	ranges := []struct {
		name string
		in   TemplateInput
		code string
	}{
		{name: "a week after the current one", in: TemplateInput{To: mustWeek(t, "2026-W43")}, code: constants.TimesheetWeekInFuture},
		{name: "more weeks than allowed", in: TemplateInput{From: mustWeek(t, "2026-W20"), To: mustWeek(t, "2026-W42")},
			code: constants.WeekRangeTooLong},
		{name: "to before from", in: TemplateInput{From: mustWeek(t, "2026-W42"), To: mustWeek(t, "2026-W41")},
			code: constants.WeekRangeInvalid},
	}
	for _, tt := range ranges {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTemplateService(nil, nil).Template(planner, tt.in)
			requireAppError(t, err, tt.code, http.StatusBadRequest)
		})
	}

	t.Run("only planners", func(t *testing.T) {
		_, err := newTemplateService(nil, nil).Template(asRole("kc-lead", constants.RoleLead), TemplateInput{})
		requireStatus(t, err, http.StatusForbidden)
	})
}

func TestWriteTemplate_FormulaCells(t *testing.T) {
	data, err := writeTemplate([]templateRow{{Email: "ada@example.test", Name: "+Ada", Program: "@Acme",
		Project: `=HYPERLINK("http://example.test")`, ProjectID: tsProjectID, Week: mustWeek(t, "2026-W41"), Planned: 20}})
	require.NoError(t, err)
	assert.Equal(t, tsHeader+`ada@example.test,'+Ada,'@Acme,"'=HYPERLINK(""http://example.test"")",`+tsProjectID+
		",2026-W41,20,\n", string(data))
}

func TestSafeCell(t *testing.T) {
	tests := map[string]string{"=SUM(A1)": "'=SUM(A1)", "+1": "'+1", "-2": "'-2", "@x": "'@x", "\tx": "'\tx", "Acme": "Acme", "": ""}
	for in, want := range tests {
		assert.Equal(t, want, safeCell(in), in)
	}
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./internal/services -run 'TestTimesheetService_Template|TestWriteTemplate|TestSafeCell' -v`
Expected: build failure, `too many arguments in call to ProvideTimesheetService` / `undefined: TemplateInput`.

- [ ] **Step 3: Implement** — in `server/internal/services/timesheet.go`:

Add `plans repositories.PlanRepository` to `timesheetService` after `projects`; add the parameter `plans repositories.PlanRepository` to `ProvideTimesheetService` after `projects` and `plans: plans,` to its literal. Add `opTimesheetTemplate = "timesheet_template"` to the `const` block with `opImportTimesheet`. Add to the `TimesheetService` interface:

```go
	// Template is a CSV with one row per person × project × week that has baseline hours or
	// actual hours in the range, with the planned and any recorded actual hours.
	Template(ctx context.Context, in TemplateInput) ([]byte, error)
```

and add:

```go
// hourPrecision rounds hours in the template to hundredths, like the numeric(6,2) columns.
const hourPrecision = 100

// TemplateInput is the range of weeks of a timesheet template. A zero week takes its
// default: To is the current week, and From the TIMESHEET_TEMPLATE_WEEKS up to To.
type TemplateInput struct {
	From isoweek.Week
	To   isoweek.Week
}

func (s *timesheetService) Template(ctx context.Context, in TemplateInput) ([]byte, error) {
	if _, err := s.authz.Require(ctx, CanImportTimesheets); err != nil {
		return nil, err
	}
	from, to, err := s.templateRange(in)
	if err != nil {
		return nil, err
	}
	baseline, err := findBaseline(ctx, s.plans, opTimesheetTemplate, true)
	if err != nil {
		return nil, err
	}
	recorded, err := s.actuals.ListInRange(ctx, from, to)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opTimesheetTemplate).WithResource(resourceTimesheet)
	}
	data, err := writeTemplate(templateRows(baseline.Assignments, recorded, from, to))
	if err != nil {
		return nil, errors.InternalError(constants.MsgTimesheetTemplateFailed, err).WithOperation(opTimesheetTemplate).WithResource(resourceTimesheet)
	}
	return data, nil
}

// templateRange applies the defaults and checks that the range is in order, not after the
// current week, and not longer than TIMESHEET_MAX_WEEKS.
func (s *timesheetService) templateRange(in TemplateInput) (from, to isoweek.Week, err error) {
	current := s.currentWeek()
	to = in.To
	if to.IsZero() {
		to = current
	}
	from = in.From
	if from.IsZero() {
		from = to.AddWeeks(1 - s.cfg.TimesheetTemplateWeeks)
	}
	switch {
	case to.Before(from):
		return from, to, errors.NewAppError(constants.WeekRangeInvalid, constants.MsgWeekRangeInvalid, errors.ErrorTypeValidation,
			http.StatusBadRequest).WithOperation(opTimesheetTemplate).WithResource(resourceTimesheet)
	case to.After(current):
		return from, to, errors.NewAppError(constants.TimesheetWeekInFuture, constants.MsgTimesheetWeekInFuture,
			errors.ErrorTypeValidation, http.StatusBadRequest).WithOperation(opTimesheetTemplate).WithResource(resourceTimesheet)
	case isoweek.WeeksBetween(from, to)+1 > s.cfg.TimesheetMaxWeeks:
		return from, to, errors.NewAppError(constants.WeekRangeTooLong, fmt.Sprintf(constants.MsgWeekRangeTooLongFormat,
			s.cfg.TimesheetMaxWeeks), errors.ErrorTypeValidation, http.StatusBadRequest).
			WithOperation(opTimesheetTemplate).WithResource(resourceTimesheet)
	}
	return from, to, nil
}

// templateRow is one line of the timesheet template; Actual is nil when nothing is recorded.
type templateRow struct {
	Email, Name, Program, Project, ProjectID string
	Week                                     isoweek.Week
	Planned                                  float64
	Actual                                   *float64
}

// templateRows sums the baseline hours per person, project and week in the range, and adds
// the recorded actual hours, also where there are no baseline hours. Rows are sorted by
// person, program, project and week.
func templateRows(assignments []models.Assignment, recorded []models.ActualHours, from, to isoweek.Week) []templateRow {
	type key struct {
		personID, projectID string
		week                isoweek.Week
	}
	inRange := func(week isoweek.Week) bool { return !week.Before(from) && !week.After(to) }
	rows := map[key]*templateRow{}
	row := func(person models.Person, project models.Project, week isoweek.Week) *templateRow {
		k := key{person.ID, project.ID, week}
		if r, ok := rows[k]; ok {
			return r
		}
		r := &templateRow{Email: person.Email, Name: person.Name, Project: project.Name, ProjectID: project.ID, Week: week}
		if project.Program != nil {
			r.Program = project.Program.Name
		}
		rows[k] = r
		return r
	}
	for _, a := range assignments {
		if a.Person == nil || a.Demand == nil || a.Demand.Project == nil {
			continue
		}
		for _, alloc := range a.Allocations {
			if inRange(alloc.Week) && alloc.Hours > hoursEpsilon {
				row(*a.Person, *a.Demand.Project, alloc.Week).Planned += alloc.Hours
			}
		}
	}
	for _, r := range recorded {
		if r.Person == nil || r.Project == nil || !inRange(r.Week) {
			continue
		}
		hours := r.Hours
		row(*r.Person, *r.Project, r.Week).Actual = &hours
	}
	out := make([]templateRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, *r)
	}
	slices.SortFunc(out, func(a, b templateRow) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Email, b.Email), cmp.Compare(a.Program, b.Program),
			cmp.Compare(a.Project, b.Project), cmp.Compare(a.ProjectID, b.ProjectID), a.Week.Compare(b.Week))
	})
	return out
}

// writeTemplate writes the header and the rows as CSV.
func writeTemplate(rows []templateRow) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(constants.TimesheetColumns); err != nil {
		return nil, err
	}
	for _, r := range rows {
		actual := ""
		if r.Actual != nil {
			actual = formatHours(*r.Actual)
		}
		record := []string{r.Email, safeCell(r.Name), safeCell(r.Program), safeCell(r.Project), r.ProjectID, r.Week.String(),
			formatHours(r.Planned), actual}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

// formatHours writes hours rounded to hundredths, without trailing zeros: 20, 18.5.
func formatHours(hours float64) string {
	return strconv.FormatFloat(math.Round(hours*hourPrecision)/hourPrecision, 'f', -1, 64)
}

// safeCell puts TimesheetTextPrefix before text a spreadsheet would run as a formula, so it
// shows as text. Names are free text typed by users; emails, ids and numbers never need it.
func safeCell(text string) string {
	if text != "" && strings.ContainsRune(constants.TimesheetFormulaPrefixes, rune(text[0])) {
		return constants.TimesheetTextPrefix + text
	}
	return text
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/services -run 'Timesheet|WriteTemplate|SafeCell|ParseTimesheet' -v`, then `go test -tags=integration ./internal/services -run 'Timesheet|ActualHours' -v`, then `gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`
Expected: PASS; lint `0 issues`.

- [ ] **Step 5: Commit**

```bash
git add internal/services/timesheet.go internal/services/timesheet_test.go internal/services/timesheet_integration_test.go
git commit -m "feat(server): timesheet template from the baseline and recorded hours

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 7: Timesheet routes, handlers and Swagger

**Files:**
- Create: `server/internal/dtos/timesheet.go`
- Create: `server/internal/handlers/timesheet.go`
- Create: `server/internal/handlers/timesheet_test.go`
- Modify: `server/cmd/server/routes/router.go`, `server/cmd/server/routes/router_test.go`
- Modify: `server/cmd/server/main.go`
- Generate: `server/docs/*`

**Interfaces:**
- Consumes: `services.TimesheetService` (Tasks 5–6), `models.ImportResult`, `services.RowProblem`.
- Produces: `handlers.TimesheetHandler` (`Template`, `Import`) and `handlers.ProvideTimesheetHandler(services.TimesheetService, *validator.Validate)`; `routes.Router(…, scenarioHandler, scheduleHandler, timesheetHandler *handlers.TimesheetHandler, verifier, cfg)` (`scheduleHandler` comes from member-calendar); `dtos.TimesheetTemplateRequest`, `dtos.ImportResultResponse`, `dtos.NewImportResultResponse`.

- [ ] **Step 1: Write the DTOs** — create `server/internal/dtos/timesheet.go`:

```go
package dtos

import (
	"strings"

	"the-slot-finder/server/internal/models"
)

// TimesheetTemplateRequest is the query of GET /timesheets/template. Both weeks are optional:
// the range defaults to the TIMESHEET_TEMPLATE_WEEKS up to the current week.
type TimesheetTemplateRequest struct {
	From string `json:"from" query:"from" validate:"omitempty,isoweek"`
	To   string `json:"to" query:"to" validate:"omitempty,isoweek"`
}

// Normalize trims the weeks.
func (r *TimesheetTemplateRequest) Normalize() {
	r.From = strings.TrimSpace(r.From)
	r.To = strings.TrimSpace(r.To)
}

// ImportResultResponse counts what a timesheet import did.
type ImportResultResponse struct {
	Created int `json:"created" example:"12"`
	Updated int `json:"updated" example:"3"`
	Skipped int `json:"skipped" example:"5"`
}

// NewImportResultResponse maps an import result.
func NewImportResultResponse(r models.ImportResult) ImportResultResponse {
	return ImportResultResponse{Created: r.Created, Updated: r.Updated, Skipped: r.Skipped}
}
```

- [ ] **Step 2: Write the failing handler tests** — create `server/internal/handlers/timesheet_test.go`:

```go
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/services"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const timesheetCSV = "person_email,project_id,week,actual_hours\nada@example.test,0192f3a4-0000-7000-8000-00000000e101,2026-W41,8\n"

// fakeTimesheets implements services.TimesheetService for handler tests.
type fakeTimesheets struct {
	template   []byte
	templateIn services.TemplateInput
	uploaded   string
	result     models.ImportResult
	err        error
}

func (f *fakeTimesheets) Template(_ context.Context, in services.TemplateInput) ([]byte, error) {
	f.templateIn = in
	return f.template, f.err
}

func (f *fakeTimesheets) Import(_ context.Context, in services.ImportInput) (*models.ImportResult, error) {
	data, err := io.ReadAll(in.File)
	if err != nil {
		return nil, err
	}
	f.uploaded = string(data)
	if f.err != nil {
		return nil, f.err
	}
	return &f.result, nil
}

// serveTimesheet runs one timesheet handler with the planner signed in.
func serveTimesheet(t *testing.T, method, target string, body io.Reader, contentType string, handler echo.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	path, _, _ := strings.Cut(target, "?")
	e.Add(method, path, handler)
	req := httptest.NewRequest(method, target, body)
	if contentType != "" {
		req.Header.Set(echo.HeaderContentType, contentType)
	}
	req = req.WithContext(plannerContext())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// multipartBody is a form holding one file.
func multipartBody(t *testing.T, field, name, content string) (io.Reader, string) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile(field, name)
	require.NoError(t, err)
	_, err = part.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return &body, w.FormDataContentType()
}

func TestTimesheetHandler_Template(t *testing.T) {
	t.Run("downloads the CSV for the chosen weeks", func(t *testing.T) {
		fake := &fakeTimesheets{template: []byte(timesheetCSV)}
		h := ProvideTimesheetHandler(fake, newTestValidator(t))
		rec := serveTimesheet(t, http.MethodGet, constants.PathTimesheetTemplate+"?from=2026-W41&to=2026-W42", nil, "", h.Template)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, constants.TimesheetContentType, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, `attachment; filename="timesheet.csv"`, rec.Header().Get(echo.HeaderContentDisposition))
		assert.Equal(t, timesheetCSV, rec.Body.String())
		assert.Equal(t, "2026-W41", fake.templateIn.From.String())
		assert.Equal(t, "2026-W42", fake.templateIn.To.String())
	})

	t.Run("without weeks the service picks the defaults", func(t *testing.T) {
		fake := &fakeTimesheets{template: []byte(timesheetCSV)}
		rec := serveTimesheet(t, http.MethodGet, constants.PathTimesheetTemplate, nil, "",
			ProvideTimesheetHandler(fake, newTestValidator(t)).Template)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.True(t, fake.templateIn.From.IsZero())
		assert.True(t, fake.templateIn.To.IsZero())
	})

	t.Run("a week that does not exist", func(t *testing.T) {
		rec := serveTimesheet(t, http.MethodGet, constants.PathTimesheetTemplate+"?from=2026-W99", nil, "",
			ProvideTimesheetHandler(&fakeTimesheets{}, newTestValidator(t)).Template)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestTimesheetHandler_Import(t *testing.T) {
	t.Run("imports the uploaded file", func(t *testing.T) {
		fake := &fakeTimesheets{result: models.ImportResult{Created: 1, Updated: 2, Skipped: 3}}
		body, contentType := multipartBody(t, constants.TimesheetFormField, constants.TimesheetFileName, timesheetCSV)
		rec := serveTimesheet(t, http.MethodPost, constants.PathTimesheetImport, body, contentType,
			ProvideTimesheetHandler(fake, newTestValidator(t)).Import)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, timesheetCSV, fake.uploaded)
		var envelope dtos.BaseResponse[dtos.ImportResultResponse]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
		assert.Equal(t, dtos.ImportResultResponse{Created: 1, Updated: 2, Skipped: 3}, envelope.Data)
	})

	t.Run("no file", func(t *testing.T) {
		rec := serveTimesheet(t, http.MethodPost, constants.PathTimesheetImport, bytes.NewBufferString("{}"), echo.MIMEApplicationJSON,
			ProvideTimesheetHandler(&fakeTimesheets{}, newTestValidator(t)).Import)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		var envelope dtos.BaseResponse[json.RawMessage]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
		assert.Equal(t, constants.ValidationError, envelope.Meta.ErrorCode)
	})

	t.Run("row problems come back in the data", func(t *testing.T) {
		problems := []services.RowProblem{{Line: 5, Column: constants.TimesheetColumnPersonEmail, Code: constants.TimesheetPersonUnknown}}
		fake := &fakeTimesheets{err: errors.NewAppError(constants.TimesheetInvalid, constants.MsgTimesheetInvalid,
			errors.ErrorTypeValidation, http.StatusBadRequest).WithContext(constants.TimesheetProblemsKey, problems)}
		body, contentType := multipartBody(t, constants.TimesheetFormField, constants.TimesheetFileName, timesheetCSV)
		rec := serveTimesheet(t, http.MethodPost, constants.PathTimesheetImport, body, contentType,
			ProvideTimesheetHandler(fake, newTestValidator(t)).Import)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		var envelope dtos.BaseResponse[map[string][]services.RowProblem]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
		assert.Equal(t, constants.TimesheetInvalid, envelope.Meta.ErrorCode)
		assert.Equal(t, problems, envelope.Data[constants.TimesheetProblemsKey])
	})
}
```

Add RBAC rows to the table in `TestRouter_People_RBAC` (`cmd/server/routes/router_test.go`):

```go
		{name: "lead cannot download the timesheet template", method: http.MethodGet, path: constants.PathTimesheetTemplate, token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot download the timesheet template", method: http.MethodGet, path: constants.PathTimesheetTemplate, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "lead cannot import a timesheet", method: http.MethodPost, path: constants.PathTimesheetImport, token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot import a timesheet", method: http.MethodPost, path: constants.PathTimesheetImport, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no token cannot import a timesheet", method: http.MethodPost, path: constants.PathTimesheetImport, wantStatus: http.StatusUnauthorized},
```

and in the helper that calls `Router` (member-calendar renamed it `newTestRouterWith`) add `timesheets := handlers.ProvideTimesheetHandler(nil, nil)` and pass `timesheets` to `Router` after `schedule`:

```go
	timesheets := handlers.ProvideTimesheetHandler(nil, nil)
	return Router(health, identity, person, absence, program, project, plan, staffing, insights, events, scenarios, schedule,
		timesheets, verifier, cfg)
```

- [ ] **Step 3: Run them to verify they fail**

Run: `go test ./internal/handlers ./cmd/server/routes -run 'Timesheet|RBAC' -v`
Expected: build failure, `undefined: ProvideTimesheetHandler`.

- [ ] **Step 4: Implement the handler** — create `server/internal/handlers/timesheet.go`:

```go
package handlers

import (
	"fmt"
	"net/http"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// TimesheetHandler serves the timesheet template download and the import of actual hours.
type TimesheetHandler struct {
	BaseHandler
	timesheets services.TimesheetService
	validate   *validator.Validate
}

// ProvideTimesheetHandler creates the timesheet handler.
func ProvideTimesheetHandler(timesheets services.TimesheetService, validate *validator.Validate) *TimesheetHandler {
	return &TimesheetHandler{BaseHandler: *NewBaseHandler(), timesheets: timesheets, validate: validate}
}

// Template godoc
// @Summary Timesheet template
// @Description A CSV with one row per person × project × week that has baseline or actual hours in the range, with planned hours and any actual hours already recorded. Weeks default to the TIMESHEET_TEMPLATE_WEEKS up to the current week; a week after the current one gets TIMESHEET_WEEK_IN_FUTURE and a range over TIMESHEET_MAX_WEEKS gets WEEK_RANGE_TOO_LONG.
// @Tags Timesheets
// @Produce text/csv
// @Security BearerAuth
// @Param from query string false "First week" example(2026-W39)
// @Param to query string false "Last week" example(2026-W42)
// @Success 200 {file} file
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /timesheets/template [get]
func (h *TimesheetHandler) Template(c *echo.Context) error {
	var req dtos.TimesheetTemplateRequest
	if err := echo.BindQueryParams(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	var in services.TemplateInput
	var err error
	if in.From, err = weekOrZero(req.From); err != nil {
		return h.HandleError(c, err)
	}
	if in.To, err = weekOrZero(req.To); err != nil {
		return h.HandleError(c, err)
	}
	data, err := h.timesheets.Template(c.Request().Context(), in)
	if err != nil {
		return h.HandleError(c, err)
	}
	c.Response().Header().Set(echo.HeaderContentDisposition,
		fmt.Sprintf(constants.ContentDispositionAttachmentFormat, constants.TimesheetFileName))
	return c.Blob(http.StatusOK, constants.TimesheetContentType, data)
}

// Import godoc
// @Summary Import a timesheet
// @Description Upload a filled-in template as the multipart field "file". Every row is checked first: any problem saves nothing and returns 400 TIMESHEET_INVALID with data.problems [{line, column, code}]. A row with an empty actual_hours is skipped; importing again overwrites the same person, project and week. A file over TIMESHEET_MAX_BYTES or TIMESHEET_MAX_ROWS gets 413 TIMESHEET_TOO_LARGE. Records no change event.
// @Tags Timesheets
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Filled-in timesheet template (CSV)"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ImportResultResponse}
// @Failure 400 {object} object{meta=dtos.Meta,data=object{problems=[]services.RowProblem}}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 413 {object} object{meta=dtos.Meta}
// @Router /timesheets/import [post]
func (h *TimesheetHandler) Import(c *echo.Context) error {
	header, err := c.FormFile(constants.TimesheetFormField)
	if err != nil {
		return h.ValidationErrorResponse(c, constants.MsgTimesheetFileMissing,
			map[string]string{constants.TimesheetFormField: constants.MsgTimesheetFileMissing})
	}
	file, err := header.Open()
	if err != nil {
		return h.HandleError(c, errors.ValidationError(constants.MsgTimesheetUnreadable, err))
	}
	defer func() { _ = file.Close() }()
	result, err := h.timesheets.Import(c.Request().Context(), services.ImportInput{File: file, Size: header.Size})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgTimesheetImported, dtos.NewImportResultResponse(*result), nil)
}

// weekOrZero parses a week the isoweek tag has checked; empty is the zero week.
func weekOrZero(value string) (isoweek.Week, error) {
	if value == "" {
		return isoweek.Week{}, nil
	}
	return dtos.ParseWeek(value)
}
```

Wire it: in `cmd/server/routes/router.go` add the parameter `timesheetHandler *handlers.TimesheetHandler,` after `scheduleHandler *handlers.ScheduleHandler,` and, after the scenario routes:

```go

	// Timesheets: planners download the template and import actual hours.
	timesheetRoles := middlewares.RequireRole(constants.TimesheetRoles...)
	v1.GET(constants.PathTimesheetTemplate, timesheetHandler.Template, authenticated, timesheetRoles)
	v1.POST(constants.PathTimesheetImport, timesheetHandler.Import, authenticated, timesheetRoles)
```

In `cmd/server/main.go`:
- add `timesheetHandler *handlers.TimesheetHandler,` to `NewHTTPServer` after `scheduleHandler`
- pass it to `routes.Router(…, scenarioHandler, scheduleHandler, timesheetHandler, verifier, cfg)`
- add `handlers.ProvideTimesheetHandler,` to `fx.Provide` after `handlers.ProvideScheduleHandler,`

- [ ] **Step 5: Run the tests, regenerate Swagger, check the side**

Run: `go test ./internal/handlers ./cmd/server/... -v -run 'Timesheet|RBAC|AppOptions'`, then `make swagger-load` (or the `go run .../swag@v1.16.6 init ...` command from Task 1), then `gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`
Expected: PASS; `docs/swagger.yaml` gains `/timesheets/template` and `/timesheets/import`; lint `0 issues`.

- [ ] **Step 6: Try it on the running stack**

Run: `make up` at the repo root (or `make server`), sign in as `planner` in the client, copy the access token from the browser's devtools (Application → Session storage), then:
`curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:3000/api/v1/timesheets/template" | head -3`
Expected: the CSV header and baseline rows of the seed (Checkout Revamp, Data Migration).

- [ ] **Step 7: Commit**

```bash
git add internal/dtos/timesheet.go internal/handlers/timesheet.go internal/handlers/timesheet_test.go cmd/server docs
git commit -m "feat(server): timesheet template and import routes for planners

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 8: The member schedule shows actual hours per week

**Files:**
- Modify: `server/internal/models/schedule.go` (from member-calendar: `models.ScheduleWeek`)
- Modify: `server/internal/services/schedule.go` (from member-calendar: `scheduleService`, `ProvideScheduleService`, `Get`)
- Modify: `server/internal/services/schedule_test.go` (from member-calendar: `scheduleRepos`, `scheduleFixture`, `newScheduleService`, `TestScheduleService_Get`)
- Modify: `server/internal/dtos/schedule.go` (from member-calendar: `ScheduleWeekResponse`, `NewScheduleResponse`)
- Generate: `server/docs/*`

**Interfaces:**
- Consumes: from member-calendar, `models.ScheduleWeek` (aliased in services as `type ScheduleWeek = models.ScheduleWeek`) and `ScheduleService.Get(ctx, ScheduleInput{PersonID string; From, To isoweek.Week}) (*PersonSchedule, error)`; `ActualHoursRepository.ByPersonWeeks` (Task 2).
- Produces: `models.ScheduleWeek.ActualHours *float64` (JSON `actual_hours`, `null` when none); `withActualHours(weeks []ScheduleWeek, worked map[isoweek.Week]float64)`; `ProvideScheduleService(plans, people, absences, actuals repositories.ActualHoursRepository, authz, cfg)`, where `actuals` is new, after `absences`.

- [ ] **Step 1: Write the failing tests**

In `server/internal/services/schedule_test.go`:
- add `actuals *MockActualHoursRepository` to `scheduleRepos`
- in `scheduleFixture`, build it and give Sam 30 recorded hours in W42:

```go
	repos.actuals = new(MockActualHoursRepository)
	repos.actuals.On("ByPersonWeeks", mock.Anything, schedPersonID, from, to).Return(map[isoweek.Week]float64{from: 30}, nil)
```

- in `newScheduleService`, pass `repos.actuals` after `repos.absences`: `ProvideScheduleService(repos.plans, repos.people, repos.absences, repos.actuals, ProvideAuthzService(people), cfg)`
- add this case inside `TestScheduleService_Get`:

```go
	t.Run("weeks carry the hours recorded by timesheets", func(t *testing.T) {
		schedule, err := newScheduleService(scheduleFixture(t, scheduleAssignments(t))).Get(planner, sam)
		require.NoError(t, err)
		require.NotNil(t, schedule.Weeks[0].ActualHours)
		assert.InDelta(t, 30, *schedule.Weeks[0].ActualHours, hoursEpsilon)
		assert.Nil(t, schedule.Weeks[1].ActualHours, "W43 has no recorded hours")
	})
```

Append:

```go
func TestWithActualHours(t *testing.T) {
	w41, w42 := mustWeek(t, "2026-W41"), mustWeek(t, "2026-W42")
	weeks := []ScheduleWeek{{Week: w41}, {Week: w42}}
	withActualHours(weeks, map[isoweek.Week]float64{w41: 34.5})
	require.NotNil(t, weeks[0].ActualHours)
	assert.InDelta(t, 34.5, *weeks[0].ActualHours, 1e-9)
	assert.Nil(t, weeks[1].ActualHours, "a week without recorded hours stays null")
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./internal/services -run 'TestScheduleService_Get|TestWithActualHours' -v`
Expected: build failure, `unknown field actuals in struct literal` / `weeks[0].ActualHours undefined`.

- [ ] **Step 3: Implement**

In `internal/models/schedule.go`, add to `ScheduleWeek` after `Tentative`:

```go
	// ActualHours is what timesheets recorded for the week over all projects; nil when none.
	ActualHours *float64
```

In `internal/services/schedule.go`:
- add the field `actuals repositories.ActualHoursRepository` to `scheduleService` after `absences`
- add the parameter `actuals repositories.ActualHoursRepository` to `ProvideScheduleService` after `absences`, and `actuals: actuals,` to its literal
- in `Get`, replace the block from `weeks := isoweek.Range(in.From, in.To)` to the end of the function with:

```go
	worked, err := s.actuals.ByPersonWeeks(ctx, in.PersonID, in.From, in.To)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opPersonSchedule).WithResource(resourceSchedule)
	}
	weeks := isoweek.Range(in.From, in.To)
	scheduled := s.scheduleWeeks(*person, weeks, absences, assignments)
	withActualHours(scheduled, worked)
	return &PersonSchedule{
		WorkingWeekdays: s.cfg.WorkingWeekdays,
		Weeks:           scheduled,
		Days:            s.scheduleDays(person.ID, weeks, absences),
	}, nil
}

// withActualHours sets each week's actual hours from worked; the weeks it lacks stay nil.
func withActualHours(weeks []ScheduleWeek, worked map[isoweek.Week]float64) {
	for i := range weeks {
		if hours, ok := worked[weeks[i].Week]; ok {
			weeks[i].ActualHours = &hours
		}
	}
}
```

In `internal/dtos/schedule.go`, add to `ScheduleWeekResponse` after `Tentative`:

```go
	// ActualHours is null until a timesheet import recorded hours for the week.
	ActualHours *float64 `json:"actual_hours" example:"34"`
```

and in `NewScheduleResponse` add `ActualHours: w.ActualHours,` to the `ScheduleWeekResponse{...}` literal.

FX resolves the new `ProvideScheduleService` parameter by itself (`repositories.ProvideActualHoursRepository` is provided since Task 2). If `cmd/server/routes/router_test.go` or a handler test calls `ProvideScheduleService` directly, add `nil` for `actuals` there.

- [ ] **Step 4: Run the tests, regenerate Swagger, check the side**

Run: `go test ./internal/services ./internal/dtos -run 'Schedule|WithActualHours' -v`, then `make swagger-load`, then `gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`
Expected: PASS (`TestAppOptions_Validate` included); `docs/swagger.yaml` shows `actual_hours` on `dtos.ScheduleWeekResponse`; lint `0 issues`.

- [ ] **Step 5: Commit**

```bash
git add internal/models/schedule.go internal/services/schedule.go internal/services/schedule_test.go internal/dtos/schedule.go docs
git commit -m "feat(server): the member schedule shows actual hours per week

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 9: Seed actual hours for the demo

The demo needs Burn % with real numbers before anyone imports a file. The seed fills the template the way a planner would and imports it through `TimesheetService`, so it uses the same checks as the app. Data Migration burns above plan and Checkout Revamp below it. Patient App gets no actual hours, so the live import can still be shown on it.

**Files:**
- Create: `server/cmd/seed/actuals.go`
- Modify: `server/cmd/seed/main.go` (call it after `seedPlans`)
- Modify: `server/cmd/seed/plans_integration_test.go` (config fields and a test)

**Interfaces:**
- Consumes: `services.ProvideTimesheetService(actuals, people, projects, plans, authz, tx, clk, cfg)`, `TimesheetService.Template(ctx, services.TemplateInput{})` (zero weeks = the default range up to the current week), `TimesheetService.Import(ctx, services.ImportInput{File, Size})`, `constants.TimesheetColumn*`, `seedPlanner`, `seedConfig()`.
- Produces: `seedActuals(database *db.PostgresDB, cfg *config.Config) (int, error)`, returning how many rows it created.

- [ ] **Step 1: Write the failing integration test**

In `server/cmd/seed/plans_integration_test.go`, add the timesheet settings to `seedConfig()`:

```go
		TimesheetTemplateWeeks: 4, TimesheetMaxWeeks: 12, TimesheetMaxWeekHours: 80,
		TimesheetMaxBytes: 1 << 20, TimesheetMaxRows: 5000,
```

Then add the test:

```go
func TestSeedActuals_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	require.NoError(t, pdb.Transaction(seed))
	_, err := seedPlans(pdb, seedConfig())
	require.NoError(t, err)

	created, err := seedActuals(pdb, seedConfig())
	require.NoError(t, err)

	var saved []models.ActualHours
	require.NoError(t, pdb.Find(&saved).Error)
	assert.Len(t, saved, created)
	allowed := map[string]bool{}
	for name := range actualFactors {
		id, err := projectID(pdb.DB, name)
		require.NoError(t, err)
		allowed[id] = true
	}
	for _, row := range saved {
		assert.True(t, allowed[row.ProjectID], "only the seeded projects get actual hours")
		assert.Positive(t, row.Hours)
	}

	again, err := seedActuals(pdb, seedConfig())
	require.NoError(t, err)
	assert.Zero(t, again, "running the seed again adds nothing")
}
```

The number of rows depends on today's date (the template covers the last weeks up to now), so the test checks which projects get hours and that a second run adds nothing, not an exact count.

- [ ] **Step 2: Run it to see it fail**

Run: `cd server && go test -tags=integration ./cmd/seed -run TestSeedActuals_Postgres -v`
Expected: build failure, `undefined: seedActuals` and `undefined: actualFactors`.

- [ ] **Step 3: Write `seedActuals`**

Create `server/cmd/seed/actuals.go`:

```go
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

// seedActuals fills the timesheet template for the demo projects and imports it. Rows that
// already hold actual hours are left alone, so running the seed again adds nothing.
func seedActuals(database *db.PostgresDB, cfg *config.Config) (int, error) {
	clk, err := clock.ProvideClock(cfg)
	if err != nil {
		return 0, err
	}
	people := repositories.ProvidePersonRepository(database)
	finder := services.ProvidePersonFinder(services.ProvidePeopleService(people, db.ProvideTxManager(database), cfg))
	timesheets := services.ProvideTimesheetService(repositories.ProvideActualHoursRepository(database), people,
		repositories.ProvideProjectRepository(database), repositories.ProvidePlanRepository(database),
		services.ProvideAuthzService(finder), db.ProvideTxManager(database), clk, cfg)
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
	if err != nil || len(records) == 0 {
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
```

If `Template` or `Import` in Tasks 5–6 ended up with other parameter names, match them; the constructor's argument order is the one Task 6 fixed.

- [ ] **Step 4: Call it from the seed**

In `server/cmd/seed/main.go`, after the `Scenarios added` log line:

```go
	actuals, err := seedActuals(database, cfg)
	if err != nil {
		logger.Sugar.Fatalf("seed actual hours: %v", err)
	}
	logger.Sugar.Infof("Actual hours added: %d rows", actuals)
```

- [ ] **Step 5: Run the tests**

Run: `cd server && go test -tags=integration ./cmd/seed -v && go build ./... && go vet ./... && make lint`
Expected: PASS, and lint reports 0 issues.

Then on the local stack: `make seed` twice. The first run logs `Actual hours added: N rows` with N > 0 (today is in W42, so Data Migration has W41–W42 and Checkout Revamp has W42); the second logs `0 rows`.

- [ ] **Step 6: Commit**

```bash
git add server/cmd/seed/actuals.go server/cmd/seed/main.go server/cmd/seed/plans_integration_test.go
git commit -m "feat(seed): actual hours for Data Migration and Checkout Revamp

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Client

Work in `client/`. Every client commit: `npm run format`, then `npx tsc -b && npm run lint && npm test && npm run build`. Never commit while lint reports anything. New text gets a key in `src/shared/i18n/language-key.ts`, in its case-insensitive sorted place; `npx tsc -b` lists any key that is missing.

### Task 10: File download and blob errors through the HTTP client

**Files:**
- Create: `client/src/shared/api/blob.ts`, `client/src/shared/api/blob.test.ts`
- Modify: `client/src/shared/api/http.ts`
- Create: `client/src/utils/functions/file.ts`, `client/src/utils/functions/file.test.ts`
- Modify: `client/src/utils/functions/index.ts`

**Interfaces:**
- Produces: `readBlobEnvelope(data: unknown): Promise<unknown>` (`@/shared/api/blob`); `request<Blob>({ …, responseType: 'blob' })` resolves with the `Blob`; a failed blob request rejects with an `ApiError` that has the server's `code` and `message`; `saveFile(blob: Blob, fileName: string): void` (`@/utils/functions`).

- [ ] **Step 1: Write the failing tests**

`client/src/shared/api/blob.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { readBlobEnvelope } from './blob'

const ENVELOPE = { meta: { error_code: 'WEEK_RANGE_TOO_LONG', message: 'The range may cover at most 12 weeks' }, data: null }
const CSV = 'person_email,week\nada@example.test,2026-W41\n'

describe('readBlobEnvelope', () => {
  it('reads a JSON envelope that arrived as a Blob', async () => {
    const blob = new Blob([JSON.stringify(ENVELOPE)], { type: 'application/json' })
    expect(await readBlobEnvelope(blob)).toEqual(ENVELOPE)
  })

  it('gives undefined for a Blob that is not JSON', async () => {
    expect(await readBlobEnvelope(new Blob([CSV]))).toBeUndefined()
  })

  it('leaves other data unchanged', async () => {
    expect(await readBlobEnvelope(ENVELOPE)).toBe(ENVELOPE)
  })
})
```

`client/src/utils/functions/file.test.ts`:

```ts
import { afterEach, describe, expect, it, vi } from 'vitest'
import { saveFile } from './file'

const OBJECT_URL = 'blob:http://localhost/template'
const FILE_NAME = 'timesheet-2026-W39-2026-W42.csv'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('saveFile', () => {
  it('clicks a download link for the blob, then releases its URL', () => {
    const blob = new Blob(['a,b\n'])
    URL.createObjectURL = vi.fn(() => OBJECT_URL)
    URL.revokeObjectURL = vi.fn()
    const links: HTMLAnchorElement[] = []
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(function (this: HTMLAnchorElement) {
      links.push(this)
    })

    saveFile(blob, FILE_NAME)

    expect(URL.createObjectURL).toHaveBeenCalledWith(blob)
    expect(links).toHaveLength(1)
    expect(links[0].download).toBe(FILE_NAME)
    expect(links[0].href).toBe(OBJECT_URL)
    expect(URL.revokeObjectURL).toHaveBeenCalledWith(OBJECT_URL)
  })
})
```

- [ ] **Step 2: Run them to verify they fail**

Run: `npx vitest run src/shared/api/blob.test.ts src/utils/functions/file.test.ts`
Expected: FAIL, `Failed to resolve import "./blob"` and `"./file"`.

- [ ] **Step 3: Implement**

`client/src/shared/api/blob.ts`:

```ts
/**
 * A request made with `responseType: 'blob'` receives even its error envelope as a Blob. This
 * reads it back as JSON so the error keeps the server's code and message; a Blob that is not
 * JSON gives undefined, and any other data is returned unchanged.
 */
export async function readBlobEnvelope(data: unknown): Promise<unknown> {
  if (!(data instanceof Blob)) return data
  try {
    return JSON.parse(await data.text()) as unknown
  } catch {
    return undefined
  }
}
```

`client/src/shared/api/http.ts`: import it (`import { readBlobEnvelope } from './blob'`), then change the two interceptors:

```ts
http.interceptors.response.use(
  (response) => {
    captureCsrfToken(response)
    // A file download (responseType 'blob') has no envelope: the Blob itself is the data.
    if (response.data instanceof Blob) return { data: response.data, meta: undefined } as unknown as AxiosResponse
    // `?? {}` covers empty bodies such as 204 No Content.
    const { data, meta } = (response.data ?? {}) as Partial<ApiEnvelope<unknown>>
    // Axios types expect an AxiosResponse here; `requestWithMeta` types the real shape.
    return { data, meta } as unknown as AxiosResponse
  },
  async (error: unknown) => {
    if (axios.isAxiosError(error) && error.response) {
      captureCsrfToken(error.response)
      // A failed download carries its error envelope as a Blob.
      error.response.data = await readBlobEnvelope(error.response.data)
    }
    return Promise.reject(toApiError(error))
  },
)
```

`client/src/utils/functions/file.ts`:

```ts
/** Saves a Blob as a file that the browser downloads, e.g. a CSV template. */
export function saveFile(blob: Blob, fileName: string): void {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  link.click()
  URL.revokeObjectURL(url)
}
```

`client/src/utils/functions/index.ts`: add `export { saveFile } from './file'`.

- [ ] **Step 4: Run the tests to verify they pass, then check the side**

Run: `npx vitest run src/shared/api/blob.test.ts src/utils/functions/file.test.ts`, then `npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: PASS; lint clean.

- [ ] **Step 5: Commit**

```bash
git add src/shared/api src/utils/functions
git commit -m "feat(client): download files and read blob errors through the HTTP client

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 11: Programs and projects show burn from actual hours

**Files:**
- Modify: `client/src/shared/programs/types.ts`, `client/src/shared/programs/constants.ts`, `client/src/shared/programs/utils.ts`, `client/src/shared/programs/index.ts`
- Create: `client/src/shared/programs/utils.test.ts`
- Modify: `client/src/features/program-overview/components/metrics-card/component.tsx`
- Create: `client/src/features/program-overview/components/metrics-card/component.test.tsx`
- Modify: `client/src/features/program-projects/components/project-row/component.tsx`
- Modify: `client/src/features/risks/component.test.tsx` (fixture)
- Modify: `client/src/shared/i18n/language-key.ts`

**Interfaces:**
- Consumes: `actual_hours`, `burn_percent` on the server's progress metrics (Task 1, Task 3).
- Produces: `ProgressMetrics.actual_hours: number | null`, `ProgressMetrics.burn_percent: number | null`; `BurnMetric { label: PlainKey; value: number | null }`; `BURN_LABELS`; `burnMetric(metrics: ProgressMetrics): BurnMetric` (`@/shared/programs`).

- [ ] **Step 1: Write the failing tests**

`client/src/shared/programs/utils.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import type { ProgressMetrics } from './types'
import { burnMetric } from './utils'

const metrics = (overrides: Partial<ProgressMetrics>): ProgressMetrics => ({
  time_percent: 40,
  scope_percent: 30,
  staffing_percent: 80,
  planned_burn_percent: 20,
  burn_percent: null,
  budget_hours: 400,
  planned_hours: 320,
  forecast_hours: 320,
  gap_hours: 80,
  planned_to_date_hours: 80,
  actual_hours: null,
  behind_on_delivery: false,
  understaffed: false,
  at_risk: false,
  over_planned: false,
  past_end: false,
  ...overrides,
})

describe('burnMetric', () => {
  it('shows planned burn until the project has actual hours', () => {
    expect(burnMetric(metrics({}))).toEqual({ label: 'Planned burn', value: 20 })
  })

  it('shows burn from actual hours once they exist, even at 0 %', () => {
    expect(burnMetric(metrics({ actual_hours: 0, burn_percent: 0 }))).toEqual({ label: 'Burn', value: 0 })
  })
})
```

`client/src/features/program-overview/components/metrics-card/component.test.tsx`:

```tsx
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import type { ProgramOverview, ProgressMetrics } from '@/shared/programs'
import { MetricsCard } from './component'

const metrics = (overrides: Partial<ProgressMetrics>): ProgressMetrics => ({
  time_percent: 40,
  scope_percent: 30,
  staffing_percent: 80,
  planned_burn_percent: 20,
  burn_percent: null,
  budget_hours: 400,
  planned_hours: 320,
  forecast_hours: 320,
  gap_hours: 80,
  planned_to_date_hours: 80,
  actual_hours: null,
  behind_on_delivery: false,
  understaffed: false,
  at_risk: false,
  over_planned: false,
  past_end: false,
  ...overrides,
})

const overview = (values: ProgressMetrics): ProgramOverview => ({
  program_id: 'g1',
  projects: 2,
  metrics: values,
  staffing_by_expertise: [],
})

afterEach(cleanup)

describe('MetricsCard', () => {
  it('shows planned burn and no actual figure before any timesheet import', () => {
    render(<MetricsCard overview={overview(metrics({}))} />)
    expect(screen.getByText('Planned burn')).toBeTruthy()
    expect(screen.queryByText('Actual')).toBeNull()
  })

  it('shows burn and the actual hours once they are imported', () => {
    render(<MetricsCard overview={overview(metrics({ actual_hours: 100, burn_percent: 25 }))} />)
    expect(screen.getByText('Burn')).toBeTruthy()
    expect(screen.queryByText('Planned burn')).toBeNull()
    expect(screen.getByText('Actual').nextElementSibling?.textContent).toBe('100 h')
  })
})
```

In `client/src/features/risks/component.test.tsx`, add `burn_percent: null,` and `actual_hours: null,` to the `metrics` fixture.

- [ ] **Step 2: Run them to verify they fail**

Run: `npx vitest run src/shared/programs/utils.test.ts src/features/program-overview/components/metrics-card/component.test.tsx`
Expected: FAIL, `burnMetric is not a function` / `Unable to find an element with the text: Actual`.

- [ ] **Step 3: Implement**

`client/src/shared/programs/types.ts`: add `import type { PlainKey } from '@/shared/i18n'` and, in `ProgressMetrics`:

```ts
  /** Actual hours ÷ budget; null until a timesheet import recorded actual hours. */
  burn_percent: number | null
```

after `planned_burn_percent`, and

```ts
  /** Σ actual hours from timesheet imports; null until any are recorded. */
  actual_hours: number | null
```

after `planned_to_date_hours`. Then add:

```ts
/** The burn a progress view shows, with its label. */
export interface BurnMetric {
  label: PlainKey
  value: number | null
}
```

`client/src/shared/programs/constants.ts`:

```ts
/** Labels of the two burns: from actual hours, or from planned hours until actual hours exist. */
export const BURN_LABELS = {
  actual: 'Burn',
  planned: 'Planned burn',
} as const satisfies Record<string, PlainKey>
```

`client/src/shared/programs/utils.ts` (import `BURN_LABELS` and the `BurnMetric` type):

```ts
/** The burn to show: from actual hours once the project has any, else planned burn. */
export function burnMetric(metrics: ProgressMetrics): BurnMetric {
  return metrics.burn_percent === null
    ? { label: BURN_LABELS.planned, value: metrics.planned_burn_percent }
    : { label: BURN_LABELS.actual, value: metrics.burn_percent }
}
```

`client/src/shared/programs/index.ts`: export `BURN_LABELS`, `burnMetric` and the `BurnMetric` type next to their neighbours.

`metrics-card/component.tsx`: import `burnMetric` from `@/shared/programs`; before the JSX add `const burn = burnMetric(metrics)`; replace the Planned burn bar with

```tsx
        <MetricBar label={t(burn.label)} value={burn.value} />
```

change the figures grid to `lg:grid-cols-6`, and after the **Planned to date** figure add

```tsx
        {metrics.actual_hours !== null && <Figure label={t('Actual')} hours={metrics.actual_hours} />}
```

Update its doc comment to "Time, Scope, Burn and Staffing with their warnings, and the cost figures."

`project-row/component.tsx`: import `burnMetric`; before the JSX add `const burn = metrics ? burnMetric(metrics) : null`; change the condition to `{metrics && burn && (` and replace the Planned burn bar with

```tsx
            <MetricBar compact label={t(burn.label)} value={burn.value} />
```

Add the keys `Actual` and `Burn` to `language-key.ts`.

- [ ] **Step 4: Run the tests to verify they pass, then check the side**

Run: `npx vitest run src/shared/programs src/features/program-overview src/features/risks`, then `npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: PASS; lint clean.

- [ ] **Step 5: Commit**

```bash
git add src/shared/programs src/features/program-overview src/features/program-projects src/features/risks src/shared/i18n/language-key.ts
git commit -m "feat(client): programs and projects show burn from actual hours

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 12: The Import timesheet dialog on the People page

**Files:**
- Create: `client/src/resources/endpoints/timesheets.ts`; modify `client/src/resources/endpoints/index.ts`
- Modify: `client/src/resources/rbac/permissions.ts`
- Create: `client/src/features/timesheet-import/{index.ts,component.tsx,types.ts,constants.ts,utils.ts,api.ts,use-timesheet-import.ts,utils.test.ts,component.test.tsx}`
- Create: `client/src/features/timesheet-import/components/problem-table/{index.ts,component.tsx,types.ts}`
- Modify: `client/src/features/people/types.ts`, `client/src/features/people/component.tsx`
- Modify: `client/src/pages/people/component.tsx`
- Modify: `client/src/shared/i18n/language-key.ts`

**Interfaces:**
- Consumes: `request` with `responseType: 'blob'` and `saveFile` (Task 10); the server routes and `TIMESHEET_INVALID` with `data.problems` (Task 7).
- Produces: `TIMESHEET_ENDPOINTS` (`@/resources/endpoints`); permission `'timesheets:import'` (planner); `TimesheetImport` (`@/features/timesheet-import`); `PeopleTableProps.actions?: ReactNode`.

- [ ] **Step 1: Endpoints, permission and the feature's data layer**

`client/src/resources/endpoints/timesheets.ts`:

```ts
export const TIMESHEET_ENDPOINTS = {
  template: '/timesheets/template',
  import: '/timesheets/import',
} as const
```

`client/src/resources/endpoints/index.ts`: add `export { TIMESHEET_ENDPOINTS } from './timesheets'`.

`client/src/resources/rbac/permissions.ts`: add `'timesheets:import': [ROLES.PLANNER],` after `'people:manage'`.

`client/src/features/timesheet-import/types.ts`:

```ts
/** Problem codes the server reports for a timesheet import (TIMESHEET_INVALID). */
export type RowProblemCode =
  | 'COLUMN_MISSING'
  | 'CSV_MALFORMED'
  | 'PERSON_UNKNOWN'
  | 'PROJECT_UNKNOWN'
  | 'WEEK_INVALID'
  | 'WEEK_IN_FUTURE'
  | 'HOURS_INVALID'
  | 'ROW_DUPLICATE'

/** One problem of an imported file: the line, the column ('' for the whole row) and its code. */
export interface RowProblem {
  line: number
  column: string
  code: string
}

/** The data of a TIMESHEET_INVALID error. */
export interface TimesheetInvalidDetails {
  problems: RowProblem[]
}

/** What an import saved: new rows, replaced values and rows skipped for an empty actual_hours. */
export interface ImportResult {
  created: number
  updated: number
  skipped: number
}

/** The ISO weeks a template covers. */
export interface WeekRange {
  from: string
  to: string
}
```

`client/src/features/timesheet-import/constants.ts`:

```ts
import type { PlainKey } from '@/shared/i18n'
import type { RowProblemCode } from './types'

/** Error code of an import with row problems; the dialog lists them instead of a toast. */
export const TIMESHEET_INVALID = 'TIMESHEET_INVALID'

/** Multipart field the server reads the file from. */
export const TIMESHEET_FORM_FIELD = 'file'

/** Files the picker offers. */
export const TIMESHEET_FILE_ACCEPT = '.csv,text/csv'

/** Past weeks the template covers by default, up to the current week (TIMESHEET_TEMPLATE_WEEKS). */
export const DEFAULT_TEMPLATE_WEEKS = 4

/** The downloaded template is named `timesheet-<from>-<to>.csv`. */
export const TEMPLATE_FILE_PREFIX = 'timesheet'
export const CSV_EXTENSION = '.csv'

/** Names and ids of the dialog's fields. */
export const RANGE_FIELDS = { from: 'from', to: 'to' } as const
export const FIELD_IDS = { from: 'timesheet-from', to: 'timesheet-to', file: 'timesheet-file' } as const

/** What each problem code means. */
export const ROW_PROBLEM_LABELS: Record<RowProblemCode, PlainKey> = {
  COLUMN_MISSING: 'The header lacks this column.',
  CSV_MALFORMED: 'The file is not valid CSV from this line.',
  PERSON_UNKNOWN: 'No person has this email.',
  PROJECT_UNKNOWN: 'No project has this id.',
  WEEK_INVALID: 'Not an ISO week such as 2026-W41.',
  WEEK_IN_FUTURE: 'The week is after the current week.',
  HOURS_INVALID: 'Hours must be a number from 0 to the weekly limit.',
  ROW_DUPLICATE: 'The same person, project and week appear more than once.',
}
```

`client/src/features/timesheet-import/utils.ts`:

```ts
import type { ApiError } from '@/shared/api/types'
import type { PlainKey } from '@/shared/i18n'
import { addIsoWeeks, currentIsoWeek } from '@/utils/functions'
import { CSV_EXTENSION, DEFAULT_TEMPLATE_WEEKS, ROW_PROBLEM_LABELS, TEMPLATE_FILE_PREFIX, TIMESHEET_INVALID } from './constants'
import type { RowProblem, RowProblemCode, TimesheetInvalidDetails, WeekRange } from './types'

/** The last DEFAULT_TEMPLATE_WEEKS weeks, up to the current one. */
export function defaultRange(): WeekRange {
  const to = currentIsoWeek()
  return { from: addIsoWeeks(to, 1 - DEFAULT_TEMPLATE_WEEKS), to }
}

/** The downloaded template's name, e.g. timesheet-2026-W39-2026-W42.csv. */
export function templateFileName({ from, to }: WeekRange): string {
  return [TEMPLATE_FILE_PREFIX, from, to].join('-') + CSV_EXTENSION
}

/** The row problems of a failed import; empty for any other error. */
export function rowProblems(error: ApiError | null): RowProblem[] {
  if (error?.code !== TIMESHEET_INVALID) return []
  return (error.details as Partial<TimesheetInvalidDetails> | undefined)?.problems ?? []
}

/** What a problem code means, or null for a code this client does not know yet. */
export function problemLabel(code: string): PlainKey | null {
  return code in ROW_PROBLEM_LABELS ? ROW_PROBLEM_LABELS[code as RowProblemCode] : null
}
```

`client/src/features/timesheet-import/api.ts`:

```ts
import { TIMESHEET_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import { TIMESHEET_FORM_FIELD } from './constants'
import type { ImportResult, WeekRange } from './types'

export function downloadTemplate(range: WeekRange) {
  return request<Blob>({ url: TIMESHEET_ENDPOINTS.template, params: range, responseType: 'blob' })
}

export function importTimesheet(file: File) {
  const form = new FormData()
  form.append(TIMESHEET_FORM_FIELD, file)
  return request<ImportResult>({ url: TIMESHEET_ENDPOINTS.import, method: 'POST', data: form })
}
```

`client/src/features/timesheet-import/use-timesheet-import.ts`:

```ts
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { saveFile } from '@/utils/functions'
import { downloadTemplate, importTimesheet } from './api'
import { TIMESHEET_INVALID } from './constants'
import type { WeekRange } from './types'
import { templateFileName } from './utils'

/** Downloads the template for a range of weeks and saves it as a CSV file. */
export function useDownloadTemplate() {
  return useMutation({
    mutationFn: async (range: WeekRange) => saveFile(await downloadTemplate(range), templateFileName(range)),
  })
}

/**
 * Imports a filled-in template. Row problems stay in the dialog instead of a toast. Actual
 * hours feed the program, project and person views, so every active query refetches.
 */
export function useImportTimesheet() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => importTimesheet(file),
    meta: { quietErrorCodes: [TIMESHEET_INVALID] },
    onSuccess: () => queryClient.invalidateQueries(),
  })
}
```

- [ ] **Step 2: Write the failing tests**

`client/src/features/timesheet-import/utils.test.ts`:

```ts
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/shared/api/types'
import { defaultRange, problemLabel, rowProblems, templateFileName } from './utils'

const PROBLEM = { line: 5, column: 'person_email', code: 'PERSON_UNKNOWN' }

afterEach(() => {
  vi.useRealTimers()
})

describe('timesheet import utils', () => {
  it('defaults to the last four weeks up to this one', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-10-14T09:00:00Z'))
    expect(defaultRange()).toEqual({ from: '2026-W39', to: '2026-W42' })
  })

  it('names the template after its weeks', () => {
    expect(templateFileName({ from: '2026-W39', to: '2026-W42' })).toBe('timesheet-2026-W39-2026-W42.csv')
  })

  it('reads the row problems of TIMESHEET_INVALID only', () => {
    const invalid = new ApiError({ status: 400, code: 'TIMESHEET_INVALID', message: 'x', details: { problems: [PROBLEM] } })
    const other = new ApiError({ status: 400, code: 'WEEK_RANGE_TOO_LONG', message: 'x' })
    expect(rowProblems(invalid)).toEqual([PROBLEM])
    expect(rowProblems(other)).toEqual([])
    expect(rowProblems(null)).toEqual([])
  })

  it('labels known codes and leaves unknown ones', () => {
    expect(problemLabel('PERSON_UNKNOWN')).toBe('No person has this email.')
    expect(problemLabel('SOMETHING_NEW')).toBeNull()
  })
})
```

`client/src/features/timesheet-import/component.test.tsx`:

```tsx
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/shared/api/types'
import { TimesheetImport } from './component'

const hooks = vi.hoisted(() => ({
  download: vi.fn(),
  upload: vi.fn(),
  reset: vi.fn(),
  error: null as ApiError | null,
}))

vi.mock('./use-timesheet-import', () => ({
  useDownloadTemplate: () => ({ mutate: hooks.download, isPending: false }),
  useImportTimesheet: () => ({ mutate: hooks.upload, reset: hooks.reset, error: hooks.error, isPending: false }),
}))

const CSV = 'person_email,project_id,week,actual_hours\n'
const FILE_NAME = 'timesheet.csv'

beforeEach(() => {
  hooks.error = null
  vi.clearAllMocks()
})
afterEach(cleanup)

describe('TimesheetImport', () => {
  it('imports the chosen file', async () => {
    render(<TimesheetImport />)
    fireEvent.click(screen.getByRole('button', { name: 'Import timesheet' }))
    const file = new File([CSV], FILE_NAME, { type: 'text/csv' })
    fireEvent.change(await screen.findByLabelText('Timesheet file'), { target: { files: [file] } })
    fireEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(hooks.upload).toHaveBeenCalledWith(file, expect.anything())
  })

  it('downloads the template for the chosen weeks', async () => {
    render(<TimesheetImport />)
    fireEvent.click(screen.getByRole('button', { name: 'Import timesheet' }))
    fireEvent.click(await screen.findByRole('button', { name: 'Download template' }))
    expect(hooks.download).toHaveBeenCalledWith({ from: expect.any(String), to: expect.any(String) })
  })

  it('lists the row problems instead of saving', async () => {
    hooks.error = new ApiError({
      status: 400,
      code: 'TIMESHEET_INVALID',
      message: 'Some rows of the timesheet have problems; nothing was saved',
      details: {
        problems: [
          { line: 5, column: 'person_email', code: 'PERSON_UNKNOWN' },
          { line: 7, column: '', code: 'ROW_DUPLICATE' },
        ],
      },
    })
    render(<TimesheetImport />)
    fireEvent.click(screen.getByRole('button', { name: 'Import timesheet' }))
    expect(await screen.findByText('No person has this email.')).toBeTruthy()
    expect(screen.getByText('The same person, project and week appear more than once.')).toBeTruthy()
    expect(screen.getByText('person_email')).toBeTruthy()
  })
})
```

Note: opening the dialog calls `reset()`. The mocked `error` stays set, so the problem table still shows in the last test.

- [ ] **Step 3: Run them to verify they fail**

Run: `npx vitest run src/features/timesheet-import`
Expected: FAIL, `Failed to resolve import "./component"`.

- [ ] **Step 4: Write the components**

`components/problem-table/types.ts`:

```ts
import type { RowProblem } from '../../types'

export interface ProblemTableProps {
  problems: RowProblem[]
}
```

`components/problem-table/component.tsx`:

```tsx
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { problemLabel } from '../../utils'
import type { ProblemTableProps } from './types'

/** The problems of a failed import: line, column and what is wrong. */
export function ProblemTable({ problems }: ProblemTableProps) {
  const { t } = useTranslation()

  return (
    <Table scroll="both">
      <TableHeader>
        <TableRow>
          <TableHead>{t('Line')}</TableHead>
          <TableHead>{t('Column')}</TableHead>
          <TableHead>{t('Problem')}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {problems.map((problem) => {
          const label = problemLabel(problem.code)
          return (
            <TableRow key={`${problem.line}-${problem.column}-${problem.code}`}>
              <TableCell className="tabular-nums">{problem.line}</TableCell>
              <TableCell>{problem.column || t('—')}</TableCell>
              <TableCell>{label ? t(label) : problem.code}</TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
```

`components/problem-table/index.ts`: `export { ProblemTable } from './component'`

`client/src/features/timesheet-import/component.tsx`:

```tsx
import { Upload } from 'lucide-react'
import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { WeekPicker } from '@/shared/components/week-picker'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { ProblemTable } from './components/problem-table'
import { FIELD_IDS, RANGE_FIELDS, TIMESHEET_FILE_ACCEPT } from './constants'
import type { ImportResult, WeekRange } from './types'
import { useDownloadTemplate, useImportTimesheet } from './use-timesheet-import'
import { defaultRange, rowProblems } from './utils'

/** Import timesheet: download the template for past weeks, then upload it with actual hours. */
export function TimesheetImport() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [range, setRange] = useState<WeekRange>(defaultRange)
  const [file, setFile] = useState<File | null>(null)
  const download = useDownloadTemplate()
  const upload = useImportTimesheet()
  const problems = rowProblems(upload.error)

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    setFile(null)
    upload.reset()
  }
  const handleOpen = () => handleOpenChange(true)
  const handleWeekChange = (value: string, name: string) => setRange((current) => ({ ...current, [name]: value }))
  const handleDownload = () => download.mutate(range)
  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    setFile(event.target.files?.[0] ?? null)
    upload.reset()
  }
  const handleImported = ({ created, updated, skipped }: ImportResult) => {
    toast.success(t('{created} rows added, {updated} updated, {skipped} skipped.', { created, updated, skipped }))
    handleOpenChange(false)
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (file) upload.mutate(file, { onSuccess: handleImported })
  }

  return (
    <>
      <Button variant="outline" onClick={handleOpen}>
        <Upload />
        {t('Import timesheet')}
      </Button>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className="sm:max-w-2xl">
          <form className="flex min-h-0 flex-col gap-5" onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>{t('Import timesheet')}</DialogTitle>
              <DialogDescription>
                {t(
                  'Download the template for past weeks, fill in actual_hours, then import it. One row with a problem saves nothing.',
                )}
              </DialogDescription>
            </DialogHeader>
            <section className="flex flex-col gap-2">
              <h3 className="text-sm font-medium">{t('Download the template')}</h3>
              <div className="flex flex-wrap items-end gap-3">
                <div className="flex flex-col gap-1">
                  <Label htmlFor={FIELD_IDS.from}>{t('From week')}</Label>
                  <WeekPicker
                    id={FIELD_IDS.from}
                    name={RANGE_FIELDS.from}
                    value={range.from}
                    onChange={handleWeekChange}
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <Label htmlFor={FIELD_IDS.to}>{t('To week')}</Label>
                  <WeekPicker id={FIELD_IDS.to} name={RANGE_FIELDS.to} value={range.to} onChange={handleWeekChange} />
                </div>
                <Button type="button" variant="outline" disabled={download.isPending} onClick={handleDownload}>
                  {t('Download template')}
                </Button>
              </div>
            </section>
            <section className="flex flex-col gap-2">
              <h3 className="text-sm font-medium">{t('Import the filled-in file')}</h3>
              <Label htmlFor={FIELD_IDS.file}>{t('Timesheet file')}</Label>
              <Input id={FIELD_IDS.file} type="file" accept={TIMESHEET_FILE_ACCEPT} onChange={handleFileChange} />
            </section>
            {problems.length > 0 && (
              <div className="flex max-h-64 min-h-0 flex-col gap-2">
                <p className="text-sm text-destructive">{t('Nothing was saved. Fix these rows and import again.')}</p>
                <ProblemTable problems={problems} />
              </div>
            )}
            <DialogFooter>
              <Button type="submit" disabled={!file || upload.isPending}>
                {upload.isPending ? t('Importing…') : t('Import')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  )
}
```

`client/src/features/timesheet-import/index.ts`: `export { TimesheetImport } from './component'`

People page slot:
- `features/people/types.ts`: add `import type { ReactNode } from 'react'` and, in `PeopleTableProps`, `/** Buttons shown before Add person, e.g. Import timesheet. */ actions?: ReactNode`.
- `features/people/component.tsx`: take `actions` from the props and render the Section action as

```tsx
      action={
        <div className="flex flex-wrap gap-2">
          {actions}
          <Button onClick={handleAdd}>{t('Add person')}</Button>
        </div>
      }
```

- `pages/people/component.tsx`: import `TimesheetImport` from `@/features/timesheet-import` and add `actions={can(user?.role, 'timesheets:import') && <TimesheetImport />}` to the `<PeopleTable … />` element.

Add the new keys to `language-key.ts` (run `npx tsc -b` to list them): `Column`, `Download template`, `Download the template`, `Download the template for past weeks, fill in actual_hours, then import it. One row with a problem saves nothing.`, `Hours must be a number from 0 to the weekly limit.`, `Import`, `Import the filled-in file`, `Import timesheet`, `Importing…`, `Line`, `No person has this email.`, `No project has this id.`, `Not an ISO week such as 2026-W41.`, `Nothing was saved. Fix these rows and import again.`, `Problem`, `The file is not valid CSV from this line.`, `The header lacks this column.`, `The same person, project and week appear more than once.`, `The week is after the current week.`, `Timesheet file`, `To week`, `{created} rows added, {updated} updated, {skipped} skipped.`. Skip any that already exist.

- [ ] **Step 5: Run the tests to verify they pass, then check the side**

Run: `npx vitest run src/features/timesheet-import`, then `npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: PASS; lint clean. The structure check accepts the folder, which holds only allowed file names and has one child level.

- [ ] **Step 6: Try it in the browser**

With the stack and the server from Tasks 1–8 running: `npm run dev`, sign in as `planner`, open **People** → **Import timesheet** → **Download template**. Expected: `timesheet-<from>-<to>.csv` downloads with the seed's baseline rows. Fill `actual_hours` for two rows and leave one email wrong, then import. Expected: the problem table lists that line and nothing is saved. Fix the email and import again. Expected: a toast with the counts, and the program page shows **Burn** and **Actual**.

- [ ] **Step 7: Commit**

```bash
git add src/resources src/features/timesheet-import src/features/people src/pages/people src/shared/i18n/language-key.ts
git commit -m "feat(client): import a timesheet from the People page

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 13: The member calendar shows worked hours

**Files:**
- Modify: `client/src/shared/people/types.ts` (from member-calendar: `ScheduleWeek`)
- Create: `client/src/features/person-calendar/components/week-row/components/worked-hours/{index.ts,component.tsx,types.ts,component.test.tsx}`
- Modify: `client/src/features/person-calendar/components/week-row/component.tsx` (from member-calendar: `WeekRow`)
- Modify: every test fixture that builds a `ScheduleWeek`: `client/src/features/person-calendar/component.test.tsx` (its `week()` helper) and any other file `npx tsc -b` reports
- Modify: `client/src/shared/i18n/language-key.ts`

**Interfaces:**
- Consumes: `actual_hours` on each schedule week (Task 8).
- Produces: `ScheduleWeek.actual_hours: number | null` (`@/shared/people`); `WorkedHours({ hours: number | null })`.

- [ ] **Step 1: Write the failing test** — `worked-hours/component.test.tsx`:

```tsx
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { WorkedHours } from './component'

afterEach(cleanup)

describe('WorkedHours', () => {
  it('shows the hours a timesheet recorded', () => {
    render(<WorkedHours hours={34.5} />)
    expect(screen.getByText('Worked 34.5 h')).toBeTruthy()
  })

  it('shows nothing before any import', () => {
    const { container } = render(<WorkedHours hours={null} />)
    expect(container.textContent).toBe('')
  })
})
```

- [ ] **Step 2: Run it to verify it fails**

Run: `npx vitest run src/features/person-calendar/components/week-row/components/worked-hours`
Expected: FAIL, `Failed to resolve import "./component"`.

- [ ] **Step 3: Implement**

`client/src/shared/people/types.ts`, in `ScheduleWeek` after `tentative`:

```ts
  /** Hours a timesheet import recorded for the week; null when none are recorded. */
  actual_hours: number | null
```

`worked-hours/types.ts`:

```ts
export interface WorkedHoursProps {
  /** Recorded actual hours of the week; null hides the line. */
  hours: number | null
}
```

`worked-hours/component.tsx`:

```tsx
import { useTranslation } from '@/shared/i18n'
import type { WorkedHoursProps } from './types'

/** "Worked N h" under a week's load, once a timesheet import recorded actual hours. */
export function WorkedHours({ hours }: WorkedHoursProps) {
  const { t } = useTranslation()
  if (hours === null) return null

  return <span className="text-xs text-muted-foreground tabular-nums">{t('Worked {hours} h', { hours })}</span>
}
```

`worked-hours/index.ts`: `export { WorkedHours } from './component'`

`week-row/component.tsx`: import `WorkedHours` from `./components/worked-hours`, and in the last `TableCell`, wrap the load button so the worked hours sit under it:

```tsx
      <TableCell className="text-right">
        <div className="flex flex-col items-end gap-0.5">
          <button
            type="button"
            aria-pressed={selected}
            aria-label={t('Hours in {week}', { week: week.week })}
            className={cn(LOAD_BUTTON_CLASS, BAND_CLASSES[week.band], week.tentative && 'bg-hatched')}
            onClick={handleSelect}
          >
            {week.band === OVERLOADED_BAND && <TriangleAlert className="size-3.5" aria-hidden />}
            {t('{load}/{capacity} h', { load: week.load, capacity: week.capacity })}
          </button>
          <WorkedHours hours={week.actual_hours} />
        </div>
      </TableCell>
```

Add `actual_hours: null,` to every `ScheduleWeek` fixture that `npx tsc -b` reports (at least the `week()` helper in `features/person-calendar/component.test.tsx`), and add the key `Worked {hours} h` to `language-key.ts`.

- [ ] **Step 4: Run the tests to verify they pass, then check the side**

Run: `npx vitest run src/features/person-calendar`, then `npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: PASS; lint clean (worked-hours is the third and last allowed level).

- [ ] **Step 5: Commit**

```bash
git add src/shared/people src/features/person-calendar src/shared/i18n/language-key.ts
git commit -m "feat(client): the member calendar shows worked hours

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 14: Check end to end, mark the feature done, open the PR

**Files:**
- Modify: `docs/features/timesheet-import.md`, `docs/features/README.md`

- [ ] **Step 1: Run every check**

Run at the repo root: `make check`, then `cd server && go test -tags=integration ./internal/services -run 'Timesheet|ActualHours|Schedule' -v`
Expected: lint, tests and build pass on both sides; the integration tests PASS.

- [ ] **Step 2: Walk the acceptance criteria on the real stack** (`make up`, seed data from `make seed`, client on `:5173`):
  1. Template for W41–W42: one row per person × project × week with baseline hours, with planned hours.
  2. Import actual hours: the program shows Actual hours and Burn % = actual ÷ budget, and so do its projects.
  3. A file with an unknown email on line 5 and a future week on line 7: nothing saved, both problems listed.
  4. The same file twice: rows updated, not duplicated (the toast says `0 rows added, N updated`).
  5. As `lead` and as `specialist`: `GET /api/v1/timesheets/template` and `POST /api/v1/timesheets/import` answer `403`.
  6. The person page and My schedule show "Worked N h" on the imported weeks.

- [ ] **Step 3: Update the docs**

In `docs/features/timesheet-import.md`, set `**Status:** Done (PR #<number>)`, tick every task, and add an **As built** section after "Not in this feature" with what differs from the spec:
- `updated_by` holds the planner's Keycloak subject, not a person id: a planner may have no person record. Root README §6 already shows it that way.
- `ImportResult` lives in `models`, so `dtos` can map it.
- The download is named `timesheet.csv` by the server; the client saves it as `timesheet-<from>-<to>.csv`.
- Text cells starting with `= + - @` are written with a leading `'` in the template.
- A file over the limits gets `413 TIMESHEET_TOO_LARGE`.

In `docs/features/README.md`, set row 12 to **Done**.

- [ ] **Step 4: Commit and open the PR**

```bash
git add docs README.md
git commit -m "docs: timesheet import is built

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git push -u origin feat/timesheet-import
gh pr create --base main --title "Timesheet import: actual hours and burn" --body "## What

- Planners download a CSV template (person × project × week from the baseline) and import actual hours; one bad row saves nothing and the dialog lists every problem.
- Programs and projects show Actual hours and Burn % (actual ÷ budget), falling back to Planned burn.
- The member calendar shows Worked N h.

## Checks

- make check; integration tests for the repository and the all-or-nothing import
- the six acceptance criteria of docs/features/timesheet-import.md, walked on the seeded stack

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```

Wait for CI. Merge only when the user says "merge", with `gh pr merge <n> --rebase` (keep the branch, never `--admin`).
