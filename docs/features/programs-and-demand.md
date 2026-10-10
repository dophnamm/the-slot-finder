# Programs and demand

**Goal:** leads describe what each program needs: its projects, their ordered phases, and demand per expertise. This is the demand side for staffing (G1) in the [root README](../../README.md#1-problem-and-scope). A phase that ends later moves the phases after it.

**Roles:** lead (own programs and projects), planner (all), specialist (none).
**Target:** 12 Oct · **Depends on:** [people-and-skills](people-and-skills.md) · **Branch:** `feat/programs-and-demand`

## In scope

- PROGRAM:
  - name, client, start and end dates, goal, objectives, lead
  - status changed only through actions (`hold`, `resume`, `start-closing`, `close`, `cancel`)
- PROJECT:
  - priority, start_earliest, deadline, scope and done points, lead, `created_at`
  - status `incoming` / `on_hold` / `done` / `cancelled` through actions; `planned` and `active` arrive with staffing-draft
- PROJECT_PHASE: ordered phases, each with a start and an end week, and **slip propagation**.
- DEMAND: expertise, minimum level, headcount, budget hours, optional phase, profile **`flat` only**.
- **Transaction manager**, first needed here: slip propagation writes several phases at once.
- **`week_start date` columns**, first needed here: phases are the first stored weeks.
- `authz` gains per-program checks: `CanEditProgram(programID)` and `CanEditProject(projectID)` (planner, or the lead of that program or project).

## Not now

- Profiles `front` and `back` → [Later](README.md#later), design in [root README §5](../../README.md#5-planning-engine-the-core).
- `confidence` and the pipeline forecast → [Later](README.md#later).
- Forecast override → [program-metrics](program-metrics.md) keeps forecast = planned.
- Milestones and program assistants → [Later](README.md#later).
- Recording `PhaseSlipped` as a change event → [replan-on-change](replan-on-change.md). It adds the outbox insert to the same transaction.

## Server

**Endpoints** (subset of [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET, POST | `/programs` | List or create programs | planner, lead |
| GET, PATCH | `/programs/{id}` | Program record; edit it | GET planner, lead · PATCH planner, lead (own) |
| POST | `/programs/{id}/actions/{action}` | Change program state | planner, lead (own) |
| GET, POST | `/programs/{id}/projects` | List or add projects | planner, lead (own for POST) |
| GET | `/projects?status=` | Projects across programs, e.g. the intake queue | planner, lead |
| GET, PATCH | `/projects/{id}` | Project record; edit it, including points | planner, lead (own) |
| POST | `/projects/{id}/actions/{action}` | Change project state | planner, lead (own) |
| GET, PUT | `/projects/{id}/phases` | Ordered phases (PUT replaces the list) | planner, lead (own) |
| PATCH | `/phases/{id}` | Edit one phase; a later end week shifts later phases | planner, lead (own) |
| PUT | `/projects/{id}/demands` | Replace the project's demand | planner, lead (own) |

**Tables and migration:** `program`, `project`, `project_phase`, `demand`. One Atlas migration (`make migrate-generate name=programs_and_demand`).

**Transactions (`TxManager`)**
- `TxManager.Do(ctx, fn)` begins a transaction, stores it in `ctx`, runs `fn(ctx)`, and commits, or rolls back on error or panic. A nested `Do` reuses the open transaction.
- Repositories get their handle with `dbFrom(ctx)`. It returns the transaction in `ctx`, or the plain DB. So repository methods keep the same signature inside and outside a transaction.
- Services own the transaction boundary. Handlers and repositories never begin one.

**Weeks (`isoweek.Week`)**
- The API speaks ISO week strings (`2026-W42`). Go code uses `isoweek.Week`.
- The database stores the week's Monday in a `date` column named `*_week_start`, e.g. `start_week_start`, `end_week_start`. Range queries and ordering are plain date comparisons.
- A GORM `Scanner` / `Valuer` on `isoweek.Week` maps the type to and from that column.

**Rules**
- Phases have a unique `sequence` per project. `end_week ≥ start_week`. A phase may overlap the previous one.
- **Slip propagation:**
  - moving a phase's end later by N weeks moves every later phase of the project by N weeks, in one `TxManager.Do`
  - moving an end earlier moves nothing else
  - the response returns all changed phases
- A project whose last phase ends after its deadline is returned with `at_risk: true`. The deadline is never moved automatically.
- A demand without a phase spans the project's start week through its deadline week.
- A demand with a profile other than `flat` gets a validation error.
- Program and project status change only through actions. Payloads that carry `status` are rejected.

## Client

- `/programs`: list with status and timeline.
- `/programs/:id`:
  - header (name, status, timeline) and actions
  - Projects tab: rows with status and the current phase
  - "add project"
- `/projects/:id`:
  - project form
  - phase timeline editor, which previews the shift of later phases before saving
  - demand form (expertise, minimum level, headcount, budget hours, phase or "whole project")
- `/intake`: incoming projects across programs.
- Feature folders: `features/programs`, `features/program-overview` and `features/program-projects` (replace `mock-data.ts` with API data), `features/project-phases`, `features/demand-form`, `features/intake`.

## Acceptance criteria

1. With Build W42–W47 and Test W46–W51, moving Build's end to W49 gives Test W48–W53. Both changes are saved together, and the response lists both phases.
2. Moving Build's end earlier, to W46, leaves Test as it was.
3. If one phase update fails during propagation, no phase changes.
4. A project whose last phase ends after the deadline shows `at_risk: true`.
5. A demand without a phase covers the project's start week through its deadline week.
6. The lead of program A gets `403` editing program B. A planner can edit both.
7. `PATCH /programs/{id}` with `status` in the body is rejected. `POST …/actions/hold` changes it.
8. A demand with profile `front` gets a validation error.
9. Weeks round-trip: `2026-W53` is stored as 2026-12-28 and returned as `2026-W53`.

## Tasks

**Server**
- [ ] `TxManager` (`Do`, `dbFrom`) in `internal/db`, provided through FX; update the writing-go-server-code skill (multi-step writes use `TxManager.Do`)
- [ ] `isoweek.Week` GORM `Scanner` / `Valuer` for `date` columns
- [ ] Models, repositories, services, handlers, DTOs for programs, projects, phases and demand
- [ ] Slip propagation in the programs service
- [ ] Actions and lifecycle rules; `authz` per-program checks
- [ ] Migration, constants, Swagger

**Client**
- [ ] Programs list, program page with Projects tab, add project
- [ ] Project page: form, phase timeline editor with shift preview, demand form
- [ ] Intake queue

**Tests**
- [ ] Slip propagation: later shift, earlier no shift, rollback on failure (integration test against Postgres)
- [ ] `TxManager`: commit, rollback on error, rollback on panic, nested reuse
- [ ] RBAC: lead own vs other program
- [ ] Week round-trip through the database
