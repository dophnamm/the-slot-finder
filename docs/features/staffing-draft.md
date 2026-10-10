# Staffing draft

**Goal:** the Planner gets a squad for incoming projects in seconds, with a reason for every choice, edits it, and accepts it into the baseline. This covers G1 in the [root README](../../README.md#1-problem-and-scope).

**Roles:** planner (generate, edit, accept, reject), lead (view drafts that touch their programs).
**Target:** 12–13 Oct · **Depends on:** [programs-and-demand](programs-and-demand.md) · **Branch:** `feat/staffing-draft`

## In scope

- PLAN (`baseline` and `draft`), ASSIGNMENT, ALLOCATION.
- `internal/engine` (standard library only) with the **greedy** strategy:
  - **Hard constraints:** expertise at ≥ minimum level · weekly hours ≤ capacity × `maxUtilization` · at most `maxExpertisesPerPersonWeek` expertises per person per week · no allocation below `minHoursPerWeek` except a demand's last remainder · only inside the demand's window · locked allocations untouched · never a stretch assignment
  - **Soft objectives:** priority and deadline order · primary-expertise match · load near `targetUtilization` · continuity on the same project
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
- [ ] `internal/engine`: types, greedy strategy, reasons, unmet codes
- [ ] Snapshot builder in the planning service
- [ ] Plans, assignments, allocations: models, repositories, services, handlers, DTOs
- [ ] Atomic accept with the conditional UPDATE and the partial unique index; audit record
- [ ] Draft edits, kind rules, stretch confirmation, reject
- [ ] Migration, constants, Swagger

**Client**
- [ ] Propose squad on intake and the program page
- [ ] Draft review: grid, reasons, unmet, diff, edit, lock, accept, reject, 409 handling

**Tests**
- [ ] Engine golden tests on fixed snapshots, and property tests for every hard constraint
- [ ] Accept: stale version, concurrent accepts (integration), unique index
- [ ] Service tests for kind and stretch rules
