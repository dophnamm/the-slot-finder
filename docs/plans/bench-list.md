# Bench list Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** List the people who stay under the underused band in each of the next `BENCH_WEEKS` weeks, on the dashboard's Risks card, each linking to their person page.

**Architecture:** A pure rule `metrics.Bench` decides who is on the bench from per-week capacity and load. `InsightsService.Bench` reuses the heatmap computation for the caller (same baseline, capacity, people and visibility) over the bench window and feeds each row to the rule. `GET /insights/bench` returns `{ weeks, people }`. The client's risks feature gets a `useBench` hook and a `BenchList` child rendered inside the Risks card.

**Tech Stack:** Go (Echo v5, GORM, testify), React + TypeScript (TanStack Query, Vitest, Testing Library).

**Spec:** [docs/features/bench-list.md](../features/bench-list.md)

## Global Constraints

- No literals: route path, role list, messages and settings are constants or config (`server/internal/constants`, `config.Config`); client endpoint in `src/resources/endpoints`, routes via `ROUTES`, text through `t()` with keys in `client/src/shared/i18n/language-key.ts` sorted alphabetically (same ordering convention as the file: compare case-insensitively, lowercase before uppercase on a tie).
- Same visibility as the heatmap: a planner sees everyone, a manager only the people of their programs' teams (`AuthzService.VisiblePeople`); specialists get `403` (`constants.InsightRoles` = planner, manager).
- Rule (spec): on the bench when, in each of the next `BENCH_WEEKS` weeks starting with the current week, capacity is above 0 and load ÷ capacity is below `UTILIZATION_UNDERUSED_BELOW` (default 0.5). Exactly at the threshold is off. Free hours = Σ (capacity − load) over those weeks. Sorted by free hours, most first, then by name.
- Swagger is regenerated (`make swagger-load`, or `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init` with the flags of the `swagger-load` target in `server/Makefile`), never edited by hand.
- Client component rules from the `structuring-client-components` skill (`component.tsx` holds only imports and one component, no inline function props; children in `components/<child>/`); no `style={...}`.
- Load the matching skills: `writing-go-server-code`, `testing-go-server-code`, `structuring-server-modules` (server); `translating-client-text`, `structuring-client-components`, `organizing-client-shared-code`, `styling-client-components` (client).
- Commits: local only on `feat/bench-list`, never push; never stage `docs/features/README.md` or `docs/features/people-import.md` (they hold another feature's uncommitted work). Server and client never in the same commit. Run `make check-server` / `make check-client` once before each task's last commit. Conventional subject; the message ends with the trailer `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

1. A person on leave for one of the weeks (0 capacity) is not on the bench. Test: Task 1 case "a week with no capacity".
2. Utilization exactly at the threshold (18 h of 36 h with 0.5) is off the bench. Test: Task 1 case "exactly at the threshold".
3. A manager never sees people outside their teams' members. Test: Task 2 case "manager sees only their teams".
4. A person with no baseline hours at all (load 0, capacity > 0 every week) is on the bench with all their capacity free. Test: Task 1 case "no load at all".
5. The dashboard still renders when the bench request fails or nobody is on the bench. Test: Task 4 cases "empty state" and "load error".

---

### Task 1: The bench rule and `BENCH_WEEKS`

**Files:**
- Modify: `server/internal/metrics/metrics.go` (next to `SustainedOverload` ~187-189)
- Test: `server/internal/metrics/metrics_test.go` (new `TestBench`, next to `TestSustainedOverload` ~218)
- Modify: `server/internal/config/config.go` (field next to `SustainedOverloadWeeks` ~227; load next to `getEnvAsInt("SUSTAINED_OVERLOAD_WEEKS", 3)` ~382)
- Modify: `server/cmd/server/.env.example` (next to `SUSTAINED_OVERLOAD_WEEKS` ~134-135)

**Interfaces:**
- Produces:
```go
// BenchWeek is one week of a person's capacity and baseline load, in hours.
type BenchWeek struct {
	Capacity float64
	Load     float64
}

// Bench tells whether a person is on the bench for these weeks (README §7): in every week their
// capacity is above 0 and load ÷ capacity is below underusedBelow. freeHours is Σ (capacity − load).
// No weeks means not on the bench.
func Bench(weeks []BenchWeek, underusedBelow float64) (freeHours float64, onBench bool)
```
- Produces: `config.Config.BenchWeeks int`, env `BENCH_WEEKS`, default 4; `.env.example` gets `# Weeks ahead a person must stay underused to be on the bench list` and `BENCH_WEEKS=4`.

- [ ] **Step 1: Write the failing test** — `TestBench` (table, `underusedBelow` 0.5):
  - "on the bench": four weeks of `{36, 10}` → `104, true`
  - "no load at all": four weeks of `{36, 0}` → `144, true`
  - "one week at or above the band": `{36,10},{36,10},{36,20},{36,10}` → `_, false`
  - "exactly at the threshold": `{36,10},{36,18},{36,10},{36,10}` → `_, false`
  - "a week with no capacity": `{36,10},{0,0},{36,10},{36,10}` → `_, false`
  - "no weeks": `nil` → `0, false`
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/metrics -run TestBench -v` (undefined `Bench`)
- [ ] **Step 3: Implement** `Bench` and the `BenchWeeks` setting as above.
- [ ] **Step 4: Run, expect PASS** — `cd server && go test ./internal/metrics ./internal/config`
- [ ] **Step 5: Commit** `feat(server): add the bench rule and BENCH_WEEKS`

### Task 2: `InsightsService.Bench`

**Files:**
- Modify: `server/internal/models/insights.go` (next to `Heatmap` ~10)
- Modify: `server/internal/services/insights.go` (interface ~44-65; new method after `Heatmap` ~90-104)
- Test: `server/internal/services/insights_test.go` (new `TestInsightsService_Bench`, next to `TestInsightsService_Heatmap_ManagerSeesTheirTeams` ~190; reuse `heatmapFixture`, `heatmapConfig`, `newInsightsService`, `heatmapPerson`, `skill`)

**Interfaces:**
- Consumes: `metrics.Bench`, `metrics.BenchWeek`, `config.Config.BenchWeeks` (Task 1); existing `s.authz.Require(ctx, CanViewInsights)`, `s.authz.VisiblePeople(ctx)`, `s.heatmap(ctx, HeatmapInput, PeopleScope)`, `isoweek.Of(s.clock.Now())`, `models.HeatmapRow{Person, PrimaryExpertise, Cells}`, `models.HeatmapCell{Capacity, Load}`.
- Produces:
```go
// Bench is the people idle for the next Weeks weeks.
type Bench struct {
	Weeks  int
	People []BenchPerson
}

// BenchPerson is one person on the bench and their free hours over the window.
type BenchPerson struct {
	Person           Person
	PrimaryExpertise Expertise
	FreeHours        float64
}
```
and on `InsightsService`: `Bench(ctx context.Context) (*models.Bench, error)`.

- [ ] **Step 1: Write the failing test** — `TestInsightsService_Bench` (subtests, fixture clock in a fixed week):
  - "lists the underused people with their free hours": a person with 36 h capacity and 10 h baseline load in each of the next 4 weeks is listed with `FreeHours` 104; a person at 30 h load is not listed; `Weeks` is 4
  - "sorted by free hours, then name": two people on the bench with 104 h and 144 h free → the 144 h one first; two with equal free hours → by name
  - "manager sees only their teams": as `TestInsightsService_Heatmap_ManagerSeesTheirTeams` — an underused person outside the manager's teams is not listed
  - "specialist is refused": `CanViewInsights` fails → error, as the heatmap tests check
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/services -run TestInsightsService_Bench -v`
- [ ] **Step 3: Implement** — `Require(ctx, CanViewInsights)`; `scope := VisiblePeople(ctx)`; `from := isoweek.Of(s.clock.Now())`, `to := from.AddWeeks(weeks - 1)` with `weeks := max(s.cfg.BenchWeeks, 1)`; `heatmap, err := s.heatmap(ctx, HeatmapInput{From: &from, To: &to}, scope)`; for each row build `[]metrics.BenchWeek` from its cells and call `metrics.Bench(cells, s.cfg.UtilizationUnderusedBelow)`; keep the ones on the bench; sort by `FreeHours` desc, then `Person.Name`, then `Person.ID`. Return `&models.Bench{Weeks: weeks, People: ...}` (an empty slice, never nil).
- [ ] **Step 4: Run, expect PASS** — `cd server && go test ./internal/services`
- [ ] **Step 5: Commit** `feat(server): compute the bench list from the heatmap`

### Task 3: `GET /insights/bench`

**Files:**
- Modify: `server/internal/constants/routes.go` (`PathBench = "/insights/bench"` next to `PathCapacityForecast` ~56-59)
- Modify: `server/internal/constants/messages.go` (`MsgBenchRetrieved = "Bench retrieved"`, next to `MsgCapacityForecastRetrieved` ~158)
- Modify: `server/internal/dtos/insights.go` (response + mapper)
- Modify: `server/internal/handlers/insights.go` (new `Bench` handler, after `CapacityForecast` ~137)
- Modify: `server/cmd/server/routes/router.go` (`registerInsightRoutes` ~178-186: `v1.GET(constants.PathBench, insightsHandler.Bench, authenticated, roles)`)
- Regenerate: `server/docs`
- Test: `server/internal/dtos/insights_test.go` (new `TestNewBenchResponse`), `server/cmd/server/routes/router_test.go` (RBAC rows in `TestRouter_People_RBAC` ~244)

**Interfaces:**
- Consumes: `InsightsService.Bench`, `models.Bench` (Task 2).
- Produces (client relies on these json keys):
```go
// BenchResponse is GET /insights/bench.
type BenchResponse struct {
	Weeks  int                   `json:"weeks"`
	People []BenchPersonResponse `json:"people"`
}

// BenchPersonResponse is one person on the bench.
type BenchPersonResponse struct {
	PersonID         string  `json:"person_id"`
	Name             string  `json:"name"`
	PrimaryExpertise string  `json:"primary_expertise"`
	FreeHours        float64 `json:"free_hours"`
}

func NewBenchResponse(b models.Bench) BenchResponse
```
`PrimaryExpertise` is the expertise name; `People` is `[]` (never `null`) when empty.

- [ ] **Step 1: Write the failing tests** — `TestNewBenchResponse`: one person → `{weeks: 4, people: [{person_id, name: "Quinn", primary_expertise: "QA", free_hours: 104}]}`; empty bench → `People` is an empty, non-nil slice. RBAC rows: specialist `GET constants.PathBench` → `403`; no-role token → `403`; manager → not `403`.
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/dtos -run TestNewBenchResponse -v && go test ./cmd/server/routes -run TestRouter_People_RBAC -v`
- [ ] **Step 3: Implement** — DTO and mapper; handler `Bench` with a godoc in the style of `CapacityForecast` (`@Summary Bench list`, `@Description People under UTILIZATION_UNDERUSED_BELOW in each of the next BENCH_WEEKS weeks, with their free hours, most free first. A manager sees the people of their teams.`, `@Tags Insights`, `@Success 200 {object} object{meta=dtos.Meta,data=dtos.BenchResponse}`, `@Failure 403`, `@Router /insights/bench [get]`), responding with `constants.MsgBenchRetrieved`; route; constants; regenerate Swagger and check `git diff --stat server/docs` shows only the new path and types.
- [ ] **Step 4: Run, expect PASS** — `make check-server` from the repo root
- [ ] **Step 5: Commit** `feat(server): add GET /insights/bench`

### Task 4: Bench list on the Risks card (client)

**Files:**
- Modify: `client/src/resources/endpoints/insights.ts` (`bench: '/insights/bench'`)
- Modify: `client/src/features/risks/api.ts` (`getBench()`), `types.ts` (`BenchPerson`, `BenchData`), `constants.ts` (`BENCH_QUERY_NAME = 'bench'`)
- Create: `client/src/features/risks/use-bench.ts`
- Create: `client/src/features/risks/components/bench-list/` (`component.tsx`, `index.ts`, `types.ts` if it takes props)
- Modify: `client/src/features/risks/component.tsx` (render `<BenchList />` after the sustained-overload section)
- Modify: `client/src/shared/i18n/language-key.ts`
- Test: `client/src/features/risks/components/bench-list/component.test.tsx` (create); `client/src/features/risks/component.test.tsx` (add `vi.mock('./use-bench', ...)` so the existing tests keep passing)

**Interfaces:**
- Consumes: `GET /insights/bench` → `{ weeks: number; people: { person_id: string; name: string; primary_expertise: string; free_hours: number }[] }` (Task 3).
- Produces: `useBench()` = `useQuery({ queryKey: [...INSIGHTS_QUERY_KEY, BENCH_QUERY_NAME], queryFn: getBench })` (refreshed with the other insights when a draft is accepted).

- [ ] **Step 1: Write the failing test** (`bench-list/component.test.tsx`, `vi.mock` of `../../use-bench`, wrapped in `MemoryRouter`, like `features/risks/component.test.tsx`):
  - "lists people with their free hours": data `{ weeks: 4, people: [{ person_id: 'p1', name: 'Quinn', primary_expertise: 'QA', free_hours: 104 }] }` → heading `Bench (1)`; a link named `Quinn` with `href` `/people/p1`; text `QA`; text `104 h free in 4 weeks`
  - "empty state": `{ weeks: 4, people: [] }` → text `Nobody is on the bench`
  - "load error": the hook returns an error → the existing `LoadError` renders and nothing throws
- [ ] **Step 2: Run, expect FAIL** — `cd client && npx vitest run src/features/risks`
- [ ] **Step 3: Implement**
  - Keys: `'Bench ({count})'`, `'{hours} h free in {weeks} weeks'`, `'Nobody is on the bench'`.
  - Each row: `TextLink` to `generatePath(ROUTES.personDetail, { id: person.person_id })` with the name, then the primary expertise, then `t('{hours} h free in {weeks} weeks', { hours, weeks })`; format hours with the existing hours helper in `@/utils/functions` if there is one, otherwise round to a whole number.
  - Same section markup and classes as the sustained-overload list in `features/risks/component.tsx` (`h3` `text-sm font-medium`, `ul` `divide-y`); loading uses `Skeleton` rows like the Risks card.
- [ ] **Step 4: Run, expect PASS** — `cd client && npx vitest run src/features/risks && npm run lint && npx tsc -b`, then `make check-client` from the repo root
- [ ] **Step 5: Commit** `feat(client): show the bench list on the Risks card`

### Task 5: Docs

Written by the docs agent from the diff of Tasks 1–4, reviewed before commit. Server and root docs in one commit; a separate `docs(client): ...` commit only if `client/README.md` changes.

- [ ] `README.md` §7: the "Bench list" row is built (drop "(later)" and the suggested actions); any `BENCH_WEEKS` / bench mention elsewhere
- [ ] `server/README.md`: move `GET /insights/bench` out of the Later routes into the routes table; the "Team signals" row (bench built); config table gets `BENCH_WEEKS` (default 4) and it leaves the "Later settings" list
- [ ] `client/README.md`: the Risks card text mentions the bench list; remove "`/insights`: bench list" from Later
- [ ] `docs/features/team-heatmap.md`: drop the bench bullet under Not now
- [ ] `docs/features/README.md`: row 19 → Done; remove "Bench list" from the "## Later" list (commit only those lines: the file has another feature's uncommitted row)
- [ ] `docs/features/bench-list.md`: status Done, tasks ticked
- [ ] `docs/user-manual.md`: the dashboard section describes the bench list
- [ ] `README.md:3` and `CLAUDE.md:9`: nineteen features built (sign-in through bench list)
- [ ] Commit `docs: describe the bench list`

### Task 6: Final check

- [ ] `make check` from the repo root passes
- [ ] `git log --oneline feat/skill-and-program-events..HEAD` shows only this feature's commits; nothing pushed
