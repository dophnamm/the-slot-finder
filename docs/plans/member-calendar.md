# Member calendar Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A planner or lead opens `/people/:id` and sees that person's month: absences and holidays per day, load against capacity per ISO week, and the hours per project behind each week. A specialist sees the same calendar for themselves on `/me`.

**Architecture:** A new server endpoint `GET /people/{id}/schedule?from=&to=` computes the weeks with the same capacity, load and band code as the heatmap (`capacity.Weekly`, `capacity.Tentative`, `utilizationBand`), plus the per-day absences and holidays. The client adds a `shared/people` module (types, API, hooks), a `features/person-calendar` month grid, a `pages/person-detail` page and replaces the My schedule placeholder. People rows open the person page. A planner can approve a planned absence (`POST /absences/{id}/approve`), which records an `AbsenceAdded` change event like any absence change.

**Tech Stack:** Go (Echo v5, Uber FX, GORM, testify), React + TypeScript (TanStack Query, React Router, shadcn on Base UI, Tailwind v4, dayjs through `@/utils/functions`, Vitest + Testing Library).

**Spec:** [`docs/features/member-calendar.md`](../features/member-calendar.md). Read it before starting; this plan argues from it.

## Global Constraints

- All code comments, docs and UI text are in English. UI text goes through `t()` with the English text as the key, declared in `client/src/shared/i18n/language-key.ts` in its alphabetical place (case-insensitive; keys that start with `{` sit at the end of the list). Use `’`, never `'`, inside a key.
- **Never hardcode.** Paths, roles, messages and error codes are constants in `server/internal/constants`; settings are fields in `server/internal/config/config.go` and lines in `server/cmd/server/.env.example`. Client endpoints live in `client/src/resources/endpoints`, routes in `ROUTES`, permissions in `PERMISSIONS`, formats in `DATE_FORMAT`, everything else in a `constants.ts`.
- Setting `SCHEDULE_MAX_WEEKS`, default `8` → config field `ScheduleMaxWeeks`.
- Path constant `constants.PathPersonSchedule = "/people/:id/schedule"`.
- Approve: `constants.PathAbsenceApprove = "/absences/:id/approve"` (planner only); `AvailabilityService.Approve(ctx context.Context, id string) (*models.Absence, error)`; `404 ABSENCE_NOT_FOUND` when missing; `409` with the new code `constants.AbsenceNotPlanned = "ABSENCE_NOT_PLANNED"` when the absence is already approved or is a holiday. The change event is `constants.EventAbsenceAdded` with the new status in its payload: the constants hold only `AbsenceAdded` and `AbsenceRemoved`, and the spec says to use `AbsenceAdded`.
- Service interface `services.ScheduleService` with `Get(ctx context.Context, in ScheduleInput) (*PersonSchedule, error)`; `ScheduleInput{PersonID string; From, To isoweek.Week}`.
- Types `PersonSchedule`, `ScheduleWeek`, `ScheduleProject`, `ScheduleDay`, `ScheduleAbsence` with exactly the fields listed in Task 1. They are defined in `internal/models/schedule.go` (the `dtos` package cannot import `services`, which already imports `dtos`) and re-exported in `services` as type aliases, so `services.PersonSchedule` is the same type. Later features add `ActualHours *float64` to `ScheduleWeek` and `Source string` to `ScheduleAbsence` in `models/schedule.go`; do not add them here.
- JSON of the schedule: `working_weekdays` (ISO numbers, 1 = Monday … 7 = Sunday), `weeks[]{week, capacity, load, utilization, band, tentative, projects[]{program_id, program_name, project_id, project_name, expertise, hours}}`, `days[]{date, absences[]{id, type, status, hours_per_day}, holidays[]{id}}`. Holidays have no name in the data model, so each holiday is sent by its `id` (see Review Focus and the report).
- Client names: endpoint `PEOPLE_ENDPOINTS.schedule(id)`; hook `useSchedule(personId: string, from: string, to: string)`; types `PersonSchedule`, `ScheduleWeek`, `ScheduleDay`, `ScheduleAbsence` exported from `@/shared/people`; `features/person-calendar` exports `PersonCalendar` with props `{ personId: string }`; `pages/person-detail` exports `PersonDetailPage`; `ROUTES.personDetail = '/people/:id'`; permission `'people:view'` (planner, lead).
- Server and client never share a commit. Small commits. Each commit passes its own side's checks:
  - server (`cd server`): `go build ./... && go vet ./... && go test ./...` and `make lint`
  - client (`cd client`): `npx tsc -b && npm run lint && npm test && npm run build` (run `npm run format` first)
- Every commit message ends with a blank line and `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- Swagger (`server/docs`) is regenerated, never edited: `make swagger-load`, or without the `swag` binary `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d ./cmd/server,./internal/handlers,./internal/middlewares,./internal/services,./internal/repositories,./internal/models,./internal/utils,./internal/config,./internal/constants,./internal/dtos,./internal/logger,./internal/db --output ./docs` from `server/`.
- No new tables, no migration.
- Branch: `feat/member-calendar` from `main`. One PR for the feature. Do not merge.

## Review Focus

1. **A month that spans six ISO weeks** (November 2026: W44–W49), and January, which starts in the previous ISO year (2026-W53). The grid must show every week and ask the server for them, inside `SCHEDULE_MAX_WEEKS = 8`. Pinned by the `monthIsoWeeks` tests in Task 8 and the "next month" test in Task 9.
2. **Leave that starts before the visible range, a company holiday inside it, and another person's leave.** The day cells show only the days inside the range, holidays on everyone's calendar, never another person's leave. Pinned by the days case in Task 1.
3. **Who may act.** A specialist asking for someone else's schedule gets `403` while a lead reading anyone gets `200` (Tasks 1 and 2). A lead or specialist approving an absence gets `403`, a planner `200` (Task 4, through the real router). The Approve button shows only where the list is given `canApprove` (Task 7).
4. **Approving twice, or approving a holiday.** A double click or a stale page sends a second approve; it must answer `409 ABSENCE_NOT_PLANNED` and record no second change event. Pinned in Task 4.
5. **Load on the calendar equals load on the heatmap** for the same week, and a person without baseline hours shows `0 / capacity` with `[]` lists, never `null`. Both use the same allocations and `utilizationBand`; pinned in Task 1 (40 h on 28 h → overloaded; empty assignments) and Task 2 (`"projects":[]`).

## File Structure

Server (`server/`):

| File | Change | Responsibility |
|---|---|---|
| `internal/models/schedule.go` | Create | The computed schedule types (never stored) |
| `internal/services/schedule.go` | Create | `ScheduleService`: access check, range check, weeks and days |
| `internal/services/schedule_test.go` | Create | Table tests with the existing repository mocks |
| `internal/dtos/schedule.go` | Create | `ScheduleRequest`, response types, `NewScheduleResponse` |
| `internal/handlers/schedule.go` | Create | `ScheduleHandler.Get` with Swagger annotations |
| `internal/handlers/schedule_test.go` | Create | httptest: 200 with JSON field names, 400 on bad query |
| `internal/handlers/person.go` | Modify | `PersonHandler.Get` for `GET /people/{id}` (the person page header) |
| `internal/handlers/person_test.go` | Modify | `fakePeople.Get` and `TestPersonHandler_Get` |
| `internal/services/absence.go`, `internal/services/absence_test.go` | Modify | `AvailabilityService.Approve` and its table test |
| `internal/handlers/absence.go` | Modify | `AbsenceHandler.Approve` |
| `internal/handlers/absence_test.go` | Create | httptest for Approve |
| `internal/constants/error_codes.go` | Modify | `AbsenceNotPlanned` |
| `internal/config/config.go` | Modify | `ScheduleMaxWeeks` |
| `cmd/server/.env.example` | Modify | `SCHEDULE_MAX_WEEKS=8` |
| `internal/constants/routes.go` | Modify | `PathPersonSchedule`, `PathAbsenceApprove` |
| `internal/constants/roles.go` | Modify | `PersonScheduleRoles` |
| `internal/constants/messages.go` | Modify | `MsgScheduleRetrieved`, `MsgPersonRetrieved`, `MsgAbsenceApproved`, `MsgAbsenceNotPlanned`, `MsgAbsenceApproveFailed` |
| `cmd/server/routes/router.go` | Modify | New handler parameter, three routes |
| `cmd/server/routes/router_test.go` | Modify | Pass the new handler, `newTestRouterWith`, RBAC cases |
| `cmd/server/main.go` | Modify | FX providers, `NewHTTPServer` parameter |
| `docs/` (Swagger) | Regenerate | |

Client (`client/src/`):

| File | Change | Responsibility |
|---|---|---|
| `shared/utilization/{index,types,constants}.ts` | Create | `UtilizationBand`, `BAND_CLASSES`, `BAND_LABELS`, `BAND_ORDER`, `OVERLOADED_BAND`, moved out of the heatmap feature |
| `features/heatmap/{types,constants}.ts` | Modify | Re-export the band type and constants from `@/shared/utilization` |
| `shared/people/{index,types,constants,api,use-people}.ts` | Create | `Person`, `PersonSkill`, absence types and labels (moved), schedule types, `getPerson`, `getSchedule`, `usePerson`, `useSchedule`, query keys |
| `features/people/{types,constants}.ts` | Modify | Re-export `Person`, `PersonSkill`, `PEOPLE_QUERY_KEY` from `@/shared/people` |
| `features/absences/{types,constants}.ts`, `features/absences/use-absences.ts` | Modify | Re-export the absence types and labels; refresh insights after a change |
| `features/holidays/use-holidays.ts` | Modify | Refresh insights after a change |
| `features/absences/api.ts`, `features/absences/use-absences.ts`, `features/absences/use-absences.test.tsx` | Modify / Create | `approveAbsence`, `useApproveAbsence` and its test |
| `features/absences/component.tsx`, `features/absences/components/absence-row/**` | Modify | Approve button on planned absences; row test |
| `resources/endpoints/availability.ts` | Modify | `ABSENCE_ENDPOINTS.approve(id)` |
| `resources/endpoints/people.ts` | Modify | `schedule(id)` |
| `utils/constants/date.ts`, `utils/functions/date.ts`, `utils/functions/index.ts` | Modify | Month formats and helpers |
| `utils/functions/date.test.ts` | Create | Month helper tests |
| `features/person-calendar/**` | Create | `PersonCalendar`, `MonthNav`, `WeekRow` → `DayCell`, `WeekDetail` |
| `pages/person-detail/{index.ts,component.tsx}` | Create | `PersonDetailPage` |
| `pages/my-schedule/component.tsx` | Modify | The calendar instead of the placeholder |
| `features/people/components/person-row/**` | Modify | `LinkRow` to the person page, name as `TextLink`; test |
| `resources/routes/index.ts`, `resources/rbac/permissions.ts`, `app/router.tsx` | Modify | Route and permission |
| `app/breadcrumbs/**` | Modify | `PersonTrail`: People › name, "People" a link for planners only; `Trail` shows a middle crumb without `to` as plain text; test |
| `shared/i18n/language-key.ts` | Modify | New keys |

Docs: `docs/features/member-calendar.md`, `docs/features/README.md`, `server/README.md` (one API row).

---
## Server

Start: `git checkout main && git pull && git checkout -b feat/member-calendar`.

### Task 1: Schedule service

**Files:**
- Create: `server/internal/models/schedule.go`
- Create: `server/internal/services/schedule.go`
- Create: `server/internal/services/schedule_test.go`
- Modify: `server/internal/config/config.go` (Heatmap section of `Config` and of the constructor)
- Modify: `server/cmd/server/.env.example` (after the Team heatmap block)

**Interfaces:**
- Consumes (existing, `package services`): `capacityInputs(p models.Person, absences []models.Absence) (capacity.Person, []capacity.Absence)` and `capacityPolicy(cfg *config.Config) capacity.Policy` (`absence.go`); `utilizationBand(load, capacity, underusedBelow, nearFullFrom float64) (*float64, string)` (`insights.go`); `findBaseline(ctx, plans, op string, withAssignments bool) (*models.Plan, error)` (`plan.go`); `notFoundOrDatabase(err error, code, message, resource, op string) error` and `resourcePerson` (`person.go`); `hoursEpsilon`, `daysInWeek` (`staffing.go`); `repositories.PlanRepository.ListPersonAssignments(ctx, planID, personID string, from, to isoweek.Week) ([]models.Assignment, error)` (preloads `Demand.Expertise` and `Demand.Project.Program`, allocations in range); `repositories.AbsenceRepository.ListOverlapping(ctx, from, to time.Time) ([]models.Absence, error)`; `AuthzService.RequireSelfOr(ctx, personID string, check Check)`; `CanViewPeople`.
- Produces: `models.PersonSchedule`, `models.ScheduleWeek`, `models.ScheduleProject`, `models.ScheduleDay`, `models.ScheduleAbsence`; aliases `services.PersonSchedule` … `services.ScheduleAbsence`; `services.ScheduleInput`; `services.ScheduleService`; `services.ProvideScheduleService(plans repositories.PlanRepository, people repositories.PersonRepository, absences repositories.AbsenceRepository, authz AuthzService, cfg *config.Config) ScheduleService`; `config.Config.ScheduleMaxWeeks int`.

- [ ] **Step 1: Add the setting**

In `server/internal/config/config.go`, extend the heatmap block of `Config`:

```go
	// Heatmap range and utilization bands (server/README §8). Overloaded is always over 100 %.
	HeatmapDefaultWeeks       int
	HeatmapMaxWeeks           int
	UtilizationUnderusedBelow float64
	UtilizationNearFullFrom   float64
	// ScheduleMaxWeeks is the widest range of one person's schedule (a month grid needs 6).
	ScheduleMaxWeeks int
```

and the constructor, under `UtilizationNearFullFrom`:

```go
		UtilizationNearFullFrom:   getEnvAsFloat("UTILIZATION_NEAR_FULL_FROM", 0.85),
		ScheduleMaxWeeks:          getEnvAsInt("SCHEDULE_MAX_WEEKS", 8),
```

Run `gofmt -w internal/config` so the struct stays aligned.

In `server/cmd/server/.env.example`, after `UTILIZATION_NEAR_FULL_FROM=0.85`:

```dotenv

# Member calendar: the widest range of one person's schedule (a month grid needs at most 6 weeks)
SCHEDULE_MAX_WEEKS=8
```

- [ ] **Step 2: Add the schedule types**

Create `server/internal/models/schedule.go`:

```go
package models

import (
	"time"

	"the-slot-finder/server/internal/isoweek"
)

// PersonSchedule is one person's weeks and days in a range, as the member calendar shows
// them. It is computed from the baseline and availability, never stored.
type PersonSchedule struct {
	// WorkingWeekdays are the days that count (WORKING_WEEKDAYS).
	WorkingWeekdays []time.Weekday
	Weeks           []ScheduleWeek
	// Days holds only the working days with one of the person's absences or a holiday.
	Days []ScheduleDay
}

// ScheduleWeek is the person's load against capacity in one week, computed as on the
// heatmap, with the hours per project behind the load.
type ScheduleWeek struct {
	Week     isoweek.Week
	Capacity float64
	Load     float64
	// Utilization is load ÷ capacity; nil when capacity is 0.
	Utilization *float64
	Band        string
	// Tentative marks planned (not yet approved) absence in the week.
	Tentative bool
	Projects  []ScheduleProject
}

// ScheduleProject is the person's baseline hours on one project and expertise in a week.
type ScheduleProject struct {
	ProgramID   string
	ProgramName string
	ProjectID   string
	ProjectName string
	Expertise   string
	Hours       float64
}

// ScheduleDay is one working day with the person's absences and the holidays on it.
type ScheduleDay struct {
	Date     time.Time
	Absences []ScheduleAbsence
	// Holidays are the ids of the company-wide holidays on the day; holidays have no name.
	Holidays []string
}

// ScheduleAbsence is one of the person's absences on a day.
type ScheduleAbsence struct {
	ID          string
	Type        string
	Status      string
	HoursPerDay float64
}
```

- [ ] **Step 3: Write the failing test**

Create `server/internal/services/schedule_test.go`:

```go
package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	schedPersonID     = "0192f3a4-0000-7000-8000-00000000e001"
	schedOtherID      = "0192f3a4-0000-7000-8000-00000000e002"
	schedHolidayID    = "0192f3a4-0000-7000-8000-00000000e011"
	schedLeaveID      = "0192f3a4-0000-7000-8000-00000000e012"
	schedTrainingID   = "0192f3a4-0000-7000-8000-00000000e013"
	schedOtherLeaveID = "0192f3a4-0000-7000-8000-00000000e014"
	schedProgramID    = "0192f3a4-0000-7000-8000-00000000e021"
	schedCheckoutID   = "0192f3a4-0000-7000-8000-00000000e031"
	schedMigrationID  = "0192f3a4-0000-7000-8000-00000000e032"
	schedMaxWeeks     = 8
)

func octDate(d int) time.Time { return time.Date(2026, time.October, d, 0, 0, 0, 0, time.UTC) }

// scheduleAbsences: Sam's training from before the range into W42, another person's leave,
// a company holiday and Sam's planned leave in W43.
func scheduleAbsences() []models.Absence {
	self, other := schedPersonID, schedOtherID
	return []models.Absence{
		{BaseModel: models.BaseModel{ID: schedTrainingID}, PersonID: &self, FromDate: octDate(9), ToDate: octDate(13),
			HoursPerDay: 4, Type: constants.AbsenceTypeTraining, Status: constants.AbsenceStatusApproved},
		{BaseModel: models.BaseModel{ID: schedOtherLeaveID}, PersonID: &other, FromDate: octDate(20), ToDate: octDate(20),
			HoursPerDay: 8, Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusApproved},
		{BaseModel: models.BaseModel{ID: schedHolidayID}, FromDate: octDate(21), ToDate: octDate(21),
			HoursPerDay: 8, Type: constants.AbsenceTypeHoliday, Status: constants.AbsenceStatusApproved},
		{BaseModel: models.BaseModel{ID: schedLeaveID}, PersonID: &self, FromDate: octDate(22), ToDate: octDate(22),
			HoursPerDay: 8, Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusPlanned},
	}
}

// scheduleAssignments: 40 h in W42 over two projects and two expertises, 16 h in W43.
func scheduleAssignments(t *testing.T) []models.Assignment {
	t.Helper()
	w42, w43 := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")
	program := &models.Program{BaseModel: models.BaseModel{ID: schedProgramID}, Name: "Acme"}
	checkout := &models.Project{BaseModel: models.BaseModel{ID: schedCheckoutID}, ProgramID: schedProgramID, Name: "Checkout", Program: program}
	migration := &models.Project{BaseModel: models.BaseModel{ID: schedMigrationID}, ProgramID: schedProgramID, Name: "Data Migration", Program: program}
	demand := func(project *models.Project, e models.Expertise) *models.Demand {
		return &models.Demand{ProjectID: project.ID, ExpertiseID: e.ID, Expertise: e, Project: project}
	}
	return []models.Assignment{
		{PersonID: schedPersonID, Demand: demand(checkout, engineerExp),
			Allocations: []models.Allocation{{Week: w42, Hours: 20}, {Week: w43, Hours: 16}}},
		{PersonID: schedPersonID, Demand: demand(checkout, qaExp), Allocations: []models.Allocation{{Week: w42, Hours: 8}}},
		{PersonID: schedPersonID, Demand: demand(migration, engineerExp), Allocations: []models.Allocation{{Week: w42, Hours: 12}}},
	}
}

type scheduleRepos struct {
	plans    *MockPlanRepository
	people   *MockPersonRepository
	absences *MockAbsenceRepository
}

func scheduleFixture(t *testing.T, assignments []models.Assignment) scheduleRepos {
	t.Helper()
	from, to := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")
	repos := scheduleRepos{plans: new(MockPlanRepository), people: new(MockPersonRepository), absences: new(MockAbsenceRepository)}
	repos.people.On("FindOneByID", mock.Anything, schedPersonID).Return(&models.Person{BaseModel: models.BaseModel{ID: schedPersonID},
		Name: "Sam", WeeklyHours: 40, NonProjectHours: 4, Active: true}, nil)
	repos.people.On("FindOneByID", mock.Anything, schedOtherID).Return(nil, gorm.ErrRecordNotFound)
	repos.plans.On("FindBaselineSummary", mock.Anything).Return(publishedBaseline(1), nil)
	repos.plans.On("ListPersonAssignments", mock.Anything, baselineID, schedPersonID, from, to).Return(assignments, nil)
	repos.absences.On("ListOverlapping", mock.Anything, from.Monday(), to.Monday().AddDate(0, 0, daysInWeek-1)).
		Return(scheduleAbsences(), nil)
	return repos
}

func newScheduleService(repos scheduleRepos) ScheduleService {
	people := stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: schedPersonID}}}
	cfg := heatmapConfig()
	cfg.ScheduleMaxWeeks = schedMaxWeeks
	return ProvideScheduleService(repos.plans, repos.people, repos.absences, ProvideAuthzService(people), cfg)
}

func TestScheduleService_Get(t *testing.T) {
	from, to := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")
	planner := asRole("kc-planner", constants.RolePlanner)
	sam := ScheduleInput{PersonID: schedPersonID, From: from, To: to}

	t.Run("weeks are computed as on the heatmap, with hours per project", func(t *testing.T) {
		schedule, err := newScheduleService(scheduleFixture(t, scheduleAssignments(t))).Get(planner, sam)
		require.NoError(t, err)
		assert.Equal(t, heatmapConfig().WorkingWeekdays, schedule.WorkingWeekdays)
		require.Len(t, schedule.Weeks, 2)

		w42 := schedule.Weeks[0]
		assert.Equal(t, from, w42.Week)
		assert.InDelta(t, 28, w42.Capacity, hoursEpsilon, "36 h less 4 h of training on Monday and Tuesday")
		assert.InDelta(t, 40, w42.Load, hoursEpsilon)
		assert.Equal(t, constants.BandOverloaded, w42.Band)
		assert.False(t, w42.Tentative)
		require.Len(t, w42.Projects, 3)
		assert.Equal(t, ScheduleProject{ProgramID: schedProgramID, ProgramName: "Acme", ProjectID: schedCheckoutID,
			ProjectName: "Checkout", Expertise: "Engineer", Hours: 20}, w42.Projects[0], "most hours first")
		assert.Equal(t, "Data Migration", w42.Projects[1].ProjectName)
		assert.Equal(t, "QA", w42.Projects[2].Expertise, "QA cover on Checkout is its own line")

		w43 := schedule.Weeks[1]
		assert.InDelta(t, 20, w43.Capacity, hoursEpsilon, "the holiday and the planned leave both count")
		assert.InDelta(t, 16, w43.Load, hoursEpsilon)
		assert.Equal(t, constants.BandHealthy, w43.Band)
		assert.True(t, w43.Tentative, "planned leave is tentative")
	})

	t.Run("days hold the person's absences and the holidays inside the range", func(t *testing.T) {
		schedule, err := newScheduleService(scheduleFixture(t, scheduleAssignments(t))).Get(planner, sam)
		require.NoError(t, err)
		dates := make([]string, 0, len(schedule.Days))
		for _, d := range schedule.Days {
			dates = append(dates, d.Date.Format(constants.DateLayout))
		}
		assert.Equal(t, []string{"2026-10-12", "2026-10-13", "2026-10-21", "2026-10-22"}, dates,
			"training from before the range shows inside it only; another person's leave is left out")
		assert.Equal(t, []ScheduleAbsence{{ID: schedTrainingID, Type: constants.AbsenceTypeTraining,
			Status: constants.AbsenceStatusApproved, HoursPerDay: 4}}, schedule.Days[0].Absences)
		assert.Equal(t, []string{schedHolidayID}, schedule.Days[2].Holidays)
		assert.Empty(t, schedule.Days[2].Absences)
		require.Len(t, schedule.Days[3].Absences, 1)
		assert.Equal(t, constants.AbsenceStatusPlanned, schedule.Days[3].Absences[0].Status)
	})

	t.Run("a person without baseline hours", func(t *testing.T) {
		schedule, err := newScheduleService(scheduleFixture(t, []models.Assignment{})).Get(planner, sam)
		require.NoError(t, err)
		require.Len(t, schedule.Weeks, 2)
		for _, week := range schedule.Weeks {
			assert.Zero(t, week.Load)
			assert.Empty(t, week.Projects)
		}
		assert.Equal(t, constants.BandUnderused, schedule.Weeks[0].Band)
	})

	t.Run("who may read", func(t *testing.T) {
		specialist := asRole("kc-specialist", constants.RoleSpecialist)
		tests := []struct {
			name       string
			ctx        context.Context
			personID   string
			wantStatus int // 0 means success
		}{
			{name: "a lead reads anyone", ctx: asRole("kc-manager", constants.RoleManager), personID: schedPersonID},
			{name: "a specialist reads their own", ctx: specialist, personID: schedPersonID},
			{name: "a specialist cannot read someone else", ctx: specialist, personID: schedOtherID, wantStatus: http.StatusForbidden},
			{name: "an unknown person", ctx: planner, personID: schedOtherID, wantStatus: http.StatusNotFound},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := newScheduleService(scheduleFixture(t, scheduleAssignments(t))).
					Get(tt.ctx, ScheduleInput{PersonID: tt.personID, From: from, To: to})
				if tt.wantStatus == 0 {
					require.NoError(t, err)
					return
				}
				requireStatus(t, err, tt.wantStatus)
			})
		}
	})

	t.Run("range refusals", func(t *testing.T) {
		tooFar := from.AddWeeks(schedMaxWeeks)
		tests := []struct {
			name     string
			in       ScheduleInput
			wantCode string
		}{
			{name: "more weeks than SCHEDULE_MAX_WEEKS", in: ScheduleInput{PersonID: schedPersonID, From: from, To: tooFar},
				wantCode: constants.WeekRangeTooLong},
			{name: "to before from", in: ScheduleInput{PersonID: schedPersonID, From: to, To: from},
				wantCode: constants.WeekRangeInvalid},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := newScheduleService(scheduleFixture(t, scheduleAssignments(t))).Get(planner, tt.in)
				requireAppError(t, err, tt.wantCode, http.StatusBadRequest)
			})
		}
	})
}
```

- [ ] **Step 4: Run the test to see it fail**

Run: `cd server && go test ./internal/services -run TestScheduleService_Get -v`
Expected: build failure, `undefined: ScheduleService` (and `ScheduleInput`, `ProvideScheduleService`, `ScheduleProject`, `ScheduleAbsence`).

- [ ] **Step 5: Write the service**

Create `server/internal/services/schedule.go`:

```go
package services

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"time"

	"the-slot-finder/server/internal/capacity"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
)

const (
	opPersonSchedule = "person_schedule"
	resourceSchedule = "schedule"
)

// The schedule types are models (computed, never stored), so the dtos package can map them.
type (
	// PersonSchedule is one person's weeks and days in a range.
	PersonSchedule = models.PersonSchedule
	// ScheduleWeek is the person's load against capacity in one week.
	ScheduleWeek = models.ScheduleWeek
	// ScheduleProject is the person's hours on one project and expertise in a week.
	ScheduleProject = models.ScheduleProject
	// ScheduleDay is one working day with absences or holidays.
	ScheduleDay = models.ScheduleDay
	// ScheduleAbsence is one of the person's absences on a day.
	ScheduleAbsence = models.ScheduleAbsence
)

// ScheduleInput asks for one person's schedule from one ISO week through another.
type ScheduleInput struct {
	PersonID string
	From, To isoweek.Week
}

// ScheduleService builds one person's calendar from the baseline and their availability.
type ScheduleService interface {
	// Get returns the person's weeks (capacity, load, band, hours per project) and the days
	// with absences or holidays. Planners and leads may read anyone; others only themselves.
	Get(ctx context.Context, in ScheduleInput) (*PersonSchedule, error)
}

type scheduleService struct {
	plans    repositories.PlanRepository
	people   repositories.PersonRepository
	absences repositories.AbsenceRepository
	authz    AuthzService
	cfg      *config.Config
}

// ProvideScheduleService creates the schedule service.
func ProvideScheduleService(plans repositories.PlanRepository, people repositories.PersonRepository,
	absences repositories.AbsenceRepository, authz AuthzService, cfg *config.Config,
) ScheduleService {
	return &scheduleService{plans: plans, people: people, absences: absences, authz: authz, cfg: cfg}
}

func (s *scheduleService) Get(ctx context.Context, in ScheduleInput) (*PersonSchedule, error) {
	if _, err := s.authz.RequireSelfOr(ctx, in.PersonID, CanViewPeople); err != nil {
		return nil, err
	}
	if err := s.checkRange(in.From, in.To); err != nil {
		return nil, err
	}
	person, err := s.people.FindOneByID(ctx, in.PersonID)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.PersonNotFound, constants.MsgPersonNotFound, resourcePerson, opPersonSchedule)
	}
	plan, err := findBaseline(ctx, s.plans, opPersonSchedule, false)
	if err != nil {
		return nil, err
	}
	assignments, err := s.plans.ListPersonAssignments(ctx, plan.ID, in.PersonID, in.From, in.To)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgPlanLoadFailed, err).WithOperation(opPersonSchedule).WithResource(resourceSchedule)
	}
	absences, err := s.absences.ListOverlapping(ctx, in.From.Monday(), in.To.Monday().AddDate(0, 0, daysInWeek-1))
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opPersonSchedule).WithResource(resourceSchedule)
	}

	weeks := isoweek.Range(in.From, in.To)
	return &PersonSchedule{
		WorkingWeekdays: s.cfg.WorkingWeekdays,
		Weeks:           s.scheduleWeeks(*person, weeks, absences, assignments),
		Days:            s.scheduleDays(person.ID, weeks, absences),
	}, nil
}

// checkRange refuses a range that runs backwards or is longer than SCHEDULE_MAX_WEEKS.
func (s *scheduleService) checkRange(from, to isoweek.Week) error {
	if to.Before(from) {
		return errors.NewAppError(constants.WeekRangeInvalid, constants.MsgWeekRangeInvalid, errors.ErrorTypeValidation,
			http.StatusBadRequest).WithOperation(opPersonSchedule).WithResource(resourceSchedule)
	}
	if isoweek.WeeksBetween(from, to)+1 > s.cfg.ScheduleMaxWeeks {
		return errors.NewAppError(constants.WeekRangeTooLong, fmt.Sprintf(constants.MsgWeekRangeTooLongFormat, s.cfg.ScheduleMaxWeeks),
			errors.ErrorTypeValidation, http.StatusBadRequest).WithOperation(opPersonSchedule).WithResource(resourceSchedule)
	}
	return nil
}

// scheduleWeeks is capacity, load and band per week, computed as the heatmap computes them,
// with the hours per project and expertise behind the load.
func (s *scheduleService) scheduleWeeks(p models.Person, weeks []isoweek.Week, absences []models.Absence,
	assignments []models.Assignment,
) []ScheduleWeek {
	person, own := capacityInputs(p, absences)
	policy := capacityPolicy(s.cfg)
	projects := projectHoursByWeek(assignments)
	out := make([]ScheduleWeek, 0, len(weeks))
	for _, week := range weeks {
		row := ScheduleWeek{Week: week, Capacity: capacity.Weekly(person, week, own, policy),
			Tentative: capacity.Tentative(person, week, own, policy), Projects: projects[week]}
		for _, project := range row.Projects {
			row.Load += project.Hours
		}
		row.Utilization, row.Band = utilizationBand(row.Load, row.Capacity, s.cfg.UtilizationUnderusedBelow, s.cfg.UtilizationNearFullFrom)
		out = append(out, row)
	}
	return out
}

// projectHoursByWeek sums the allocations per week, project and expertise, most hours first.
// Like the heatmap, it skips assignments without their demand.
func projectHoursByWeek(assignments []models.Assignment) map[isoweek.Week][]ScheduleProject {
	type key struct {
		week                   isoweek.Week
		projectID, expertiseID string
	}
	sums := map[key]*ScheduleProject{}
	for _, a := range assignments {
		if a.Demand == nil {
			continue
		}
		entry := ScheduleProject{Expertise: a.Demand.Expertise.Name}
		if project := a.Demand.Project; project != nil {
			entry.ProjectID, entry.ProjectName, entry.ProgramID = project.ID, project.Name, project.ProgramID
			if project.Program != nil {
				entry.ProgramName = project.Program.Name
			}
		}
		for _, alloc := range a.Allocations {
			k := key{week: alloc.Week, projectID: entry.ProjectID, expertiseID: a.Demand.ExpertiseID}
			if sums[k] == nil {
				copied := entry
				sums[k] = &copied
			}
			sums[k].Hours += alloc.Hours
		}
	}
	out := map[isoweek.Week][]ScheduleProject{}
	for k, entry := range sums {
		out[k.week] = append(out[k.week], *entry)
	}
	for _, projects := range out {
		sort.Slice(projects, func(i, j int) bool {
			a, b := projects[i], projects[j]
			switch {
			case a.Hours != b.Hours:
				return a.Hours > b.Hours
			case a.ProjectName != b.ProjectName:
				return a.ProjectName < b.ProjectName
			default:
				return a.Expertise < b.Expertise
			}
		})
	}
	return out
}

// scheduleDays lists the working days in the weeks that hold one of the person's absences or
// a company-wide holiday. Other people's absences are left out.
func (s *scheduleService) scheduleDays(personID string, weeks []isoweek.Week, absences []models.Absence) []ScheduleDay {
	var days []ScheduleDay
	for _, week := range weeks {
		for _, date := range week.Days() {
			if !slices.Contains(s.cfg.WorkingWeekdays, date.Weekday()) {
				continue
			}
			day := ScheduleDay{Date: date}
			for _, a := range absences {
				if !coversDay(a, date) {
					continue
				}
				switch {
				case a.PersonID == nil:
					day.Holidays = append(day.Holidays, a.ID)
				case *a.PersonID == personID:
					day.Absences = append(day.Absences, ScheduleAbsence{ID: a.ID, Type: a.Type, Status: a.Status, HoursPerDay: a.HoursPerDay})
				}
			}
			if len(day.Absences) > 0 || len(day.Holidays) > 0 {
				days = append(days, day)
			}
		}
	}
	return days
}

// coversDay is true when the absence runs over the date. Stored dates are whole days.
func coversDay(a models.Absence, date time.Time) bool {
	return !date.Before(calendarDay(a.FromDate)) && !date.After(calendarDay(a.ToDate))
}

// calendarDay drops the time of day, so dates from the database compare with week days.
func calendarDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
```

- [ ] **Step 6: Run the test to see it pass**

Run: `cd server && go test ./internal/services -run TestScheduleService_Get -v`
Expected: `--- PASS: TestScheduleService_Get` with the five subtests passing.

- [ ] **Step 7: Run the server checks**

Run: `cd server && gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`
Expected: `gofmt -l` prints nothing; every package `ok`; lint reports `0 issues.`

- [ ] **Step 8: Commit**

```bash
git add server/internal/models/schedule.go server/internal/services/schedule.go server/internal/services/schedule_test.go \
  server/internal/config/config.go server/cmd/server/.env.example
git commit -m "feat(server): a person's schedule by week and day

Capacity, load and band per week as on the heatmap, hours per project,
and the person's absences and holidays per working day.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: `GET /people/{id}/schedule`

**Files:**
- Create: `server/internal/dtos/schedule.go`
- Create: `server/internal/handlers/schedule.go`
- Create: `server/internal/handlers/schedule_test.go`
- Modify: `server/internal/constants/routes.go`, `server/internal/constants/roles.go`, `server/internal/constants/messages.go`
- Modify: `server/cmd/server/routes/router.go`, `server/cmd/server/routes/router_test.go`, `server/cmd/server/main.go`
- Regenerate: `server/docs/`

**Interfaces:**
- Consumes: Task 1 `services.ScheduleService`, `services.ScheduleInput`, `models.PersonSchedule` (same type as `services.PersonSchedule`); existing `dtos.ParseWeek(value string) (isoweek.Week, error)`, `BaseHandler.pathID`, `h.ValidationErrorResponse`, `errors.ParseValidationErrors`, the `isoweek` validation tag.
- Produces: `constants.PathPersonSchedule`, `constants.PersonScheduleRoles`, `constants.MsgScheduleRetrieved`; `dtos.ScheduleRequest`, `dtos.ScheduleResponse` (and its parts), `dtos.NewScheduleResponse(s models.PersonSchedule) dtos.ScheduleResponse`; `handlers.ScheduleHandler`, `handlers.ProvideScheduleHandler(schedule services.ScheduleService, validate *validator.Validate) *handlers.ScheduleHandler`; `routes.Router(…, scenarioHandler, scheduleHandler *handlers.ScheduleHandler, verifier, cfg)`.

- [ ] **Step 1: Add the constants**

`server/internal/constants/routes.go`, under `PathPersonAbsences`:

```go
	PathPersonAbsences = "/people/:id/absences"
	PathPersonSchedule = "/people/:id/schedule"
```

`server/internal/constants/roles.go`, after `PersonAssignmentRoles`:

```go
	// PersonScheduleRoles may reach a person's schedule; the service lets a specialist read
	// only their own.
	PersonScheduleRoles = []string{RolePlanner, RoleManager, RoleSpecialist}
```

`server/internal/constants/messages.go`, in the "People and skills messages" block after `MsgExpertisesListed`:

```go
	MsgScheduleRetrieved     = "Schedule retrieved"
```

(Run `gofmt -w internal/constants` afterwards so the block stays aligned.)

- [ ] **Step 2: Write the failing handler test**

Create `server/internal/handlers/schedule_test.go`:

```go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSchedule records the input and returns one week without projects and one holiday.
type fakeSchedule struct {
	got *services.ScheduleInput
	err error
}

func (f *fakeSchedule) Get(_ context.Context, in services.ScheduleInput) (*services.PersonSchedule, error) {
	f.got = &in
	if f.err != nil {
		return nil, f.err
	}
	return &services.PersonSchedule{
		WorkingWeekdays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Sunday},
		Weeks:           []services.ScheduleWeek{{Week: in.From, Capacity: 36, Band: constants.BandUnderused}},
		Days: []services.ScheduleDay{{Date: time.Date(2026, time.October, 14, 0, 0, 0, 0, time.UTC),
			Holidays: []string{"0192f3a4-0000-7000-8000-0000000000f1"}}},
	}, nil
}

func TestScheduleHandler_Get(t *testing.T) {
	path := "/people/" + testPersonID + "/schedule"

	t.Run("maps the schedule with ISO weekdays and empty lists", func(t *testing.T) {
		schedule := &fakeSchedule{}
		h := ProvideScheduleHandler(schedule, newTestValidator(t))
		rec, envelope := serve(t, constants.PathPersonSchedule, http.MethodGet, path+"?from=2026-W42&to=2026-W43", "", h.Get)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		require.NotNil(t, schedule.got)
		assert.Equal(t, testPersonID, schedule.got.PersonID)
		assert.Equal(t, "2026-W42", schedule.got.From.String())
		assert.Equal(t, "2026-W43", schedule.got.To.String())

		var body dtos.ScheduleResponse
		require.NoError(t, json.Unmarshal(envelope.Data, &body))
		assert.Equal(t, []int{1, 2, 3, 4, 5, 7}, body.WorkingWeekdays, "Sunday is 7 in ISO 8601")
		require.Len(t, body.Weeks, 1)
		assert.Equal(t, "2026-W42", body.Weeks[0].Week)
		require.Len(t, body.Days, 1)
		assert.Equal(t, "2026-10-14", body.Days[0].Date)
		assert.Equal(t, "0192f3a4-0000-7000-8000-0000000000f1", body.Days[0].Holidays[0].ID)
		raw := string(envelope.Data)
		assert.Contains(t, raw, `"projects":[]`, "an empty list is [], never null")
		assert.Contains(t, raw, `"absences":[]`)
	})

	t.Run("refusals", func(t *testing.T) {
		tests := []struct {
			name       string
			target     string
			serviceErr error
			wantStatus int
			wantField  string
		}{
			{name: "missing from", target: path + "?to=2026-W43", wantStatus: http.StatusBadRequest, wantField: "from"},
			{name: "not an ISO week", target: path + "?from=2026-W42&to=2026-43", wantStatus: http.StatusBadRequest, wantField: "to"},
			{name: "not a uuid", target: "/people/abc/schedule?from=2026-W42&to=2026-W43", wantStatus: http.StatusBadRequest},
			{name: "someone else's schedule", target: path + "?from=2026-W42&to=2026-W43", wantStatus: http.StatusForbidden,
				serviceErr: errors.ForbiddenError(constants.MsgInsufficientPermissions, nil)},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h := ProvideScheduleHandler(&fakeSchedule{err: tt.serviceErr}, newTestValidator(t))
				rec, envelope := serve(t, constants.PathPersonSchedule, http.MethodGet, tt.target, "", h.Get)
				require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
				if tt.wantField != "" {
					var fields map[string]string
					require.NoError(t, json.Unmarshal(envelope.Data, &fields))
					assert.Contains(t, fields, validationKeyPrefix+tt.wantField)
				}
			})
		}
	})
}
```

- [ ] **Step 3: Run the test to see it fail**

Run: `cd server && go test ./internal/handlers -run TestScheduleHandler_Get -v`
Expected: build failure, `undefined: ProvideScheduleHandler` and `undefined: dtos.ScheduleResponse`.

- [ ] **Step 4: Write the DTOs**

Create `server/internal/dtos/schedule.go`:

```go
package dtos

import (
	"strings"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"
)

// isoSunday is Sunday in ISO 8601, where Monday is 1; time.Weekday counts Sunday as 0.
const isoSunday = 7

// ScheduleRequest is the query of GET /people/{id}/schedule.
type ScheduleRequest struct {
	From string `json:"from" query:"from" validate:"required,isoweek"`
	To   string `json:"to" query:"to" validate:"required,isoweek"`
}

// Normalize trims the weeks.
func (r *ScheduleRequest) Normalize() {
	r.From = strings.TrimSpace(r.From)
	r.To = strings.TrimSpace(r.To)
}

// ScheduleProjectResponse is the person's hours on one project and expertise in a week.
type ScheduleProjectResponse struct {
	ProgramID   string  `json:"program_id"`
	ProgramName string  `json:"program_name" example:"Acme Commerce Platform"`
	ProjectID   string  `json:"project_id"`
	ProjectName string  `json:"project_name" example:"Checkout Revamp"`
	Expertise   string  `json:"expertise" example:"Engineer"`
	Hours       float64 `json:"hours" example:"20"`
}

// ScheduleWeekResponse is the person's load against capacity in one week.
type ScheduleWeekResponse struct {
	Week     string  `json:"week" example:"2026-W42"`
	Capacity float64 `json:"capacity" example:"36"`
	Load     float64 `json:"load" example:"32"`
	// Utilization is load ÷ capacity; null when capacity is 0.
	Utilization *float64 `json:"utilization" example:"0.89"`
	Band        string   `json:"band" enums:"unavailable,underused,healthy,near_full,overloaded"`
	// Tentative marks planned (not yet approved) absence in the week.
	Tentative bool                      `json:"tentative"`
	Projects  []ScheduleProjectResponse `json:"projects"`
}

// ScheduleAbsenceResponse is one of the person's absences on a day.
type ScheduleAbsenceResponse struct {
	ID          string  `json:"id"`
	Type        string  `json:"type" enums:"leave,training"`
	Status      string  `json:"status" enums:"planned,approved"`
	HoursPerDay float64 `json:"hours_per_day" example:"8"`
}

// ScheduleHolidayResponse is a company-wide holiday on a day.
type ScheduleHolidayResponse struct {
	ID string `json:"id"`
}

// ScheduleDayResponse is a working day with the person's absences and the holidays on it.
type ScheduleDayResponse struct {
	Date     string                    `json:"date" example:"2026-10-21"`
	Absences []ScheduleAbsenceResponse `json:"absences"`
	Holidays []ScheduleHolidayResponse `json:"holidays"`
}

// ScheduleResponse is one person's calendar for a range of ISO weeks.
type ScheduleResponse struct {
	// WorkingWeekdays are ISO day numbers: 1 is Monday, 7 is Sunday.
	WorkingWeekdays []int                  `json:"working_weekdays" example:"1,2,3,4,5"`
	Weeks           []ScheduleWeekResponse `json:"weeks"`
	// Days holds only the working days with an absence or a holiday.
	Days []ScheduleDayResponse `json:"days"`
}

// NewScheduleResponse maps a schedule; lists are never nil.
func NewScheduleResponse(s models.PersonSchedule) ScheduleResponse {
	weekdays := make([]int, 0, len(s.WorkingWeekdays))
	for _, d := range s.WorkingWeekdays {
		weekdays = append(weekdays, isoWeekday(d))
	}
	weeks := make([]ScheduleWeekResponse, 0, len(s.Weeks))
	for _, w := range s.Weeks {
		projects := make([]ScheduleProjectResponse, 0, len(w.Projects))
		for _, p := range w.Projects {
			projects = append(projects, ScheduleProjectResponse{ProgramID: p.ProgramID, ProgramName: p.ProgramName,
				ProjectID: p.ProjectID, ProjectName: p.ProjectName, Expertise: p.Expertise, Hours: p.Hours})
		}
		weeks = append(weeks, ScheduleWeekResponse{Week: w.Week.String(), Capacity: w.Capacity, Load: w.Load,
			Utilization: w.Utilization, Band: w.Band, Tentative: w.Tentative, Projects: projects})
	}
	days := make([]ScheduleDayResponse, 0, len(s.Days))
	for _, d := range s.Days {
		absences := make([]ScheduleAbsenceResponse, 0, len(d.Absences))
		for _, a := range d.Absences {
			absences = append(absences, ScheduleAbsenceResponse{ID: a.ID, Type: a.Type, Status: a.Status, HoursPerDay: a.HoursPerDay})
		}
		holidays := make([]ScheduleHolidayResponse, 0, len(d.Holidays))
		for _, id := range d.Holidays {
			holidays = append(holidays, ScheduleHolidayResponse{ID: id})
		}
		days = append(days, ScheduleDayResponse{Date: d.Date.Format(constants.DateLayout), Absences: absences, Holidays: holidays})
	}
	return ScheduleResponse{WorkingWeekdays: weekdays, Weeks: weeks, Days: days}
}

// isoWeekday is the ISO 8601 number of a weekday.
func isoWeekday(d time.Weekday) int {
	if d == time.Sunday {
		return isoSunday
	}
	return int(d)
}
```

- [ ] **Step 5: Write the handler**

Create `server/internal/handlers/schedule.go`:

```go
package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// ScheduleHandler serves one person's calendar.
type ScheduleHandler struct {
	BaseHandler
	schedule services.ScheduleService
	validate *validator.Validate
}

// ProvideScheduleHandler creates the schedule handler.
func ProvideScheduleHandler(schedule services.ScheduleService, validate *validator.Validate) *ScheduleHandler {
	return &ScheduleHandler{BaseHandler: *NewBaseHandler(), schedule: schedule, validate: validate}
}

// Get godoc
// @Summary A person's schedule
// @Description Per ISO week: capacity, load, utilization band, tentative flag and hours per project from the baseline, computed as on the heatmap. Per working day: the person's absences and the company-wide holidays. A range over SCHEDULE_MAX_WEEKS gets WEEK_RANGE_TOO_LONG. Planners and leads may read anyone; others only themselves.
// @Tags People
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Param from query string true "First week" example(2026-W40)
// @Param to query string true "Last week" example(2026-W44)
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ScheduleResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /people/{id}/schedule [get]
func (h *ScheduleHandler) Get(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.ScheduleRequest
	if err := echo.BindQueryParams(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	from, err := dtos.ParseWeek(req.From)
	if err != nil {
		return h.HandleError(c, err)
	}
	to, err := dtos.ParseWeek(req.To)
	if err != nil {
		return h.HandleError(c, err)
	}
	schedule, err := h.schedule.Get(c.Request().Context(), services.ScheduleInput{PersonID: id, From: from, To: to})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgScheduleRetrieved, dtos.NewScheduleResponse(*schedule), nil)
}
```

- [ ] **Step 6: Run the handler test to see it pass**

Run: `cd server && go test ./internal/handlers -run TestScheduleHandler_Get -v`
Expected: `--- PASS: TestScheduleHandler_Get` (2 subtests, 4 refusal cases).

- [ ] **Step 7: Add the route and its RBAC test**

In `server/cmd/server/routes/router_test.go`, `newTestRouter`: create the handler and pass it after `scenarios`:

```go
	scenarios := handlers.ProvideScenarioHandler(nil, nil)
	schedule := handlers.ProvideScheduleHandler(nil, nil)
	return Router(health, identity, person, absence, program, project, plan, staffing, insights, events, scenarios, schedule, verifier, cfg)
```

and add to the table in `TestRouter_People_RBAC`:

```go
		{name: "no role cannot read a schedule", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/schedule", token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "no token cannot read a schedule", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/schedule", wantStatus: http.StatusUnauthorized},
```

Run: `cd server && go test ./cmd/server/routes -run TestRouter_People_RBAC -v`
Expected: build failure, `too many arguments in call to Router`.

In `server/cmd/server/routes/router.go`, add the parameter after `scenarioHandler`:

```go
	scenarioHandler *handlers.ScenarioHandler,
	scheduleHandler *handlers.ScheduleHandler,
	verifier auth.TokenVerifier,
```

and register the route right after `registerPeopleRoutes(...)`:

```go
	registerPeopleRoutes(v1, authenticated, personHandler, absenceHandler)
	// A specialist reads only their own schedule (checked in the service).
	v1.GET(constants.PathPersonSchedule, scheduleHandler.Get, authenticated, middlewares.RequireRole(constants.PersonScheduleRoles...))
```

In `server/cmd/server/main.go`: add `scheduleHandler *handlers.ScheduleHandler,` to `NewHTTPServer` after `scenarioHandler`, pass it to `routes.Router(…, scenarioHandler, scheduleHandler, verifier, cfg)`, and add the providers to `appOptions`:

```go
			services.ProvideScenarioService,
			services.ProvideScheduleService,
```

```go
			handlers.ProvideScenarioHandler,
			handlers.ProvideScheduleHandler,
```

Run: `cd server && go test ./cmd/server/... -v -run 'TestRouter_People_RBAC|TestAppOptions_Validate'`
Expected: both PASS (the FX graph test proves the new providers resolve).

- [ ] **Step 8: Regenerate Swagger**

Run: `cd server && make swagger-load` (or the `go run …swag@v1.16.6 init …` command from Global Constraints)
Expected: `docs/swagger.yaml`, `docs/swagger.json` and `docs/docs.go` change; `grep -c "/people/{id}/schedule" docs/swagger.yaml` prints `1`.

- [ ] **Step 9: Run the server checks**

Run: `cd server && gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`
Expected: no gofmt output, all packages `ok`, `0 issues.`

- [ ] **Step 10: Commit**

```bash
git add server/internal/dtos/schedule.go server/internal/handlers/schedule.go server/internal/handlers/schedule_test.go \
  server/internal/constants server/cmd/server server/docs
git commit -m "feat(server): GET /people/{id}/schedule

Planners and leads read anyone's schedule, others only their own.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: `GET /people/{id}` for the person page header

The spec's person page header needs the name, primary expertise, weekly hours and non-project hours. No route returns one person today (`PathPerson` only has `PATCH`), and the people list is paged.

**Files:**
- Modify: `server/internal/handlers/person.go`, `server/internal/handlers/person_test.go`
- Modify: `server/internal/constants/messages.go`
- Modify: `server/cmd/server/routes/router.go`, `server/cmd/server/routes/router_test.go`
- Regenerate: `server/docs/`

**Interfaces:**
- Consumes: existing `services.PeopleService.Get(ctx, id string) (*models.Person, error)` (404 `PERSON_NOT_FOUND` when missing), `services.CanViewPeople`, `dtos.NewPersonResponse`.
- Produces: `PersonHandler.Get`; `constants.MsgPersonRetrieved`; route `GET /api/v1/people/{id}` for `PeopleViewRoles` (planner, lead).

- [ ] **Step 1: Write the failing test**

In `server/internal/handlers/person_test.go`, add to `fakePeople`:

```go
func (f *fakePeople) Get(_ context.Context, id string) (*models.Person, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &models.Person{BaseModel: models.BaseModel{ID: id}, Name: "Alex", WeeklyHours: 40, NonProjectHours: 4}, nil
}
```

and the test:

```go
func TestPersonHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serviceErr error
		wantStatus int
	}{
		{name: "found", target: "/people/" + testPersonID, wantStatus: http.StatusOK},
		{name: "not a uuid", target: "/people/abc", wantStatus: http.StatusBadRequest},
		{name: "unknown person", target: "/people/" + testPersonID, wantStatus: http.StatusNotFound,
			serviceErr: errors.NewAppError(constants.PersonNotFound, constants.MsgPersonNotFound, errors.ErrorTypeNotFound, http.StatusNotFound)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ProvidePersonHandler(&fakePeople{err: tt.serviceErr}, services.ProvideAuthzService(nil), newTestValidator(t))
			rec, envelope := serve(t, constants.PathPerson, http.MethodGet, tt.target, "", h.Get)
			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantStatus == http.StatusOK {
				var person dtos.PersonResponse
				require.NoError(t, json.Unmarshal(envelope.Data, &person))
				assert.Equal(t, "Alex", person.Name)
				assert.NotNil(t, person.Skills)
			}
		})
	}
}
```

In `router_test.go`, `TestRouter_People_RBAC`:

```go
		{name: "specialist cannot read a person", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa", token: "specialist-token", wantStatus: http.StatusForbidden},
```

- [ ] **Step 2: Run the tests to see them fail**

Run: `cd server && go test ./internal/handlers -run TestPersonHandler_Get -v`
Expected: build failure, `h.Get undefined (type *PersonHandler has no field or method Get)`.

- [ ] **Step 3: Implement**

`server/internal/constants/messages.go`, people block after `MsgPeopleListed`:

```go
	MsgPersonRetrieved       = "Person retrieved"
```

`server/internal/handlers/person.go`, after `List`:

```go
// Get godoc
// @Summary A person
// @Description One person with their skills (primary first), for the person page.
// @Tags People
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.PersonResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /people/{id} [get]
func (h *PersonHandler) Get(c *echo.Context) error {
	ctx := c.Request().Context()
	if _, err := h.authz.Require(ctx, services.CanViewPeople); err != nil {
		return h.HandleError(c, err)
	}
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	person, err := h.people.Get(ctx, id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgPersonRetrieved, dtos.NewPersonResponse(*person), nil)
}
```

`server/cmd/server/routes/router.go`, in `registerPeopleRoutes` after the `PathPeople` GET:

```go
	v1.GET(constants.PathPerson, personHandler.Get, authenticated, middlewares.RequireRole(constants.PeopleViewRoles...))
```

- [ ] **Step 4: Run the tests to see them pass**

Run: `cd server && go test ./internal/handlers -run TestPersonHandler_Get -v && go test ./cmd/server/routes -run TestRouter_People_RBAC -v`
Expected: both PASS.

- [ ] **Step 5: Regenerate Swagger and run the server checks**

Run: `cd server && make swagger-load && gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`
Expected: Swagger gains `get` under `/people/{id}`; no gofmt output; all `ok`; `0 issues.`

- [ ] **Step 6: Commit**

```bash
git add server/internal/handlers/person.go server/internal/handlers/person_test.go server/internal/constants/messages.go \
  server/cmd/server/routes server/docs
git commit -m "feat(server): GET /people/{id} for planners and leads

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 4: `POST /absences/{id}/approve`

Until now a status could only be chosen when an absence was created. Calendar-sync brings leave in as planned, so a planner needs to approve an existing absence.

**Files:**
- Modify: `server/internal/services/absence.go`, `server/internal/services/absence_test.go`
- Modify: `server/internal/handlers/absence.go`
- Create: `server/internal/handlers/absence_test.go`
- Modify: `server/internal/constants/routes.go`, `server/internal/constants/error_codes.go`, `server/internal/constants/messages.go`
- Modify: `server/cmd/server/routes/router.go`, `server/cmd/server/routes/router_test.go`
- Regenerate: `server/docs/`

**Interfaces:**
- Consumes (existing, `package services`): `availabilityService` with `repo repositories.AbsenceRepository` (its `FindOneByID` and `Updates` come from `AbstractRepository`), `people PeopleService`, `authz AuthzService`; `saveWithEvent(ctx, write func(ctx context.Context) error, event ChangeInput) error`; `absenceChange(a *models.Absence, name string) AbsenceChange`; `asAppError(err error, message, op, resource string) error`; `CanManageAbsences`; `resourceAbsence`. Tests: `MockAbsenceRepository`, `stubPeople`, `recordedEvents`, `passthroughTx`, `asRole`, `requireStatus`, `requireAppError`, `absenceID`, `selfPersonID`.
- Produces: `AvailabilityService.Approve(ctx context.Context, id string) (*models.Absence, error)`; `AbsenceHandler.Approve`; `constants.PathAbsenceApprove`, `constants.AbsenceNotPlanned`, `constants.MsgAbsenceApproved`, `constants.MsgAbsenceNotPlanned`, `constants.MsgAbsenceApproveFailed`; in `router_test.go`, `newTestRouterWith(cfg *config.Config, availability services.AvailabilityService) *echo.Echo`.

- [ ] **Step 1: Add the constants**

`server/internal/constants/routes.go`, under `PathAbsence`:

```go
	PathAbsence        = "/absences/:id"
	PathAbsenceApprove = "/absences/:id/approve"
```

`server/internal/constants/error_codes.go`, in the absence block:

```go
	// Absence errors
	AbsenceNotFound     = "ABSENCE_NOT_FOUND"
	AbsenceDatesInvalid = "ABSENCE_DATES_INVALID"
	// AbsenceNotPlanned: only a person's planned absence can be approved (not an approved one, not a holiday).
	AbsenceNotPlanned = "ABSENCE_NOT_PLANNED"
```

`server/internal/constants/messages.go`, in the "Absence and holiday messages" block:

```go
	MsgAbsenceApproved      = "Absence approved"
	MsgAbsenceNotPlanned    = "Only a planned absence can be approved"
	MsgAbsenceApproveFailed = "Could not approve the absence"
```

Run `gofmt -w internal/constants`.

- [ ] **Step 2: Write the failing service test**

Append to `server/internal/services/absence_test.go`:

```go
func TestAvailabilityService_Approve(t *testing.T) {
	self := selfPersonID
	planned := func() *models.Absence {
		return &models.Absence{BaseModel: models.BaseModel{ID: absenceID}, PersonID: &self, HoursPerDay: 8,
			FromDate: time.Date(2026, time.October, 22, 0, 0, 0, 0, time.UTC), ToDate: time.Date(2026, time.October, 23, 0, 0, 0, 0, time.UTC),
			Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusPlanned}
	}
	approved := planned()
	approved.Status = constants.AbsenceStatusApproved
	// A holiday is never a person's; keep it planned so only the holiday rule can refuse it.
	holiday := &models.Absence{BaseModel: models.BaseModel{ID: absenceID}, Type: constants.AbsenceTypeHoliday,
		Status: constants.AbsenceStatusPlanned}
	planner := asRole("kc-planner", constants.RolePlanner)

	tests := []struct {
		name       string
		ctx        context.Context
		absence    *models.Absence
		findErr    error
		wantStatus int // 0 means approved
		wantCode   string
	}{
		{name: "a planner approves planned leave", ctx: planner, absence: planned()},
		{name: "already approved", ctx: planner, absence: approved, wantStatus: http.StatusConflict, wantCode: constants.AbsenceNotPlanned},
		{name: "a holiday", ctx: planner, absence: holiday, wantStatus: http.StatusConflict, wantCode: constants.AbsenceNotPlanned},
		{name: "not found", ctx: planner, findErr: gorm.ErrRecordNotFound, wantStatus: http.StatusNotFound, wantCode: constants.AbsenceNotFound},
		{name: "a manager cannot approve", ctx: asRole("kc-manager", constants.RoleManager), absence: planned(), wantStatus: http.StatusForbidden},
		{name: "a specialist cannot approve their own", ctx: asRole("kc-specialist", constants.RoleSpecialist), absence: planned(),
			wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAbsenceRepository)
			if tt.findErr != nil {
				repo.On("FindOneByID", mock.Anything, absenceID).Return(nil, tt.findErr)
			} else {
				repo.On("FindOneByID", mock.Anything, absenceID).Return(tt.absence, nil)
			}
			repo.On("Updates", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
				return a.Status == constants.AbsenceStatusApproved
			})).Return(nil)
			events := &recordedEvents{}
			people := stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: selfPersonID}}}
			service := ProvideAvailabilityService(repo, people, ProvideAuthzService(people), passthroughTx{}, events)

			absence, err := service.Approve(tt.ctx, absenceID)
			if tt.wantStatus != 0 {
				if tt.wantCode != "" {
					requireAppError(t, err, tt.wantCode, tt.wantStatus)
				} else {
					requireStatus(t, err, tt.wantStatus)
				}
				repo.AssertNotCalled(t, "Updates", mock.Anything, mock.Anything)
				assert.Empty(t, events.events, "a refused approve records no change event")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, constants.AbsenceStatusApproved, absence.Status)
			require.Len(t, events.events, 1)
			assert.Equal(t, constants.EventAbsenceAdded, events.events[0].Type)
			change, ok := events.events[0].Payload.(AbsenceChange)
			require.True(t, ok)
			assert.Equal(t, constants.AbsenceStatusApproved, change.Status, "the event carries the new status")
			assert.Equal(t, &self, change.PersonID)
		})
	}
}
```

- [ ] **Step 3: Run it to see it fail**

Run: `cd server && go test ./internal/services -run TestAvailabilityService_Approve -v`
Expected: build failure, `service.Approve undefined (type AvailabilityService has no field or method Approve)`.

- [ ] **Step 4: Implement the service method**

In `server/internal/services/absence.go`:
- add `opApproveAbsence = "approve_absence"` to the `const` block
- add to the `AvailabilityService` interface, after `Delete`:

```go
	// Approve turns a person's planned absence into an approved one and records the change.
	// Only planners may; an approved absence or a holiday gets 409.
	Approve(ctx context.Context, absenceID string) (*models.Absence, error)
```

- add the method after `Delete`:

```go
func (s *availabilityService) Approve(ctx context.Context, absenceID string) (*models.Absence, error) {
	if _, err := s.authz.Require(ctx, CanManageAbsences); err != nil {
		return nil, err
	}
	absence, err := s.repo.FindOneByID(ctx, absenceID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WrapError(err, constants.AbsenceNotFound, constants.MsgAbsenceNotFound,
				errors.ErrorTypeNotFound, http.StatusNotFound).WithOperation(opApproveAbsence).WithResource(resourceAbsence)
		}
		return nil, errors.DatabaseError(constants.MsgAbsenceApproveFailed, err).
			WithOperation(opApproveAbsence).WithResource(resourceAbsence)
	}
	if absence.PersonID == nil || absence.Status != constants.AbsenceStatusPlanned {
		return nil, errors.NewAppError(constants.AbsenceNotPlanned, constants.MsgAbsenceNotPlanned, errors.ErrorTypeConflict,
			http.StatusConflict).WithOperation(opApproveAbsence).WithResource(resourceAbsence)
	}

	name := ""
	if person, err := s.people.Get(ctx, *absence.PersonID); err == nil {
		name = person.Name
	}
	absence.Status = constants.AbsenceStatusApproved
	update := func(ctx context.Context) error { return s.repo.Updates(ctx, absence) }
	// The constants have only AbsenceAdded and AbsenceRemoved; an approval is the absence
	// again with its new status, so the re-plan runner sees it like any absence change.
	event := ChangeInput{Type: constants.EventAbsenceAdded, Payload: absenceChange(absence, name)}
	if err := s.saveWithEvent(ctx, update, event); err != nil {
		return nil, asAppError(err, constants.MsgAbsenceApproveFailed, opApproveAbsence, resourceAbsence)
	}
	return absence, nil
}
```

- [ ] **Step 5: Run the service test to see it pass**

Run: `cd server && go test ./internal/services -run TestAvailabilityService_Approve -v`
Expected: PASS, 6 subtests.

- [ ] **Step 6: Write the failing handler and router tests**

Create `server/internal/handlers/absence_test.go`:

```go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAbsenceID = "0192f3a4-0000-7000-8000-0000000000c1"

// fakeAvailability implements only Approve; other methods panic if called.
type fakeAvailability struct {
	services.AvailabilityService
	approved string
	err      error
}

func (f *fakeAvailability) Approve(_ context.Context, id string) (*models.Absence, error) {
	f.approved = id
	if f.err != nil {
		return nil, f.err
	}
	person := testPersonID
	day := time.Date(2026, time.October, 22, 0, 0, 0, 0, time.UTC)
	return &models.Absence{BaseModel: models.BaseModel{ID: id}, PersonID: &person, FromDate: day, ToDate: day, HoursPerDay: 8,
		Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusApproved}, nil
}

func TestAbsenceHandler_Approve(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		serviceErr error
		wantStatus int
	}{
		{name: "approved", target: "/absences/" + testAbsenceID + "/approve", wantStatus: http.StatusOK},
		{name: "not a uuid", target: "/absences/abc/approve", wantStatus: http.StatusBadRequest},
		{name: "already approved", target: "/absences/" + testAbsenceID + "/approve", wantStatus: http.StatusConflict,
			serviceErr: errors.NewAppError(constants.AbsenceNotPlanned, constants.MsgAbsenceNotPlanned, errors.ErrorTypeConflict, http.StatusConflict)},
		{name: "unknown absence", target: "/absences/" + testAbsenceID + "/approve", wantStatus: http.StatusNotFound,
			serviceErr: errors.NewAppError(constants.AbsenceNotFound, constants.MsgAbsenceNotFound, errors.ErrorTypeNotFound, http.StatusNotFound)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			availability := &fakeAvailability{err: tt.serviceErr}
			h := ProvideAbsenceHandler(availability, newTestValidator(t))
			rec, envelope := serve(t, constants.PathAbsenceApprove, http.MethodPost, tt.target, "", h.Approve)
			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, testAbsenceID, availability.approved)
				var absence dtos.AbsenceResponse
				require.NoError(t, json.Unmarshal(envelope.Data, &absence))
				assert.Equal(t, constants.AbsenceStatusApproved, absence.Status)
			}
			if tt.wantStatus == http.StatusConflict {
				assert.Equal(t, constants.AbsenceNotPlanned, envelope.Meta.ErrorCode)
			}
		})
	}
}
```


In `server/cmd/server/routes/router_test.go`, split the router helper so a test can put a service behind the absence routes. Replace `newTestRouter` with:

```go
func newTestRouter(cfg *config.Config) *echo.Echo {
	return newTestRouterWith(cfg, nil)
}

// newTestRouterWith is newTestRouter with an availability service behind the absence routes.
func newTestRouterWith(cfg *config.Config, availability services.AvailabilityService) *echo.Echo {
	verifier := fakeVerifier{
		"planner-token":    userClaims("planner", constants.RolePlanner, "offline_access"),
		"manager-token":       userClaims("manager", constants.RoleManager),
		"specialist-token": userClaims("specialist", constants.RoleSpecialist),
		"no-role-token":    userClaims("nobody", "offline_access"),
	}
	authz := services.ProvideAuthzService(nil)
	identity := handlers.ProvideIdentityHandler(authz, plannerOnlyPeople{})
	health := handlers.ProvideHealthHandler(cfg, nil)
	person := handlers.ProvidePersonHandler(nil, authz, nil)
	absence := handlers.ProvideAbsenceHandler(availability, nil)
	program := handlers.ProvideProgramHandler(nil, nil)
	project := handlers.ProvideProjectHandler(nil, nil, nil, nil)
	plan := handlers.ProvidePlanHandler(nil, nil)
	staffing := handlers.ProvideStaffingHandler(nil, nil)
	insights := handlers.ProvideInsightsHandler(nil, nil)
	events := handlers.ProvideEventHandler(nil, nil)
	scenarios := handlers.ProvideScenarioHandler(nil, nil)
	schedule := handlers.ProvideScheduleHandler(nil, nil)
	return Router(health, identity, person, absence, program, project, plan, staffing, insights, events, scenarios, schedule, verifier, cfg)
}
```

and add:

```go
// approvingAvailability approves any absence, so a request that passes the role check gets 200.
type approvingAvailability struct {
	services.AvailabilityService
}

func (approvingAvailability) Approve(_ context.Context, id string) (*models.Absence, error) {
	return &models.Absence{BaseModel: models.BaseModel{ID: id}, Status: constants.AbsenceStatusApproved}, nil
}

func TestRouter_ApproveAbsence_RBAC(t *testing.T) {
	e := newTestRouterWith(baseConfig(), approvingAvailability{})
	path := constants.PathAPIBase + "/absences/0192f3a4-0000-7000-8000-0000000000aa/approve"
	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{name: "a planner approves", token: "planner-token", wantStatus: http.StatusOK},
		{name: "a manager cannot approve", token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "a specialist cannot approve", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no token", wantStatus: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			req.RemoteAddr = "192.0.2.10:5000"
			// CSRF runs before auth; a same-origin POST passes it.
			req.Header.Set("Sec-Fetch-Site", "same-origin")
			if tt.token != "" {
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
		})
	}
}
```

Run: `cd server && go test ./internal/handlers -run TestAbsenceHandler_Approve -v`
Expected: build failure, `h.Approve undefined`.

- [ ] **Step 7: Implement the handler and the route**

In `server/internal/handlers/absence.go`, after `Delete`:

```go
// Approve godoc
// @Summary Approve a planned absence
// @Description Turns a person's planned absence into an approved one and records an AbsenceAdded change event with the new status. An approved absence or a holiday gets ABSENCE_NOT_PLANNED.
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Param id path string true "Absence id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.AbsenceResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /absences/{id}/approve [post]
func (h *AbsenceHandler) Approve(c *echo.Context) error {
	absenceID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	absence, err := h.availability.Approve(c.Request().Context(), absenceID)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgAbsenceApproved, dtos.NewAbsenceResponse(*absence), nil)
}
```

In `server/cmd/server/routes/router.go`, `registerPeopleRoutes`, after the `PathAbsence` DELETE:

```go
	v1.POST(constants.PathAbsenceApprove, absenceHandler.Approve, authenticated, middlewares.RequireRole(constants.AbsenceManageRoles...))
```

- [ ] **Step 8: Run the tests to see them pass**

Run: `cd server && go test ./internal/handlers -run TestAbsenceHandler_Approve -v && go test ./cmd/server/routes -v`
Expected: PASS, including `TestRouter_ApproveAbsence_RBAC` (planner 200, lead 403, specialist 403, no token 401) and the earlier RBAC tests.

- [ ] **Step 9: Regenerate Swagger and run the server checks**

Run: `cd server && make swagger-load && gofmt -l . && go build ./... && go vet ./... && go test ./... && make lint`
Expected: Swagger gains `post` under `/absences/{id}/approve`; no gofmt output; all `ok`; `0 issues.`

- [ ] **Step 10: Commit**

```bash
git add server/internal/services/absence.go server/internal/services/absence_test.go server/internal/handlers/absence.go \
  server/internal/handlers/absence_test.go server/internal/constants server/cmd/server/routes server/docs
git commit -m "feat(server): POST /absences/{id}/approve

Planners approve a person's planned absence; it records an AbsenceAdded
change event with the new status. Approving twice or a holiday gets 409.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Client

### Task 5: Move the utilization bands to `shared/utilization`

The calendar colours a week by its band like the heatmap does. Features never import each other, so the band type and its classes and labels move to a shared module. This is a refactor: the existing tests must stay green without changes.

**Files:**
- Create: `client/src/shared/utilization/types.ts`, `client/src/shared/utilization/constants.ts`, `client/src/shared/utilization/index.ts`
- Modify: `client/src/features/heatmap/types.ts`, `client/src/features/heatmap/constants.ts`

**Interfaces:**
- Produces: `@/shared/utilization` exports `type UtilizationBand`, `BAND_CLASSES: Record<UtilizationBand, string>`, `BAND_LABELS: Record<UtilizationBand, PlainKey>`, `BAND_ORDER: readonly UtilizationBand[]`, `OVERLOADED_BAND: UtilizationBand`.

- [ ] **Step 1: Run the heatmap tests as the baseline**

Run: `cd client && npx vitest run src/features/heatmap`
Expected: PASS (note the test count).

- [ ] **Step 2: Create the shared module**

`client/src/shared/utilization/types.ts`:

```ts
/** Utilization bands; they match the server's constants/insights.go. */
export type UtilizationBand = 'unavailable' | 'underused' | 'healthy' | 'near_full' | 'overloaded'
```

`client/src/shared/utilization/constants.ts` (the three maps move here unchanged from `features/heatmap/constants.ts`):

```ts
import type { PlainKey } from '@/shared/i18n'
import type { UtilizationBand } from './types'

export const BAND_CLASSES: Record<UtilizationBand, string> = {
  unavailable: 'bg-muted text-muted-foreground',
  underused: 'bg-underused/20',
  healthy: 'bg-healthy/25',
  near_full: 'bg-warning/35',
  overloaded: 'bg-destructive/30',
}

export const BAND_LABELS: Record<UtilizationBand, PlainKey> = {
  unavailable: 'Unavailable',
  underused: 'Underused',
  healthy: 'Healthy',
  near_full: 'Near full',
  overloaded: 'Overloaded',
}

/** Legend order, from light to heavy load. */
export const BAND_ORDER: readonly UtilizationBand[] = ['underused', 'healthy', 'near_full', 'overloaded', 'unavailable']

/** Over 100 % of capacity; the calendar flags such a week. */
export const OVERLOADED_BAND: UtilizationBand = 'overloaded'
```

`client/src/shared/utilization/index.ts`:

```ts
export { BAND_CLASSES, BAND_LABELS, BAND_ORDER, OVERLOADED_BAND } from './constants'
export type { UtilizationBand } from './types'
```

- [ ] **Step 3: Point the heatmap at it**

In `client/src/features/heatmap/types.ts`, replace

```ts
/** They match the server's constants/insights.go. */
export type UtilizationBand = 'unavailable' | 'underused' | 'healthy' | 'near_full' | 'overloaded'
```

with

```ts
import type { UtilizationBand } from '@/shared/utilization'

export type { UtilizationBand }
```

(keep the `import type { PersonRef }` line first; the file's interfaces keep using `UtilizationBand`).

In `client/src/features/heatmap/constants.ts`, delete `BAND_CLASSES`, `BAND_LABELS` and `BAND_ORDER` and the two now unused type imports, and add at the top:

```ts
export { BAND_CLASSES, BAND_LABELS, BAND_ORDER } from '@/shared/utilization'
```

The heatmap's components keep importing from `../../constants`, so nothing else changes.

- [ ] **Step 4: Run the checks**

Run: `cd client && npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: no type or lint errors; the heatmap tests pass with the same count as Step 1; build succeeds.

- [ ] **Step 5: Commit**

```bash
git add client/src/shared/utilization client/src/features/heatmap/types.ts client/src/features/heatmap/constants.ts
git commit -m "refactor(client): utilization bands in shared/utilization

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 6: `shared/people`: person, absences and schedule

**Files:**
- Create: `client/src/shared/people/types.ts`, `constants.ts`, `api.ts`, `use-people.ts`, `index.ts`, `api.test.ts`
- Modify: `client/src/resources/endpoints/people.ts`
- Modify: `client/src/features/people/types.ts`, `client/src/features/people/constants.ts`
- Modify: `client/src/features/absences/types.ts`, `client/src/features/absences/constants.ts`, `client/src/features/absences/use-absences.ts`
- Modify: `client/src/features/holidays/use-holidays.ts`

**Interfaces:**
- Consumes: Task 5 `UtilizationBand`; existing `request<T>(config)` from `@/shared/api/http`, `INSIGHTS_QUERY_KEY` from `@/utils/constants`.
- Produces (`@/shared/people`): types `Person`, `PersonSkill`, `AbsenceType`, `AbsenceStatus`, `PersonSchedule`, `ScheduleWeek`, `ScheduleProject`, `ScheduleDay`, `ScheduleAbsence`, `ScheduleHoliday`; constants `PEOPLE_QUERY_KEY`, `personQueryKey(id)`, `scheduleQueryKey(personId, from, to)`, `ABSENCE_TYPE_LABELS`, `ABSENCE_STATUS_LABELS`, `PLANNED_ABSENCE_STATUS`; functions `getPerson(id)`, `getSchedule(personId, from, to)`; hooks `usePerson(id: string)`, `useSchedule(personId: string, from: string, to: string)`. Endpoint `PEOPLE_ENDPOINTS.schedule(id)`.

- [ ] **Step 1: Write the failing test**

Create `client/src/shared/people/api.test.ts`:

```ts
import { afterEach, describe, expect, it, vi } from 'vitest'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { getPerson, getSchedule } from './api'
import { PEOPLE_QUERY_KEY, personQueryKey, scheduleQueryKey } from './constants'

const { request } = vi.hoisted(() => ({ request: vi.fn() }))
vi.mock('@/shared/api/http', () => ({ request }))

afterEach(() => {
  request.mockReset()
})

describe('people api', () => {
  it('asks for a schedule by person and week range', async () => {
    await getSchedule('p1', '2026-W40', '2026-W44')
    expect(request).toHaveBeenCalledWith({ url: '/people/p1/schedule', params: { from: '2026-W40', to: '2026-W44' } })
  })

  it('reads one person', async () => {
    await getPerson('p1')
    expect(request).toHaveBeenCalledWith({ url: '/people/p1' })
  })
})

describe('query keys', () => {
  it('keeps a schedule under insights, so accepting a staffing draft refreshes it', () => {
    expect(scheduleQueryKey('p1', '2026-W40', '2026-W44').slice(0, INSIGHTS_QUERY_KEY.length)).toEqual([...INSIGHTS_QUERY_KEY])
  })

  it('keeps a person under people, so saving a person refreshes the page header', () => {
    expect(personQueryKey('p1').slice(0, PEOPLE_QUERY_KEY.length)).toEqual([...PEOPLE_QUERY_KEY])
  })
})
```

- [ ] **Step 2: Run it to see it fail**

Run: `cd client && npx vitest run src/shared/people/api.test.ts`
Expected: FAIL, `Failed to resolve import "./api"`.

- [ ] **Step 3: Add the endpoint**

`client/src/resources/endpoints/people.ts`:

```ts
  assignments: (id: string) => `${person(id)}/assignments`,
  schedule: (id: string) => `${person(id)}/schedule`,
} as const
```

- [ ] **Step 4: Create the module**

`client/src/shared/people/types.ts` (`PersonSkill` and `Person` move here unchanged from `features/people/types.ts`; `AbsenceType` and `AbsenceStatus` from `features/absences/types.ts`):

```ts
import type { Role } from '@/resources/rbac'
import type { UtilizationBand } from '@/shared/utilization'

export interface PersonSkill {
  expertise_id: string
  expertise: string
  level: number
  primary: boolean
}

/** A person as `GET /people` returns it. Dates use the server format `2026-10-12`. */
export interface Person {
  id: string
  name: string
  email: string
  role: Role
  weekly_hours: number
  non_project_hours: number
  start_date: string | null
  end_date: string | null
  active: boolean
  skills: PersonSkill[]
}

/** Absence types and statuses; they match the server's constants/absence.go. */
export type AbsenceType = 'leave' | 'training' | 'holiday'
export type AbsenceStatus = 'planned' | 'approved'

/** A person's baseline hours on one project and expertise in a week. */
export interface ScheduleProject {
  program_id: string
  program_name: string
  project_id: string
  project_name: string
  expertise: string
  hours: number
}

/** A person's load against capacity in one ISO week, computed as on the heatmap. */
export interface ScheduleWeek {
  week: string
  capacity: number
  load: number
  /** load ÷ capacity; null when capacity is 0. */
  utilization: number | null
  band: UtilizationBand
  /** Planned (not yet approved) absence this week. */
  tentative: boolean
  projects: ScheduleProject[]
}

/** One of the person's absences on a day. */
export interface ScheduleAbsence {
  id: string
  type: AbsenceType
  status: AbsenceStatus
  hours_per_day: number
}

/** A company-wide holiday on a day; holidays have no name. */
export interface ScheduleHoliday {
  id: string
}

/** A working day with the person's absences and the holidays on it. */
export interface ScheduleDay {
  /** `2026-10-21` */
  date: string
  absences: ScheduleAbsence[]
  holidays: ScheduleHoliday[]
}

/** `GET /people/{id}/schedule`: weeks in the range, and only the days with something on them. */
export interface PersonSchedule {
  /** ISO day numbers that count: 1 is Monday, 7 is Sunday. */
  working_weekdays: number[]
  weeks: ScheduleWeek[]
  days: ScheduleDay[]
}
```

`client/src/shared/people/constants.ts`:

```ts
import type { PlainKey } from '@/shared/i18n'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import type { AbsenceStatus, AbsenceType } from './types'

export const PEOPLE_QUERY_KEY = ['people'] as const

/** One person, under the people prefix, so saving a person refreshes it. */
export const personQueryKey = (id: string) => [...PEOPLE_QUERY_KEY, 'detail', id] as const

/** A person's schedule, under the insights prefix, so accepting a staffing draft refreshes it. */
export const scheduleQueryKey = (personId: string, from: string, to: string) =>
  [...INSIGHTS_QUERY_KEY, 'schedule', personId, from, to] as const

export const ABSENCE_TYPE_LABELS: Record<AbsenceType, PlainKey> = {
  leave: 'Leave',
  training: 'Training',
  holiday: 'Holiday',
}

export const ABSENCE_STATUS_LABELS: Record<AbsenceStatus, PlainKey> = {
  planned: 'Planned',
  approved: 'Approved',
}

/** Reported or read from a calendar, not approved yet; a planner may approve it. */
export const PLANNED_ABSENCE_STATUS: AbsenceStatus = 'planned'
```

`client/src/shared/people/api.ts`:

```ts
import { PEOPLE_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { Person, PersonSchedule } from './types'

export function getPerson(id: string) {
  return request<Person>({ url: PEOPLE_ENDPOINTS.detail(id) })
}

/** One person's weeks and days from one ISO week through another. */
export function getSchedule(personId: string, from: string, to: string) {
  return request<PersonSchedule>({ url: PEOPLE_ENDPOINTS.schedule(personId), params: { from, to } })
}
```

`client/src/shared/people/use-people.ts`:

```ts
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getPerson, getSchedule } from './api'
import { personQueryKey, scheduleQueryKey } from './constants'

export function usePerson(id: string) {
  return useQuery({ queryKey: personQueryKey(id), queryFn: () => getPerson(id) })
}

/** A person's schedule; the previous month stays on screen while the next one loads. */
export function useSchedule(personId: string, from: string, to: string) {
  return useQuery({
    queryKey: scheduleQueryKey(personId, from, to),
    queryFn: () => getSchedule(personId, from, to),
    placeholderData: keepPreviousData,
  })
}
```

`client/src/shared/people/index.ts`:

```ts
export { getPerson, getSchedule } from './api'
export {
  ABSENCE_STATUS_LABELS,
  ABSENCE_TYPE_LABELS,
  PEOPLE_QUERY_KEY,
  personQueryKey,
  PLANNED_ABSENCE_STATUS,
  scheduleQueryKey,
} from './constants'
export type {
  AbsenceStatus,
  AbsenceType,
  Person,
  PersonSchedule,
  PersonSkill,
  ScheduleAbsence,
  ScheduleDay,
  ScheduleHoliday,
  ScheduleProject,
  ScheduleWeek,
} from './types'
export { usePerson, useSchedule } from './use-people'
```

- [ ] **Step 5: Re-export from the features that owned them**

`client/src/features/people/types.ts`: delete the `PersonSkill` and `Person` interfaces and add, after the `Role` import:

```ts
import type { Person } from '@/shared/people'

export type { Person, PersonSkill } from '@/shared/people'
```

`client/src/features/people/constants.ts`: replace `export const PEOPLE_QUERY_KEY = ['people'] as const` with

```ts
export { PEOPLE_QUERY_KEY } from '@/shared/people'
```

`client/src/features/absences/types.ts`: replace the two type lines (and their comment) with

```ts
import type { AbsenceStatus, AbsenceType } from '@/shared/people'

export type { AbsenceStatus, AbsenceType }
```

`client/src/features/absences/constants.ts`: delete `ABSENCE_TYPE_LABELS` and `ABSENCE_STATUS_LABELS` and the `PlainKey` import, and add

```ts
export { ABSENCE_STATUS_LABELS, ABSENCE_TYPE_LABELS } from '@/shared/people'
```

- [ ] **Step 6: Refresh the calendar and heatmap after an absence or holiday changes**

`client/src/features/absences/use-absences.ts`: import `INSIGHTS_QUERY_KEY` from `@/utils/constants`, and give both mutations this `onSuccess`:

```ts
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: absencesQueryKey(personId) })
      await queryClient.invalidateQueries({ queryKey: INSIGHTS_QUERY_KEY })
    },
```

`client/src/features/holidays/use-holidays.ts`: the same for both mutations, with `HOLIDAYS_QUERY_KEY` in place of `absencesQueryKey(personId)`.

- [ ] **Step 7: Run the test to see it pass, then all checks**

Run: `cd client && npx vitest run src/shared/people/api.test.ts`
Expected: PASS, 4 tests.

Run: `cd client && npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: all green.

- [ ] **Step 8: Commit**

```bash
git add client/src/shared/people client/src/resources/endpoints/people.ts client/src/features/people/types.ts \
  client/src/features/people/constants.ts client/src/features/absences client/src/features/holidays/use-holidays.ts
git commit -m "feat(client): shared/people with a person's schedule

Person and absence types move to shared/people; absence and holiday
changes refresh insights.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 7: Approve on planned absences

**Files:**
- Modify: `client/src/resources/endpoints/availability.ts`
- Modify: `client/src/features/absences/api.ts`, `client/src/features/absences/use-absences.ts`, `client/src/features/absences/component.tsx`
- Modify: `client/src/features/absences/components/absence-row/component.tsx`, `client/src/features/absences/components/absence-row/types.ts`
- Create: `client/src/features/absences/use-absences.test.tsx`, `client/src/features/absences/components/absence-row/component.test.tsx`
- Modify: `client/src/shared/i18n/language-key.ts`

**Interfaces:**
- Consumes: Task 4 `POST /absences/{id}/approve` (returns the absence); Task 6 `PLANNED_ABSENCE_STATUS`; existing `absencesQueryKey(personId)`, `INSIGHTS_QUERY_KEY` (the schedule's prefix), `AbsenceList({ personId, canApprove })`. Callers already pass `canApprove={can(user?.role, 'absences:approve')}` (People page dialog), `canApprove` inside the planner-only block (person page, Task 10) and `canApprove={false}` (My schedule).
- Produces: `ABSENCE_ENDPOINTS.approve(id)`; `approveAbsence(id: string)`; `useApproveAbsence(personId: string)`; `AbsenceRowProps.onApprove?: (id: string) => void`.

- [ ] **Step 1: Write the failing tests**

`client/src/features/absences/use-absences.test.tsx`:

```tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { absencesQueryKey } from './constants'
import { useApproveAbsence } from './use-absences'

const { approveAbsence } = vi.hoisted(() => ({ approveAbsence: vi.fn() }))
vi.mock('./api', () => ({ approveAbsence, createAbsence: vi.fn(), deleteAbsence: vi.fn(), listAbsences: vi.fn() }))

describe('useApproveAbsence', () => {
  it('approves, then refreshes the absences and the schedule', async () => {
    approveAbsence.mockResolvedValue({ id: 'a1', status: 'approved' })
    const queryClient = new QueryClient()
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries')
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )

    const { result } = renderHook(() => useApproveAbsence('p1'), { wrapper })
    await act(() => result.current.mutateAsync('a1'))

    expect(approveAbsence).toHaveBeenCalledWith('a1')
    expect(invalidate).toHaveBeenCalledWith({ queryKey: absencesQueryKey('p1') })
    // The schedule lives under the insights prefix (Task 6), so this refreshes the calendar too.
    expect(invalidate).toHaveBeenCalledWith({ queryKey: INSIGHTS_QUERY_KEY })
  })
})
```

`client/src/features/absences/components/absence-row/component.test.tsx`:

```tsx
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Table, TableBody } from '@/shared/ui/table'
import type { Absence } from '../../types'
import { AbsenceRow } from './component'

const planned: Absence = {
  id: 'a1',
  person_id: 'p1',
  from_date: '2026-10-22',
  to_date: '2026-10-23',
  hours_per_day: 8,
  type: 'leave',
  status: 'planned',
}
const approved: Absence = { ...planned, status: 'approved' }
const onDelete = vi.fn()

function renderRow(absence: Absence, onApprove?: (id: string) => void) {
  render(
    <Table>
      <TableBody>
        <AbsenceRow absence={absence} onDelete={onDelete} onApprove={onApprove} />
      </TableBody>
    </Table>,
  )
}

afterEach(cleanup)

describe('AbsenceRow', () => {
  it('approves a planned absence', () => {
    const onApprove = vi.fn()
    renderRow(planned, onApprove)
    fireEvent.click(screen.getByRole('button', { name: 'Approve' }))
    expect(onApprove).toHaveBeenCalledWith('a1')
  })

  it('offers no Approve on an approved absence', () => {
    renderRow(approved, vi.fn())
    expect(screen.queryByRole('button', { name: 'Approve' })).toBeNull()
  })

  it('offers no Approve to someone who may not approve', () => {
    renderRow(planned)
    expect(screen.queryByRole('button', { name: 'Approve' })).toBeNull()
  })
})
```

- [ ] **Step 2: Run them to see them fail**

Run: `cd client && npx vitest run src/features/absences`
Expected: FAIL: `useApproveAbsence` is not exported, and no button named `Approve`.

- [ ] **Step 3: Endpoint and API**

`client/src/resources/endpoints/availability.ts`:

```ts
export const ABSENCE_ENDPOINTS = {
  detail: (id: string) => `/absences/${encodeURIComponent(id)}`,
  approve: (id: string) => `/absences/${encodeURIComponent(id)}/approve`,
} as const
```

`client/src/features/absences/api.ts`, at the end:

```ts
/** Turns a planned absence into an approved one (planners only). */
export function approveAbsence(id: string) {
  return request<Absence>({ url: ABSENCE_ENDPOINTS.approve(id), method: 'POST' })
}
```

- [ ] **Step 4: The hook**

`client/src/features/absences/use-absences.ts`: add `approveAbsence` to the `./api` import and:

```ts
/** Approves a planned absence, then refreshes the list and the insights (heatmap and schedule). */
export function useApproveAbsence(personId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => approveAbsence(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: absencesQueryKey(personId) })
      await queryClient.invalidateQueries({ queryKey: INSIGHTS_QUERY_KEY })
    },
  })
}
```

(`INSIGHTS_QUERY_KEY` is already imported in this file since Task 6.)

- [ ] **Step 5: The row and the list**

`client/src/features/absences/components/absence-row/types.ts`:

```ts
import type { Absence } from '../../types'

export interface AbsenceRowProps {
  absence: Absence
  onDelete: (id: string) => void
  /** Approves a planned absence; without it the row has no Approve button. */
  onApprove?: (id: string) => void
}
```

`client/src/features/absences/components/absence-row/component.tsx`:
- import `PLANNED_ABSENCE_STATUS` from `@/shared/people`
- take `onApprove` from the props and add `const handleApprove = () => onApprove?.(absence.id)` next to `handleDelete`
- replace the last cell with:

```tsx
      <TableCell>
        <div className="flex justify-end gap-2">
          {onApprove && absence.status === PLANNED_ABSENCE_STATUS && (
            <Button variant="outline" size="sm" onClick={handleApprove}>
              {t('Approve')}
            </Button>
          )}
          <Button variant="ghost" size="icon-sm" aria-label={t('Remove absence')} onClick={handleDelete}>
            <Trash2 />
          </Button>
        </div>
      </TableCell>
```

`client/src/features/absences/component.tsx`:
- import `useApproveAbsence` with the other hooks from `./use-absences`
- in the body, after `remove`: `const approve = useApproveAbsence(personId)` and `const handleApprove = (id: string) => approve.mutate(id)`
- pass it to each row: `<AbsenceRow key={absence.id} absence={absence} onDelete={handleDelete} onApprove={canApprove ? handleApprove : undefined} />`

Add `Approve: string` to `LanguageKey` in its alphabetical place (before `Approved`).

- [ ] **Step 6: Run the tests to see them pass, then all checks**

Run: `cd client && npx vitest run src/features/absences`
Expected: PASS, 1 hook test and 3 row tests.

Run: `cd client && npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: all green.

- [ ] **Step 7: Commit**

```bash
git add client/src/resources/endpoints/availability.ts client/src/features/absences client/src/shared/i18n/language-key.ts
git commit -m "feat(client): approve a planned absence

Planners see Approve on planned absences; approving refreshes the list,
the heatmap and the person's calendar.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 8: Month helpers

**Files:**
- Modify: `client/src/utils/constants/date.ts`, `client/src/utils/functions/date.ts`, `client/src/utils/functions/index.ts`
- Create: `client/src/utils/functions/date.test.ts`

**Interfaces:**
- Consumes: existing `dayjs`, `toIsoWeek`, `isoWeekList`, `isoWeekRange` in `date.ts`.
- Produces (`@/utils/functions`): `currentMonth(): string`, `addMonths(month: string, n: number): string`, `monthIsoWeeks(month: string): string[]`, `formatMonth(month: string, format?: string): string`, `isInMonth(date: string, month: string): boolean`, `isoWeekDate(week: string, isoWeekday: number): string`, `formatIsoWeekday(isoWeekday: number, format?: string): string`, `todayDateValue(): string`. A month is `'YYYY-MM'`. `DATE_FORMAT.isoMonth`, `.monthTitle`, `.dayOfMonth`, `.weekdayShort`.

- [ ] **Step 1: Write the failing test**

Create `client/src/utils/functions/date.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { addMonths, formatIsoWeekday, formatMonth, isInMonth, isoWeekDate, monthIsoWeeks } from './date'

describe('monthIsoWeeks', () => {
  it('lists the ISO weeks that touch a month', () => {
    expect(monthIsoWeeks('2026-10')).toEqual(['2026-W40', '2026-W41', '2026-W42', '2026-W43', '2026-W44'])
  })

  it('gives six weeks to a month that starts on a Sunday and ends on a Monday', () => {
    expect(monthIsoWeeks('2026-11')).toEqual(['2026-W44', '2026-W45', '2026-W46', '2026-W47', '2026-W48', '2026-W49'])
  })

  it('starts January in the last ISO week of the year before', () => {
    expect(monthIsoWeeks('2027-01')).toEqual(['2026-W53', '2027-W01', '2027-W02', '2027-W03', '2027-W04'])
  })
})

describe('addMonths', () => {
  it('moves across the year boundary both ways', () => {
    expect(addMonths('2026-12', 1)).toBe('2027-01')
    expect(addMonths('2026-01', -1)).toBe('2025-12')
  })
})

describe('day helpers', () => {
  it('finds the date of a weekday in a week, Monday being 1', () => {
    expect(isoWeekDate('2026-W42', 1)).toBe('2026-10-12')
    expect(isoWeekDate('2026-W42', 3)).toBe('2026-10-14')
    expect(isoWeekDate('2026-W42', 7)).toBe('2026-10-18')
  })

  it('tells whether a date is in a month', () => {
    expect(isInMonth('2026-10-31', '2026-10')).toBe(true)
    expect(isInMonth('2026-11-01', '2026-10')).toBe(false)
  })

  it('formats months and weekdays for headings', () => {
    expect(formatMonth('2026-10')).toBe('October 2026')
    expect(formatIsoWeekday(1)).toBe('Mon')
    expect(formatIsoWeekday(7)).toBe('Sun')
  })
})
```

- [ ] **Step 2: Run it to see it fail**

Run: `cd client && npx vitest run src/utils/functions/date.test.ts`
Expected: FAIL, `monthIsoWeeks is not a function` (and the other new names).

- [ ] **Step 3: Add the formats**

`client/src/utils/constants/date.ts`, inside `DATE_FORMAT` after `dateTime`:

```ts
  /** A month as the calendar keeps it, e.g. "2026-10". */
  isoMonth: 'YYYY-MM',
  /** A month heading, e.g. "October 2026". */
  monthTitle: 'MMMM YYYY',
  /** Day of the month in a calendar cell, e.g. "14". */
  dayOfMonth: 'D',
  /** Short weekday in a column header, e.g. "Mon". */
  weekdayShort: 'ddd',
```

- [ ] **Step 4: Add the helpers**

Append to `client/src/utils/functions/date.ts`:

```ts
/** The month of today, in the browser's time zone: `'2026-10'`. */
export function currentMonth(): string {
  return dayjs().format(DATE_FORMAT.isoMonth)
}

/** The month `n` months after `month` (before it when `n` is negative): `('2026-12', 1)` → `'2027-01'`. */
export function addMonths(month: string, n: number): string {
  return dayjs(month).add(n, 'month').format(DATE_FORMAT.isoMonth)
}

/** Every ISO week with a day in the month, in order: `'2026-10'` → `'2026-W40'` … `'2026-W44'`. */
export function monthIsoWeeks(month: string): string[] {
  const first = dayjs(month).startOf('month')
  return isoWeekList(toIsoWeek(first), toIsoWeek(first.endOf('month')))
}

/** A month for display: `'2026-10'` → `'October 2026'`. */
export function formatMonth(month: string, format: string = DATE_FORMAT.monthTitle): string {
  return dayjs(month).format(format)
}

/** Whether a date value (`'2026-10-31'`) falls in a month (`'2026-10'`). */
export function isInMonth(date: string, month: string): boolean {
  return dayjs(date).format(DATE_FORMAT.isoMonth) === month
}

/** The date of an ISO weekday (1 = Monday … 7 = Sunday) in a week: `('2026-W42', 3)` → `'2026-10-14'`. */
export function isoWeekDate(week: string, isoWeekday: number): string {
  return isoWeekRange(week).start.isoWeekday(isoWeekday).format(DATE_FORMAT.isoDate)
}

/** The short name of an ISO weekday for a column header: `1` → `'Mon'`. */
export function formatIsoWeekday(isoWeekday: number, format: string = DATE_FORMAT.weekdayShort): string {
  return dayjs().isoWeekday(isoWeekday).format(format)
}

/** Today as a date value (`'2026-10-14'`), in the browser's time zone. */
export function todayDateValue(): string {
  return dayjs().format(DATE_FORMAT.isoDate)
}
```

Add the eight names to the `./date` export list in `client/src/utils/functions/index.ts`, keeping it alphabetical:

```ts
export {
  addIsoWeeks,
  addMonths,
  currentIsoWeek,
  currentMonth,
  dayjs,
  formatDate,
  formatDateRange,
  formatDateTime,
  formatIsoWeekday,
  formatIsoWeekRange,
  formatIsoWeekSpan,
  formatMonth,
  isInMonth,
  isIsoWeek,
  isoWeekDate,
  isoWeekList,
  isoWeekRange,
  isoWeeksBetween,
  monthIsoWeeks,
  parseDateValue,
  shortWeekLabel,
  toDateValue,
  todayDateValue,
  toIsoWeek,
} from './date'
```

- [ ] **Step 5: Run the test to see it pass, then all checks**

Run: `cd client && npx vitest run src/utils/functions/date.test.ts`
Expected: PASS, 7 tests.

Run: `cd client && npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: all green.

- [ ] **Step 6: Commit**

```bash
git add client/src/utils/constants/date.ts client/src/utils/functions/date.ts client/src/utils/functions/index.ts \
  client/src/utils/functions/date.test.ts
git commit -m "feat(client): month helpers for calendars

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 9: `features/person-calendar`

**Files:**
- Create: `client/src/features/person-calendar/index.ts`, `component.tsx`, `types.ts`, `constants.ts`, `utils.ts`, `utils.test.ts`, `component.test.tsx`
- Create: `client/src/features/person-calendar/components/month-nav/{index.ts,component.tsx,types.ts}`
- Create: `client/src/features/person-calendar/components/week-row/{index.ts,component.tsx,types.ts,constants.ts}`
- Create: `client/src/features/person-calendar/components/week-row/components/day-cell/{index.ts,component.tsx,types.ts,constants.ts}`
- Create: `client/src/features/person-calendar/components/week-detail/{index.ts,component.tsx,types.ts,utils.ts}`
- Modify: `client/src/shared/i18n/language-key.ts`

**Interfaces:**
- Consumes: Task 5 `BAND_CLASSES`, `BAND_LABELS`, `OVERLOADED_BAND`; Task 6 `useSchedule`, `ABSENCE_TYPE_LABELS`, `PLANNED_ABSENCE_STATUS`, the schedule types; Task 8 month helpers; existing `Section`, `TableLoading`, `TextLink`, `Table*`, `Badge` and `BadgeVariant`, `ROUTES.projectDetail`, `formatDate`, `formatIsoWeekRange`, `shortWeekLabel`, `cn`.
- Produces: `PersonCalendar({ personId }: PersonCalendarProps)` from `@/features/person-calendar`.

- [ ] **Step 1: Write the failing tests**

`client/src/features/person-calendar/utils.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import type { ScheduleDay } from '@/shared/people'
import { daysByDate, findWeek, orderedWeekdays, weekCells } from './utils'

const leave: ScheduleDay = {
  date: '2026-10-14',
  absences: [{ id: 'a1', type: 'leave', status: 'approved', hours_per_day: 8 }],
  holidays: [],
}

describe('orderedWeekdays', () => {
  it('puts Monday first and Sunday last', () => {
    expect(orderedWeekdays([7, 3, 1])).toEqual([1, 3, 7])
  })
})

describe('weekCells', () => {
  it('gives each working weekday its date and what is on it', () => {
    const cells = weekCells('2026-W42', [1, 2, 3, 4, 5], daysByDate([leave]))
    expect(cells.map((cell) => cell.date)).toEqual(['2026-10-12', '2026-10-13', '2026-10-14', '2026-10-15', '2026-10-16'])
    expect(cells[2]).toBe(leave)
    expect(cells[0]).toEqual({ date: '2026-10-12', absences: [], holidays: [] })
  })
})

describe('findWeek', () => {
  it('finds nothing when no week is selected', () => {
    expect(findWeek([], null)).toBeUndefined()
  })
})
```

`client/src/features/person-calendar/component.test.tsx`:

```tsx
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { PersonSchedule, ScheduleWeek } from '@/shared/people'
import { PersonCalendar } from './component'

const { useSchedule } = vi.hoisted(() => ({ useSchedule: vi.fn() }))
vi.mock('@/shared/people', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/people')>()),
  useSchedule,
}))

const week = (label: string, overrides: Partial<ScheduleWeek> = {}): ScheduleWeek => ({
  week: label,
  capacity: 36,
  load: 0,
  utilization: 0,
  band: 'underused',
  tentative: false,
  projects: [],
  ...overrides,
})

const october: PersonSchedule = {
  working_weekdays: [1, 2, 3, 4, 5],
  weeks: [
    week('2026-W40'),
    week('2026-W41'),
    week('2026-W42', {
      load: 40,
      utilization: 1.11,
      band: 'overloaded',
      projects: [
        { program_id: 'g1', program_name: 'Acme', project_id: 'p1', project_name: 'Checkout Revamp', expertise: 'Engineer', hours: 20 },
        { program_id: 'g1', program_name: 'Acme', project_id: 'p2', project_name: 'Data Migration', expertise: 'Engineer', hours: 20 },
      ],
    }),
    week('2026-W43', { tentative: true }),
    week('2026-W44'),
  ],
  days: [
    { date: '2026-10-21', absences: [], holidays: [{ id: 'h1' }] },
    { date: '2026-10-22', absences: [{ id: 'a1', type: 'leave', status: 'planned', hours_per_day: 8 }], holidays: [] },
  ],
}

function renderCalendar() {
  render(
    <MemoryRouter>
      <PersonCalendar personId="u1" />
    </MemoryRouter>,
  )
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date('2026-10-14T09:00:00'))
  useSchedule.mockReturnValue({ data: october, isPending: false })
})

afterEach(() => {
  cleanup()
  useSchedule.mockReset()
  vi.useRealTimers()
})

describe('PersonCalendar', () => {
  it('asks for the weeks of the current month and draws one row per week', () => {
    renderCalendar()
    expect(useSchedule).toHaveBeenCalledWith('u1', '2026-W40', '2026-W44')
    expect(screen.getByText('October 2026')).toBeTruthy()
    expect(screen.getAllByRole('row')).toHaveLength(6) // the header and five weeks
  })

  it('marks holidays and planned leave on their days', () => {
    renderCalendar()
    expect(screen.getByText('Holiday')).toBeTruthy()
    expect(screen.getByText('Leave (planned)')).toBeTruthy()
  })

  it('lists the hours per project of the selected week, each linking to its project', () => {
    renderCalendar()
    fireEvent.click(screen.getByRole('button', { name: 'Hours in 2026-W42' }))
    expect(screen.getByRole('link', { name: 'Checkout Revamp' }).getAttribute('href')).toBe('/programs/g1/projects/p1')
    expect(screen.getAllByText('20 h')).toHaveLength(2)
  })

  it('moves to the next month, which spans six weeks', () => {
    renderCalendar()
    fireEvent.click(screen.getByRole('button', { name: 'Next month' }))
    expect(useSchedule).toHaveBeenLastCalledWith('u1', '2026-W44', '2026-W49')
    expect(screen.getByText('November 2026')).toBeTruthy()
  })
})
```

- [ ] **Step 2: Run them to see them fail**

Run: `cd client && npx vitest run src/features/person-calendar`
Expected: FAIL, `Failed to resolve import "./utils"` and `"./component"`.

- [ ] **Step 3: Add the i18n keys**

Add to `LanguageKey` in `client/src/shared/i18n/language-key.ts`, each in its alphabetical place:

```ts
  'Hours by week against capacity, with leave and holidays by day.': string
  'Hours in {week}': string
  Load: string
  'Next month': string
  'Previous month': string
  Today: string
  Week: string
  '{type} (planned)': string
```

- [ ] **Step 4: Write the feature's own files**

`client/src/features/person-calendar/types.ts`:

```ts
export interface PersonCalendarProps {
  personId: string
}
```

`client/src/features/person-calendar/constants.ts`:

```ts
/** The week label column and the load column on either side of the weekday columns. */
export const CALENDAR_LABEL_COLUMNS = 2
```

`client/src/features/person-calendar/utils.ts`:

```ts
import type { ScheduleDay, ScheduleWeek } from '@/shared/people'
import { isoWeekDate } from '@/utils/functions'

/** The working weekdays in week order, Monday first. */
export function orderedWeekdays(weekdays: number[]): number[] {
  return [...weekdays].sort((a, b) => a - b)
}

/** The schedule's days by date, for drawing the grid. */
export function daysByDate(days: ScheduleDay[]): Map<string, ScheduleDay> {
  return new Map(days.map((day) => [day.date, day]))
}

/** One grid row: each working weekday of the week, with what is on that day. */
export function weekCells(week: string, weekdays: number[], days: Map<string, ScheduleDay>): ScheduleDay[] {
  return weekdays.map((weekday) => {
    const date = isoWeekDate(week, weekday)
    return days.get(date) ?? { date, absences: [], holidays: [] }
  })
}

/** The schedule's week with this label; nothing when no week is selected. */
export function findWeek(weeks: ScheduleWeek[], week: string | null): ScheduleWeek | undefined {
  return week === null ? undefined : weeks.find((candidate) => candidate.week === week)
}
```

`client/src/features/person-calendar/index.ts`:

```ts
export { PersonCalendar } from './component'
export type { PersonCalendarProps } from './types'
```

- [ ] **Step 5: Write the month navigation**

`components/month-nav/types.ts`:

```ts
export interface MonthNavProps {
  /** `'2026-10'` */
  month: string
  onChange: (month: string) => void
}
```

`components/month-nav/component.tsx`:

```tsx
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { addMonths, currentMonth } from '@/utils/functions'
import type { MonthNavProps } from './types'

/** Previous month, back to this month, next month. */
export function MonthNav({ month, onChange }: MonthNavProps) {
  const { t } = useTranslation()

  const handlePrevious = () => onChange(addMonths(month, -1))
  const handleToday = () => onChange(currentMonth())
  const handleNext = () => onChange(addMonths(month, 1))

  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="icon-sm" aria-label={t('Previous month')} onClick={handlePrevious}>
        <ChevronLeft />
      </Button>
      <Button variant="outline" size="sm" onClick={handleToday}>
        {t('Today')}
      </Button>
      <Button variant="outline" size="icon-sm" aria-label={t('Next month')} onClick={handleNext}>
        <ChevronRight />
      </Button>
    </div>
  )
}
```

`components/month-nav/index.ts`:

```ts
export { MonthNav } from './component'
```

- [ ] **Step 6: Write the day cell**

`components/week-row/components/day-cell/types.ts`:

```ts
import type { ScheduleDay } from '@/shared/people'

export interface DayCellProps {
  day: ScheduleDay
  /** The month on show; days outside it are dimmed. */
  month: string
}
```

`components/week-row/components/day-cell/constants.ts`:

```ts
import type { AbsenceType } from '@/shared/people'
import type { BadgeVariant } from '@/shared/ui/badge'

/** Approved absence and holidays by type. */
export const ABSENCE_BADGE: Record<AbsenceType, BadgeVariant> = {
  leave: 'violet',
  training: 'teal',
  holiday: 'pink',
}

/** Planned absence of any type, until a planner approves it. */
export const PLANNED_ABSENCE_BADGE: BadgeVariant = 'amber'

/** Today's cell is outlined, so the current day is easy to find. */
export const TODAY_CELL_CLASS = 'ring-2 ring-primary ring-inset'
```

`components/week-row/components/day-cell/component.tsx`:

```tsx
import { useTranslation } from '@/shared/i18n'
import { ABSENCE_TYPE_LABELS, PLANNED_ABSENCE_STATUS } from '@/shared/people'
import { Badge } from '@/shared/ui/badge'
import { TableCell } from '@/shared/ui/table'
import { DATE_FORMAT } from '@/utils/constants'
import { cn, formatDate, isInMonth, todayDateValue } from '@/utils/functions'
import { ABSENCE_BADGE, PLANNED_ABSENCE_BADGE, TODAY_CELL_CLASS } from './constants'
import type { DayCellProps } from './types'

/** One working day: its date, the holidays and the person's absences; planned absence is marked. */
export function DayCell({ day, month }: DayCellProps) {
  const { t } = useTranslation()

  return (
    <TableCell
      className={cn(
        'align-top',
        !isInMonth(day.date, month) && 'text-muted-foreground',
        day.date === todayDateValue() && TODAY_CELL_CLASS,
      )}
    >
      <div className="flex min-h-14 flex-col items-start gap-1">
        <span className="text-xs">{formatDate(day.date, DATE_FORMAT.dayOfMonth)}</span>
        {day.holidays.map((holiday) => (
          <Badge key={holiday.id} variant={ABSENCE_BADGE.holiday}>
            {t(ABSENCE_TYPE_LABELS.holiday)}
          </Badge>
        ))}
        {day.absences.map((absence) => (
          <Badge
            key={absence.id}
            variant={absence.status === PLANNED_ABSENCE_STATUS ? PLANNED_ABSENCE_BADGE : ABSENCE_BADGE[absence.type]}
          >
            {absence.status === PLANNED_ABSENCE_STATUS
              ? t('{type} (planned)', { type: t(ABSENCE_TYPE_LABELS[absence.type]) })
              : t(ABSENCE_TYPE_LABELS[absence.type])}
          </Badge>
        ))}
      </div>
    </TableCell>
  )
}
```

`components/week-row/components/day-cell/index.ts`:

```ts
export { DayCell } from './component'
```

- [ ] **Step 7: Write the week row**

`components/week-row/types.ts`:

```ts
import type { ScheduleDay, ScheduleWeek } from '@/shared/people'

export interface WeekRowProps {
  week: ScheduleWeek
  /** The week's working days, in column order. */
  days: ScheduleDay[]
  /** The month on show; days outside it are dimmed. */
  month: string
  selected: boolean
  onSelect: (week: string) => void
}
```

`components/week-row/constants.ts`:

```ts
/** The week's load button; its band colour comes from BAND_CLASSES, as on the heatmap. */
export const LOAD_BUTTON_CLASS =
  'inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium hover:ring-2 hover:ring-ring focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none'
```

`components/week-row/component.tsx`:

```tsx
import { TriangleAlert } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { TableCell, TableRow } from '@/shared/ui/table'
import { BAND_CLASSES, OVERLOADED_BAND } from '@/shared/utilization'
import { cn, shortWeekLabel } from '@/utils/functions'
import { DayCell } from './components/day-cell'
import { LOAD_BUTTON_CLASS } from './constants'
import type { WeekRowProps } from './types'

/** One ISO week: its working days, then its load against capacity, which selects the week. */
export function WeekRow({ week, days, month, selected, onSelect }: WeekRowProps) {
  const { t } = useTranslation()

  const handleSelect = () => onSelect(week.week)

  return (
    <TableRow data-state={selected ? 'selected' : undefined}>
      <TableCell className="text-xs text-muted-foreground">{shortWeekLabel(week.week)}</TableCell>
      {days.map((day) => (
        <DayCell key={day.date} day={day} month={month} />
      ))}
      <TableCell className="text-right">
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
      </TableCell>
    </TableRow>
  )
}
```

`components/week-row/index.ts`:

```ts
export { WeekRow } from './component'
```

- [ ] **Step 8: Write the week detail**

`components/week-detail/types.ts`:

```ts
import type { ScheduleWeek } from '@/shared/people'

export interface WeekDetailProps {
  week: ScheduleWeek
}
```

`components/week-detail/utils.ts`:

```ts
import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import type { ScheduleProject } from '@/shared/people'

/** The page of the project behind one line of the week. */
export function projectPath(project: ScheduleProject): string {
  return generatePath(ROUTES.projectDetail, { programId: project.program_id, id: project.project_id })
}
```

`components/week-detail/component.tsx`:

```tsx
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { BAND_LABELS } from '@/shared/utilization'
import { formatIsoWeekRange } from '@/utils/functions'
import type { WeekDetailProps } from './types'
import { projectPath } from './utils'

/** The selected week: load against capacity, then the hours per project, each linking to its project. */
export function WeekDetail({ week }: WeekDetailProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-3 rounded-lg border p-4 text-sm">
      <div className="flex flex-col">
        <span className="font-medium">
          {t('{week} ({dates})', { week: week.week, dates: formatIsoWeekRange(week.week) })}
        </span>
        <span className="text-muted-foreground">
          {t('{load} of {capacity} h · {band}', {
            load: week.load,
            capacity: week.capacity,
            band: t(BAND_LABELS[week.band]),
          })}
        </span>
      </div>
      {week.tentative && (
        <span className="text-xs text-muted-foreground">{t('Planned leave this week, not approved yet.')}</span>
      )}
      {week.projects.length === 0 && <span className="text-muted-foreground">{t('No assignments this week.')}</span>}
      <ul className="flex flex-col gap-2">
        {week.projects.map((project) => (
          <li key={`${project.project_id}-${project.expertise}`} className="flex items-start justify-between gap-2">
            <div className="flex min-w-0 flex-col">
              <TextLink to={projectPath(project)}>{project.project_name}</TextLink>
              <span className="text-xs text-muted-foreground">
                {t('{expertise} · {program}', { expertise: project.expertise, program: project.program_name })}
              </span>
            </div>
            <span className="shrink-0">{t('{hours} h', { hours: project.hours })}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}
```

`components/week-detail/index.ts`:

```ts
export { WeekDetail } from './component'
```

- [ ] **Step 9: Write the calendar**

`client/src/features/person-calendar/component.tsx`:

```tsx
import { useState } from 'react'
import { Section } from '@/shared/components/section'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { useSchedule } from '@/shared/people'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { currentMonth, formatIsoWeekday, formatMonth, monthIsoWeeks } from '@/utils/functions'
import { MonthNav } from './components/month-nav'
import { WeekDetail } from './components/week-detail'
import { WeekRow } from './components/week-row'
import { CALENDAR_LABEL_COLUMNS } from './constants'
import type { PersonCalendarProps } from './types'
import { daysByDate, findWeek, orderedWeekdays, weekCells } from './utils'

/** One person's month: absences and holidays by day, load against capacity and hours per project by week. */
export function PersonCalendar({ personId }: PersonCalendarProps) {
  const { t } = useTranslation()
  const [month, setMonth] = useState(currentMonth)
  const [selectedWeek, setSelectedWeek] = useState<string | null>(null)
  const weeks = monthIsoWeeks(month)
  const { data: schedule, isPending } = useSchedule(personId, weeks[0], weeks[weeks.length - 1])
  const weekdays = orderedWeekdays(schedule?.working_weekdays ?? [])
  const days = daysByDate(schedule?.days ?? [])
  const selected = findWeek(schedule?.weeks ?? [], selectedWeek)

  const handleMonthChange = (next: string) => {
    setMonth(next)
    setSelectedWeek(null)
  }

  return (
    <Section
      title={formatMonth(month)}
      description={t('Hours by week against capacity, with leave and holidays by day.')}
      action={<MonthNav month={month} onChange={handleMonthChange} />}
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Week')}</TableHead>
            {weekdays.map((weekday) => (
              <TableHead key={weekday}>{formatIsoWeekday(weekday)}</TableHead>
            ))}
            <TableHead className="text-right">{t('Load')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {schedule?.weeks.map((week) => (
            <WeekRow
              key={week.week}
              week={week}
              days={weekCells(week.week, weekdays, days)}
              month={month}
              selected={week.week === selectedWeek}
              onSelect={setSelectedWeek}
            />
          ))}
          {isPending && <TableLoading columns={weekdays.length + CALENDAR_LABEL_COLUMNS} />}
        </TableBody>
      </Table>
      {selected && <WeekDetail week={selected} />}
    </Section>
  )
}
```

- [ ] **Step 10: Run the tests to see them pass, then all checks**

Run: `cd client && npx vitest run src/features/person-calendar`
Expected: PASS, 3 utils tests and 4 component tests.

Run: `cd client && npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: all green. If the structure check flags a file, fix the file, never the check.

- [ ] **Step 11: Commit**

```bash
git add client/src/features/person-calendar client/src/shared/i18n/language-key.ts
git commit -m "feat(client): person calendar, a month of one person's weeks

Absences and holidays by day, load against capacity by week, and the
hours per project of the selected week.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 10: Person page, route, permission and breadcrumb

**Files:**
- Create: `client/src/pages/person-detail/index.ts`, `component.tsx`, `component.test.tsx`
- Create: `client/src/app/breadcrumbs/components/person-trail/{index.ts,component.tsx,types.ts}`
- Modify: `client/src/resources/routes/index.ts`, `client/src/resources/rbac/permissions.ts`, `client/src/app/router.tsx`
- Modify: `client/src/app/breadcrumbs/types.ts`, `client/src/app/breadcrumbs/component.tsx`, `client/src/app/breadcrumbs/component.test.tsx`
- Modify: `client/src/app/breadcrumbs/components/trail/component.tsx`, `client/src/app/breadcrumbs/components/trail/types.ts`
- Modify: `client/src/shared/i18n/language-key.ts`

**Interfaces:**
- Consumes: Task 6 `usePerson`; Task 7 `AbsenceList` with Approve; Task 9 `PersonCalendar`; existing `useCurrentUser`, `can`, `Section`, `Trail`.
- Produces: `ROUTES.personDetail = '/people/:id'`; permission `'people:view'`; `PersonDetailPage`; breadcrumb kind `'personDetail'`.

- [ ] **Step 1: Write the failing tests**

`client/src/pages/person-detail/component.test.tsx`:

```tsx
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/resources/routes'
import { PersonDetailPage } from './component'

const { useCurrentUser } = vi.hoisted(() => ({ useCurrentUser: vi.fn() }))
vi.mock('@/shared/auth', () => ({ useCurrentUser }))
vi.mock('@/shared/people', () => ({
  usePerson: () => ({
    data: {
      name: 'Alex Nguyen',
      weekly_hours: 40,
      non_project_hours: 4,
      skills: [{ expertise_id: 'eng', expertise: 'Engineer', level: 4, primary: true }],
    },
  }),
}))
vi.mock('@/features/person-calendar', () => ({ PersonCalendar: () => null }))
vi.mock('@/features/absences', () => ({ AbsenceList: () => null }))

function renderPage(role: string) {
  useCurrentUser.mockReturnValue({ data: { role } })
  render(
    <MemoryRouter initialEntries={['/people/u1']}>
      <Routes>
        <Route path={ROUTES.personDetail} element={<PersonDetailPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

afterEach(() => {
  cleanup()
  useCurrentUser.mockReset()
})

describe('PersonDetailPage', () => {
  it('heads the page with the person and their hours', () => {
    renderPage('planner')
    expect(screen.getByText('Alex Nguyen')).toBeTruthy()
    expect(screen.getByText('Engineer, 40 h per week, 4 h of them non-project')).toBeTruthy()
  })

  it('shows the absences to a planner only', () => {
    renderPage('planner')
    expect(screen.getByText('Absences')).toBeTruthy()
    cleanup()
    renderPage('lead')
    expect(screen.queryByText('Absences')).toBeNull()
  })
})
```

In `client/src/app/breadcrumbs/component.test.tsx`, add the mocks under the existing `vi.mock('@/shared/programs', …)`:

```tsx
vi.mock('@/shared/people', () => ({ usePerson: () => ({ data: { name: 'Alex Nguyen' } }) }))
const { useCurrentUser } = vi.hoisted(() => ({ useCurrentUser: vi.fn() }))
vi.mock('@/shared/auth', () => ({ useCurrentUser }))
```

the route in `renderAt` (after the `ROUTES.people` route):

```tsx
      { path: ROUTES.personDetail, element: <Breadcrumbs />, handle: handle('personDetail') },
```

and the test:

```tsx
  it('walks a planner from People to the person', () => {
    useCurrentUser.mockReturnValue({ data: { role: 'planner' } })
    renderAt('/people/u1')
    expect(screen.getByRole('link', { name: 'People' }).getAttribute('href')).toBe(ROUTES.people)
    expect(screen.getByText('Alex Nguyen').getAttribute('aria-current')).toBe('page')
  })

  it('shows People as plain text to a lead, who cannot open the People list', () => {
    useCurrentUser.mockReturnValue({ data: { role: 'lead' } })
    renderAt('/people/u1')
    expect(screen.queryByRole('link', { name: 'People' })).toBeNull()
    expect(screen.getByText('People').getAttribute('aria-current')).toBeNull()
    expect(screen.getByText('Alex Nguyen').getAttribute('aria-current')).toBe('page')
  })
```

- [ ] **Step 2: Run them to see them fail**

Run: `cd client && npx vitest run src/pages/person-detail src/app/breadcrumbs`
Expected: FAIL: the page test cannot resolve `./component`, and the new breadcrumb tests find no person trail because `ROUTES.personDetail` does not exist yet.

- [ ] **Step 3: Route and permission**

`client/src/resources/routes/index.ts`, after `people`:

```ts
  people: '/people',
  // Nested under People, so the sidebar marks People on a person's page.
  personDetail: '/people/:id',
```

`client/src/resources/rbac/permissions.ts`, before `'people:manage'`:

```ts
  'people:view': [ROLES.PLANNER, ROLES.MANAGER],
```

- [ ] **Step 4: The page**

`client/src/pages/person-detail/component.tsx`:

```tsx
import { useParams } from 'react-router'
import { AbsenceList } from '@/features/absences'
import { PersonCalendar } from '@/features/person-calendar'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { usePerson } from '@/shared/people'

/** One person: their month calendar and, for planners, their absences. */
export function PersonDetailPage() {
  const { t } = useTranslation()
  const { id = '' } = useParams()
  const { data: user } = useCurrentUser()
  const { data: person } = usePerson(id)
  const primary = person?.skills.find((skill) => skill.primary)

  return (
    <div className="flex flex-col gap-8">
      <Section
        variant="page"
        title={person?.name ?? t('Loading…')}
        description={
          person &&
          t('{expertise}, {hours} h per week, {nonProject} h of them non-project', {
            expertise: primary?.expertise ?? t('No primary expertise'),
            hours: person.weekly_hours,
            nonProject: person.non_project_hours,
          })
        }
      />
      <PersonCalendar personId={id} />
      {can(user?.role, 'absences:approve') && (
        <Section
          title={t('Absences')}
          description={t('Approved absence always lowers capacity; planned absence does when the setting is on.')}
        >
          <AbsenceList personId={id} canApprove />
        </Section>
      )}
    </div>
  )
}
```

Leads see the absences on the calendar only: the absence routes let a non-planner read and add only their own absences, so the list (with its add form) is for planners.

`client/src/pages/person-detail/index.ts`:

```ts
export { PersonDetailPage } from './component'
```

Add to `LanguageKey`, in its alphabetical place (with the `{` keys at the end):

```ts
  '{expertise}, {hours} h per week, {nonProject} h of them non-project': string
```

- [ ] **Step 5: Router and breadcrumb**

`client/src/app/router.tsx`: import `PersonDetailPage` from `@/pages/person-detail` (keep the imports sorted) and add after the `people:manage` block:

```tsx
          {
            element: <RequirePermission permission="people:view" />,
            children: [{ path: ROUTES.personDetail, Component: PersonDetailPage, handle: crumb('personDetail') }],
          },
```

`client/src/app/breadcrumbs/types.ts`: add `| 'personDetail'` after `| 'people'`.

`client/src/app/breadcrumbs/components/person-trail/types.ts`:

```ts
export interface PersonTrailProps {
  personId: string
}
```

`client/src/app/breadcrumbs/components/person-trail/component.tsx`:

```tsx
import { can } from '@/resources/rbac'
import { ROUTES } from '@/resources/routes'
import { useCurrentUser } from '@/shared/auth'
import { useTranslation } from '@/shared/i18n'
import { usePerson } from '@/shared/people'
import { Trail } from '../trail'
import type { PersonTrailProps } from './types'

/** People › the person's name. "People" is a link only for those who may open the People list. */
export function PersonTrail({ personId }: PersonTrailProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: person } = usePerson(personId)
  const peopleLink = can(user?.role, 'people:manage') ? ROUTES.people : undefined

  return <Trail crumbs={[{ label: t('People'), to: peopleLink }, { label: person?.name ?? t('Loading…') }]} />
}
```

`Trail` showed every crumb before the last as a link, and a crumb without `to` as the current page. A middle crumb without `to` must be plain text instead, not `aria-current="page"`. In `client/src/app/breadcrumbs/components/trail/component.tsx`, replace the `BreadcrumbItem` body inside the `crumbs.map`:

```tsx
            <BreadcrumbItem>
              {index === crumbs.length - 1 ? (
                <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
              ) : crumb.to ? (
                <BreadcrumbLink render={<Link to={crumb.to} />}>{crumb.label}</BreadcrumbLink>
              ) : (
                <span>{crumb.label}</span>
              )}
            </BreadcrumbItem>
```

and update the comment in `trail/types.ts`:

```ts
/** A step with its text ready to show. The last step is the current page; a step before it without `to` is plain text. */
```

`client/src/app/breadcrumbs/components/person-trail/index.ts`:

```ts
export { PersonTrail } from './component'
```

`client/src/app/breadcrumbs/component.tsx`: import `PersonTrail` from `./components/person-trail` and add before the `scenarioDetail` line:

```tsx
  if (kind === 'personDetail') return <PersonTrail personId={id} />
```

- [ ] **Step 6: Run the tests to see them pass, then all checks**

Run: `cd client && npx vitest run src/pages/person-detail src/app/breadcrumbs`
Expected: PASS (2 page tests, 5 breadcrumb tests).

Run: `cd client && npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: all green.

- [ ] **Step 7: Commit**

```bash
git add client/src/pages/person-detail client/src/app client/src/resources client/src/shared/i18n/language-key.ts
git commit -m "feat(client): person page with the month calendar

Planners and leads open /people/:id; breadcrumbs read People › name.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 11: People rows open the person; My schedule shows the calendar

**Files:**
- Modify: `client/src/features/people/components/person-row/component.tsx`
- Create: `client/src/features/people/components/person-row/component.test.tsx`
- Modify: `client/src/pages/my-schedule/component.tsx`
- Create: `client/src/pages/my-schedule/component.test.tsx`
- Modify: `client/src/shared/i18n/language-key.ts`

**Interfaces:**
- Consumes: Task 9 `PersonCalendar`; Task 10 `ROUTES.personDetail`; existing `LinkRow`, `TextLink`, `AbsenceList`, `useCurrentUser`, `PagePlaceholder`, `Section`.
- Produces: nothing new for other tasks.

- [ ] **Step 1: Write the failing tests**

`client/src/features/people/components/person-row/component.test.tsx`:

```tsx
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Table, TableBody } from '@/shared/ui/table'
import type { Person } from '../../types'
import { PersonRow } from './component'

vi.mock('@/shared/expertises', () => ({ useExpertiseBadge: () => () => 'blue' }))

const person: Person = {
  id: 'p1',
  name: 'Alex Nguyen',
  email: 'alex@example.test',
  role: 'specialist',
  weekly_hours: 40,
  non_project_hours: 4,
  start_date: null,
  end_date: null,
  active: true,
  skills: [],
}

afterEach(cleanup)

describe('PersonRow', () => {
  it('links the name to the person page and keeps Edit as its own action', () => {
    const onEdit = vi.fn()
    render(
      <MemoryRouter>
        <Table>
          <TableBody>
            <PersonRow person={person} onEdit={onEdit} />
          </TableBody>
        </Table>
      </MemoryRouter>,
    )
    expect(screen.getByRole('link', { name: 'Alex Nguyen' }).getAttribute('href')).toBe('/people/p1')
    fireEvent.click(screen.getByRole('button', { name: 'Edit' }))
    expect(onEdit).toHaveBeenCalledWith(person)
  })
})
```

`client/src/pages/my-schedule/component.test.tsx`:

```tsx
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { MySchedulePage } from './component'

const { useCurrentUser } = vi.hoisted(() => ({ useCurrentUser: vi.fn() }))
vi.mock('@/shared/auth', () => ({ useCurrentUser }))
vi.mock('@/features/person-calendar', () => ({
  PersonCalendar: ({ personId }: { personId: string }) => personId,
}))
vi.mock('@/features/absences', () => ({ AbsenceList: () => null }))

afterEach(() => {
  cleanup()
  useCurrentUser.mockReset()
})

describe('MySchedulePage', () => {
  it('shows the calendar of the signed-in person', () => {
    useCurrentUser.mockReturnValue({ data: { role: 'specialist', person_id: 'u1' }, isPending: false })
    render(<MySchedulePage />)
    expect(screen.getByText('u1')).toBeTruthy()
    expect(screen.getByText('My absences')).toBeTruthy()
  })

  it('explains a missing person record', () => {
    useCurrentUser.mockReturnValue({ data: { role: 'specialist', person_id: null }, isPending: false })
    render(<MySchedulePage />)
    expect(screen.getByText('No person record')).toBeTruthy()
  })
})
```

- [ ] **Step 2: Run them to see them fail**

Run: `cd client && npx vitest run src/features/people/components/person-row src/pages/my-schedule`
Expected: FAIL: no link named `Alex Nguyen` (the name is plain text today), and no text `u1` (My schedule shows the placeholder).

- [ ] **Step 3: People rows**

In `client/src/features/people/components/person-row/component.tsx`:
- import `generatePath` from `react-router`, `ROUTES` from `@/resources/routes`, `LinkRow` from `@/shared/components/link-row`, `TextLink` from `@/shared/components/text-link`; drop `TableRow` from the table import
- in the body, before `handleEdit`: `const personPath = generatePath(ROUTES.personDetail, { id: person.id })`
- replace `<TableRow>` … `</TableRow>` with `<LinkRow to={personPath}>` … `</LinkRow>`
- replace `<span className="font-medium">{person.name}</span>` with `<TextLink to={personPath}>{person.name}</TextLink>`

The Absences and Edit buttons keep their own actions: `LinkRow` ignores clicks inside buttons and links.

- [ ] **Step 4: My schedule**

Replace `client/src/pages/my-schedule/component.tsx`:

```tsx
import { AbsenceList } from '@/features/absences'
import { PersonCalendar } from '@/features/person-calendar'
import { useCurrentUser } from '@/shared/auth'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'

export function MySchedulePage() {
  const { t } = useTranslation()
  const { data: user, isPending } = useCurrentUser()

  if (isPending) return null

  return (
    <div className="flex flex-col gap-8">
      <Section variant="page" title={t('My schedule')} description={t('Your hours by week, leave and holidays.')} />
      {user?.person_id ? (
        <>
          <PersonCalendar personId={user.person_id} />
          <Section title={t('My absences')} description={t('Report leave or training. A planner approves it.')}>
            <AbsenceList personId={user.person_id} canApprove={false} />
          </Section>
        </>
      ) : (
        <PagePlaceholder
          title={t('No person record')}
          description={t('Your account is not linked to a person yet. Ask a planner to add you with the same email.')}
        />
      )}
    </div>
  )
}
```

In `language-key.ts`: add `'Your hours by week, leave and holidays.': string` in its place, and delete `'Your assignments by week appear here once a staffing draft that includes you is accepted.': string` after checking it has no other user: `grep -rn "Your assignments by week appear" client/src` should list only `language-key.ts`.

- [ ] **Step 5: Run the tests to see them pass, then all checks**

Run: `cd client && npx vitest run src/features/people/components/person-row src/pages/my-schedule`
Expected: PASS, 3 tests.

Run: `cd client && npm run format && npx tsc -b && npm run lint && npm test && npm run build`
Expected: all green.

- [ ] **Step 6: Commit**

```bash
git add client/src/features/people/components/person-row client/src/pages/my-schedule client/src/shared/i18n/language-key.ts
git commit -m "feat(client): people rows open the person; My schedule shows the calendar

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 12: Check it end to end, update the docs, open the PR

**Files:**
- Modify: `docs/features/member-calendar.md`, `docs/features/README.md`, `server/README.md`

- [ ] **Step 1: Check on the real stack**

Run: `make dev` (or `make infra`, then `make server` and `make client`), then `make seed`.
1. Sign in as `planner` and open People: click Alex's row. Expected: `/people/<id>` with "People › Alex …", this month's grid, and the weeks' load matching the dashboard heatmap for the same weeks.
2. Click a week's load. Expected: its projects with hours; each name opens the project page.
3. Click **Next month** for a month with six ISO weeks (November 2026). Expected: six rows.
4. Add planned leave for Alex in the absences section. Expected: the calendar shows "Leave (planned)" on those days without a reload, and the week turns tentative.
5. Sign in as `specialist` and open My schedule. Expected: the same calendar for their own person.
6. As `specialist`, call `GET /api/v1/people/<someone else>/schedule?from=2026-W42&to=2026-W43` from the browser console with the session token. Expected: `403`.
7. As `planner`, on Alex's page click **Approve** on the planned leave from step 4. Expected: the badge turns approved in the list and on the calendar, the week is no longer tentative, and Events lists a new `AbsenceAdded`.
8. As `manager`, open Alex's page. Expected: calendar only, no absence list, and "People" in the breadcrumb is plain text.

Stop the dev servers you started; leave any server the user already had running.

- [ ] **Step 2: Push and open the PR**

```bash
git push -u origin feat/member-calendar
gh pr create --base main --head feat/member-calendar --title "Member calendar" --body "<what changed, the checks run, the As built notes from Step 3>

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```

Note the PR number it prints.

- [ ] **Step 3: Update the docs**

`docs/features/member-calendar.md`:
- status line: `**Status:** Done (PR #<number>)`
- tick every task
- add an **As built** list under "Not in this feature":
  - The schedule types live in `models/schedule.go` with aliases in `services`, because `dtos` cannot import `services`.
  - Approving records `AbsenceAdded` with the new status; there is no separate event type. Approving twice or a holiday gets `409 ABSENCE_NOT_PLANNED`.
  - The band classes and labels moved to `shared/utilization`; person and absence types, their labels and `PLANNED_ABSENCE_STATUS` to `shared/people`.
  - The breadcrumb `Trail` shows a middle step without a link as plain text (used for "People" when the user cannot open the list).

`docs/features/README.md`: set row 11 (Member calendar) to **Done**.

`server/README.md` already lists `GET /people/{id}`, the schedule and the approve rows (added with the spec). Check that they match what was built and fix any wording that drifted.

- [ ] **Step 4: Commit and push the docs**

```bash
git add docs/features/member-calendar.md docs/features/README.md server/README.md  # server/README only if it needed a fix
git commit -m "docs: member calendar done

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git push
```

Wait for CI (`gh pr checks <number> --watch`). Do not merge: the user says when.
