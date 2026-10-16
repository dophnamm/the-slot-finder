# Staffing draft

**Goal:** the Planner gets a squad for incoming projects in seconds, with a reason for every choice, edits it, and accepts it into the baseline. This covers G1 in the [root README](../../README.md#1-problem-and-scope).

**Roles:** planner (generate, edit, accept, reject), lead (view drafts that touch their programs).
**Target:** 12–13 Oct · **Depends on:** [programs-and-demand](programs-and-demand.md) · **Branch:** `feat/staffing-draft` · **Status:** Done (PR #8)

## In scope

- PLAN (`baseline` and `draft`), ASSIGNMENT, ALLOCATION.
- `internal/engine` (standard library only) with the **greedy** strategy:
  - **Hard constraints:** expertise at ≥ minimum level · weekly hours ≤ capacity × `maxUtilization` · at most `maxExpertisesPerPersonWeek` expertises per person per week · no allocation below `minHoursPerWeek` except a demand's last remainder · only inside the demand's window · locked allocations untouched · never a stretch assignment
  - **Soft objectives:** priority and deadline order · primary-expertise match · load near `targetUtilization` · continuity on the same project · level close to the demand's minimum level (soft: every level above the minimum lowers the score, nobody is excluded)
  - **Layout:** `flat` profile, and `headcount` seats with one person per seat when possible
  - **Output:** a reason on every assignment, and `unmet` with a reason code
- Draft generation for chosen projects, review, edit (add or remove a person, change weekly hours, lock), stretch by hand (flagged, confirmed on accept), reject.
- **Atomic accept** into the baseline.
- On accept: covered projects move to `planned` and get `staffed_at` (time-to-staff for G1).

## Not now

- Skill tags, preferred person, alternatives and their swap → [Later](README.md#later), design in [root README §5](../../README.md#5-planning-engine-the-core).
- Profiles `front` and `back`, and the repair pass → [Later](README.md#later).
- What-if scenarios and plan churn → [Later](README.md#later).
- Re-plan drafts from change events → [replan-on-change](replan-on-change.md).

## Server

**Endpoints** (subset of [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/plans/current` | The baseline | planner, lead |
| POST | `/staffing-drafts` | Generate a draft: `{ projectIds }` | planner |
| GET | `/staffing-drafts/{id}` | Draft with diff against the baseline, reasons and unmet | planner, lead |
| PATCH | `/staffing-drafts/{id}/assignments` | Add or remove an assignment, change weekly hours, lock or unlock | planner |
| POST | `/staffing-drafts/{id}/accept` | Publish (`{ baseVersion }`) | planner |
| POST | `/staffing-drafts/{id}/reject` | Discard | planner |
| GET | `/people/{id}/assignments?from=&to=` | A person's baseline assignments | planner, lead, the person themselves |

**Tables and migration:** `plan`, `assignment`, `allocation`, plus a partial unique index. One Atlas migration (`make migrate-generate name=staffing_draft`).

**Atomic accept**, in one `TxManager.Do`:
1. Supersede the current baseline only if its version is still the one the Planner saw:
   `UPDATE plan SET status = 'superseded' WHERE kind = 'baseline' AND status = 'published' AND version = :baseVersion`.
2. **0 rows updated → `409 PLAN_VERSION_CONFLICT`.** Nothing is written, and the Client offers to regenerate the draft.
3. Publish the draft as the new baseline with `version = baseVersion + 1`, set the projects' status and `staffed_at`, and write the audit record.
4. A partial unique index allows only one published baseline:
   `CREATE UNIQUE INDEX plan_one_published_baseline ON plan (kind) WHERE kind = 'baseline' AND status = 'published'`.
   If two accepts race past step 1, the second commit fails on this index and also returns `409`.

At startup the server creates an empty published baseline (version 0) if none exists. The unique index makes this safe to repeat, and the first accept follows the same path as every other.

**Rules**
- The planning service builds the engine snapshot: members' capacity from availability, demand windows from phases (or project start to deadline), and the baseline as `existing`.
- `ASSIGNMENT.kind` is set by the server: `primary`, `cross_expertise`, or `stretch` (by hand only, locked, confirmed on accept).
- Policy values come from config (`PLANNING_*` in [server/README §8](../../server/README.md#8-settings)).
- Write payloads never carry `kind`, `status` or computed hours totals.

## Engine

- Sort demand by priority, then deadline, then scarcity of the expertise.
- For each seat (`hours ÷ headcount`), score the candidates who pass every hard constraint. Take the best, and spread its hours evenly over the window (`flat`), never below `minHoursPerWeek` except the last remainder.
- When nobody has enough capacity for a whole seat, split the seat between two people and say so in the reason.
- Record anything left in `unmet` with `NO_SKILL`, `LEVEL_TOO_LOW`, `NO_CAPACITY_BEFORE_DEADLINE` or `BLOCKED_BY_LOCKED`. Never panic.
- The same input always gives the same output: ties break by person id.

**As built**
- Candidates rank by: holds the whole seat, then primary-expertise match, then the weighted score (deadline coverage, primary fit, balance, stability, continuity, levelFit), then person id. So a primary person wins whenever one can hold the seat, and a cross-expertise person wins when no primary person can.
- The planning service passes each demand's window as its list of ISO weeks, so the engine does no week arithmetic. `PLANNING_HOUR_STEP` (default 1) rounds allocations to whole hours.
- A demand never gets more distinct people than its `headcount`: once it has that many, only people already on it can take more hours, and what they cannot hold stays unmet (`NO_CAPACITY_BEFORE_DEADLINE`). With `headcount` 1, a week the only holder cannot work (leave, or a re-plan) gets nobody else, so those hours stay unmet until the Planner raises the headcount or adds a person by hand, whereas `headcount` 2 leaves a free seat for the other person.
- Headcount is lowered when a seat would be smaller than one minimum week. Per demand, at most one allocation is below `minHoursPerWeek`: the last remainder.
- `BLOCKED_BY_LOCKED` when the locked hours of qualified people inside the window are at least the missing hours; otherwise `NO_CAPACITY_BEFORE_DEADLINE`.
- Existing assignments: for demand outside the request they are fixed load; for requested demand, locked allocations are kept and unlocked ones are planned again, preferring the same person (stability).
- A draft is a full copy of the baseline with the covered demand planned again. Accept publishes that plan as the baseline, `version = baseVersion + 1`, and also rejects a draft whose own base version differs from `baseVersion`.
- "Covered projects move to `planned`" means incoming projects whose every demand gets all its hours. Partly staffed projects stay `incoming`.
- Stretch confirmation covers stretch on the draft's own demand only; stretch copied from the baseline was confirmed when it was accepted.
- Draft edits are one batch: `{ add, remove, set }`; `set` with 0 hours removes the week. Adding a person creates an assignment without hours; the Planner then types hours in the grid.
- The audit record is an `audit_events` row (actor, action `plan.accepted`, versions before and after, planned projects).
- `GET /staffing-drafts/{id}` lists changes against the baseline in force (added, removed, changed hours per person and demand). Conflicts and the score are not returned yet.
- Client: `shared/staffing` (types, API, hooks), `shared/components/propose-squad-button`, `features/draft-review`. A `409 PLAN_VERSION_CONFLICT` on accept is shown on the page, not as a toast.

## Client

- `/intake` and `/programs/:id`: **Propose squad** calls `POST /staffing-drafts` and opens the draft.
- `/staffing-drafts/:id`:
  - assignment grid (person × week, grouped by assignment) with reasons and kind, and the unmet list
  - diff against the baseline
  - edit hours, add a person, lock
  - accept with confirmation (and an extra one for stretch), reject
  - on `409 PLAN_VERSION_CONFLICT`, offer to regenerate
- Feature folder: `features/draft-review`.

## Acceptance criteria

1. On the seed data, `POST /staffing-drafts` for one incoming project returns a draft in under 2 seconds, with a reason on every assignment.
2. The engine never exceeds capacity × `maxUtilization`, never goes below the minimum level, never schedules outside the demand window, never changes a locked allocation, and never proposes stretch (property tests).
3. No weekly allocation is below `minHoursPerWeek`, except a demand's last remainder.
4. A primary-expertise person is chosen over a cross-expertise one when both fit.
5. Demand nobody can cover appears in `unmet` with a reason code.
6. Accepting with a stale `baseVersion` returns `409`, and the baseline is unchanged.
7. Two accepts at the same time: exactly one wins and the other gets `409` (integration test).
8. After accepting, the projects are `planned`, `staffed_at` is set, and `/plans/current` shows the new version.
9. Adding a below-level person by hand creates a locked `stretch` assignment, and accept asks for confirmation.

## Tasks

**Server**
- [x] `internal/engine`: types, greedy strategy, reasons, unmet codes
- [x] Snapshot builder in the planning service
- [x] Plans, assignments, allocations: models, repositories, services, handlers, DTOs
- [x] Atomic accept with the conditional UPDATE and the partial unique index; audit record
- [x] Draft edits, kind rules, stretch confirmation, reject
- [x] Migration, constants, Swagger

**Client**
- [x] Propose squad on intake and the program page
- [x] Draft review: grid, reasons, unmet, diff, edit, lock, accept, reject, 409 handling

**Tests**
- [x] Engine golden tests on fixed snapshots, and property tests for every hard constraint
- [x] Accept: stale version, concurrent accepts (integration), unique index
- [x] Service tests for kind and stretch rules
