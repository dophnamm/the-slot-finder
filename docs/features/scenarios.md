# Scenarios

**Goal:** a planner tries a what-if, such as "Checkout Revamp slips two weeks" or "Alex is on leave from 26 Oct to 6 Nov". They see who would gain or lose hours, which demand would go unmet and how much the plan would change, without touching the baseline or any real data. This is the **Scenario** of the [root README](../../README.md#6-data-model): a `PLAN` of kind `scenario` with `assumptions`.

**Roles:** planner (create, view, run again, delete).
**Target:** 12–13 Oct · **Depends on:** [staffing-draft](staffing-draft.md), [replan-on-change](replan-on-change.md) · **Branch:** `feat/scenarios` · **Status:** Done (PR #24)

## In scope

- **Assumptions.** A scenario holds one or more of these, combined:
  - `shift_project`: a project's start, deadline and phases move by N weeks (−`SCENARIO_MAX_SHIFT_WEEKS` to +`SCENARIO_MAX_SHIFT_WEEKS`, not 0)
  - `add_absence`: a person is away from one date to another for some hours per day, counted like approved leave
- **Computation.** The server computes a scenario and only writes the scenario itself:
  1. Load the baseline, then apply the assumptions in memory. Shifted projects get new dates and phase weeks, and the extra absence lowers that person's capacity.
  2. Pick the projects to plan again: the shifted ones, plus every project whose baseline hours no longer fit someone's capacity. This is the same check the re-plan runner uses.
  3. Run the engine on those projects with the re-plan stability weight, keeping the rest of the baseline as fixed load.
  4. Save the result as a `PLAN` of kind `scenario` with its name, its `assumptions` and the baseline version it was built on.
- **Result.** For each scenario the server returns:
  - the changes against the baseline: before and after hours per person and demand
  - plan churn: how many people changed and how many hours moved
  - the unmet demand, with reason codes
  - the projects that were planned again
- **Lifecycle:**
  - list and delete scenarios
  - a scenario is **stale** once the baseline has moved on
  - **Run again** computes it once more, with the same assumptions, against the current baseline
- **Making it real** is outside the scenario. The planner moves the phases on the project page or adds the leave, and the existing re-plan flow then creates a staffing draft. The scenario page links to the project and to the person's absences.

**As built**
- **A shifted project keeps its people.** Before planning, its baseline hours move by the same weeks. The engine then re-plans only where the new weeks do not fit, such as a holiday week. Without this, hours outside the new window became a fresh seat for someone else.
- **The diff compares week by week.** Hours are compared per person and demand for each week, so a project that only moves still shows up.
  - The same hours in other weeks show as **Moved**.
  - **Hours moved** counts the hours placed in a week beyond what that week had in the baseline.
  - **People changed** counts the distinct people with any change.
- The migration also adds `summary` (jsonb: people changed, hours moved, unmet hours), so the list does not recompute every scenario.
- An unstaffed (incoming) project that is shifted is planned the way a staffing draft would plan it. This shows who could take it at the new dates.
- The scenario page has its own changes table, because it shows weeks. The draft review keeps its table. The change badge colours moved to `shared/staffing`.
- The projects planned again come with their program id, so the page links straight to each project.
- The staffing-draft routes answer `404` for a scenario id (`findDraftPlan`).
- Checked on the real stack, with Checkout Revamp accepted into the baseline:
  - "+2 weeks": four people moved two weeks, and 9 QA hours went to Hoa in the Christmas week.
  - "Alex away 26 Oct – 6 Nov": Chi took Alex's 64 hours.

## Not now

- Turning a scenario straight into a staffing draft. That draft would plan for changes that did not happen. → [Later](README.md#later)
- More assumption types: a new project, a planned hire (a future `start_date`), a demand change → [Later](README.md#later).
- The AI what-if chat → [Later](README.md#later).
- Scenarios for managers.

## Server

| Method | Path | Purpose | Roles |
|---|---|---|---|
| POST | `/scenarios` | Create and compute: `{ name, assumptions: [{ type, ... }] }` | planner |
| GET | `/scenarios` | List, newest first: name, base version, stale flag, churn | planner |
| GET | `/scenarios/{id}` | Assumptions, changes against the baseline, churn, unmet, projects planned again | planner |
| POST | `/scenarios/{id}/rerun` | Compute again against the current baseline | planner |
| DELETE | `/scenarios/{id}` | Delete the scenario | planner |

The [server README](../../server/README.md#3-api-restjson) lists `GET /scenarios/{id}/diff`. This spec puts the diff into `GET /scenarios/{id}`, as for staffing drafts, and updates that table.

**Rules**
- **Migration:** `plans` gets `name` (text, null for other kinds) and `assumptions` (jsonb, default `[]`); new kind `scenario`. One Atlas migration: `make migrate-generate name=scenarios`.
- **Validation:**
  - the project exists and is not done or cancelled
  - the shift is within the limit and not 0
  - the person exists and from ≤ to
  - hours per day follow the absence rules
  - at most `SCENARIO_MAX_ASSUMPTIONS` assumptions
- **Settings:** `SCENARIO_MAX_SHIFT_WEEKS` (default 12) and `SCENARIO_MAX_ASSUMPTIONS` (default 10) in `cmd/server/.env.example` and [server/README §8](../../server/README.md#8-settings).
- **No side effects:** no change events, no audit record, no writes outside the scenario's own plan. The staffing-draft routes do not find a scenario, so it can never be accepted.
- **Code:** a `ScenarioService` (`services/scenario.go`) reuses the draft builder (its snapshot takes the extra absences), `conflictingProjects`, `buildDraft` and `diffPlans`.

## Client

- `/scenarios`:
  - a bordered table: name, assumptions in short, baseline version (with a "Stale" badge), people changed, hours moved, created
  - a **New scenario** button
- **New scenario** dialog:
  - a name
  - assumption rows, added and removed one at a time; each row is either **Move a project** (project, weeks) or **Add leave** (person, from, to, hours per day)
  - **Create** opens the scenario
- `/scenarios/:id`:
  - a page header: name, "Stale" badge, **Run again**, **Delete**
  - the assumptions
  - the figures: people changed, hours moved, unmet hours
  - the changes table
  - the unmet demand
  - a "Make it real" note linking to the project or the person
- Breadcrumbs: Scenarios › name. Each scenario row opens its page.
- The changes table moves from `features/draft-review` to `shared/components`, so the draft review and the scenario page share it.

## Acceptance criteria

1. With Checkout Revamp staffed in the baseline, "Checkout Revamp +2 weeks" moves its allocations two weeks later, or puts hours in unmet where nobody is free. The baseline, the project and its phases are unchanged afterwards.
2. With Alex staffed in late October, "Alex on leave 26 Oct – 6 Nov" takes those hours off Alex. Someone else takes them, or they appear in unmet with a reason; another person can only take them if the demand has a free seat (headcount of 2 or more), so with headcount 1 they stay unmet.
3. Two assumptions combine in one scenario.
4. The changes, churn and unmet shown match the scenario's plan against the baseline.
5. After a draft is accepted, the scenario shows "Stale"; Run again computes it against the new baseline.
6. A manager or a specialist gets `403` on every scenario route. `POST /staffing-drafts/{scenarioId}/accept` gets `404`.
7. Deleting a scenario removes its plan, assignments and allocations.

## Tasks

**Server**
- [x] Migration (`name`, `assumptions`), constants, settings
- [x] Assumption validation and in-memory application (shift a project, add an absence)
- [x] `ScenarioService`: create, list, get with diff, run again, delete; handlers, DTOs, routes, Swagger
- [x] Tests: assumptions applied in memory, projects picked for planning, no writes outside the scenario, stale flag, RBAC

**Client**
- [x] `shared/scenarios`: types, API, hooks; endpoints, routes and breadcrumbs
- [x] Scenarios list and New scenario dialog
- [x] Scenario page with its own changes table (see As built)
- [x] Component tests for the dialog rows and the scenario page

**Estimate:** about one day (server about 5 hours, client about 4 hours).
