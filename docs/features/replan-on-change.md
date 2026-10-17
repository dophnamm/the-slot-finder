# Re-plan on change

**Goal:** when an absence or a phase slip breaks the plan, the Planner gets a re-plan draft with a diff, without recalculating anything by hand. This covers G3 in the [root README](../../README.md#1-problem-and-scope).

**Roles:** specialist and manager cause changes (leave, phase slips), planner reviews the draft, manager sees the banner on their program.
**Target:** 15 Oct · **Depends on:** [staffing-draft](staffing-draft.md) · **Branch:** `feat/replan-on-change` · **Status:** Done (PR #14)

## In scope

- CHANGE_EVENT as a **transactional outbox**:
  - `AbsenceAdded`, `AbsenceRemoved`, `PhaseSlipped` and `ProjectSlipped` are inserted in the same transaction as the change itself
  - no event without its change, and no change without its event
- A **Postgres-backed job runner** that claims pending events with `FOR UPDATE SKIP LOCKED`.
- Impact assessment: does the baseline now overload someone, or leave demand uncovered or outside its window?
- `replan` through the same engine: the baseline as `existing`, a raised `stability` weight, and locked allocations kept.
- A draft linked to its events, with a diff against the baseline.
- A re-plan banner on the program page, and the change history.
- **Debounce and merge** as designed in [root README §8.2](../../README.md#82-a-change-happens--re-plan): events within `REPLAN_DEBOUNCE` give one run, one open auto draft per baseline, and `409 DRAFT_SUPERSEDED`.

**Fallback** if late at the 14 Oct checkpoint: no debounce or merge. One event gives one draft, and an open auto draft stays as it is.

## Not now

- `CapacityChanged` and `PersonLeft` → [Later](README.md#later), with capacity changes.
- Email and Slack alerts → [Later](README.md#later). The alert is the banner and the change history.

## Server

**Endpoints** (subset of [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/events?programId=` | Change history and processing status; drives the banner | planner, manager |
| GET | `/staffing-drafts/{id}` | The re-plan draft, with its linked events | planner, manager |

**Tables and migration:**
- `change_event` with the design columns (`type`, `payload`, `occurred_at`, `status`, `draft_id`)
- runner columns: `attempts`, `available_at` and `last_error`
- an index on `(status, available_at)`
- `make migrate-generate name=replan_on_change`

**Outbox and runner**
- The service that saves a change calls `events.Record(ctx, event)` inside the same `TxManager.Do`. If the change rolls back, so does the event.
- The runner starts from an FX lifecycle hook and stops on shutdown. Every `REPLAN_POLL_INTERVAL` it opens a transaction and claims up to `REPLAN_BATCH_SIZE` rows:
  `SELECT … FROM change_event WHERE status = 'pending' AND available_at <= now() ORDER BY occurred_at FOR UPDATE SKIP LOCKED LIMIT :batch`.
  Two server instances never claim the same row.
- **Debounce:** an event is only claimed once it is older than `REPLAN_DEBOUNCE`, so events close together are claimed in one batch.
- The batch ends as `no_impact` or `draft_created` (with the draft id). An event that keeps failing ends as `failed`.
- On error, `attempts` goes up, `available_at` moves by `REPLAN_RETRY_BACKOFF × attempts`, and `last_error` is kept. After `REPLAN_MAX_ATTEMPTS` the status becomes `failed`, which shows in the change history.
- Processing is idempotent. Re-running a batch regenerates the same open auto draft instead of creating a second one.
- **Merge:** if an auto draft is open for the baseline, it is regenerated from the baseline plus all linked events, keeping manual assignments and locks. The old version becomes `superseded`, and a write to it gets `409 DRAFT_SUPERSEDED` ([server/README §6](../../server/README.md#6-background-jobs)).

**As built**
- `GET /events?program_id=` uses snake case like the rest of the API.
- Debounce is per event, as specified: an event is claimed once it is older than `REPLAN_DEBOUNCE`. Two events close together may land in two batches; the second then merges into the open auto draft, so there is still one draft linked to both.
- The impact assessment checks the whole baseline: a person over capacity × `PLANNING_MAX_UTILIZATION` in a week, or hours outside the demand's current window. The re-plan scope is the conflicting projects plus the open auto draft's projects, keeping only incoming, planned or active projects with demand.
- The engine now keeps unlocked hours of the starting plan that still fit (window, skill, free hours, minimum) and plans only the gap, in as many seats as the headcount has free. Before, one week of leave could move a whole demand to someone else.
- Each event stores the projects it touches: the project for a slip, and for absence the person's projects with hours in those dates (everyone's for a holiday). The program history and banner filter on them.
- Merge safety: an edit touches the draft (`updated_at`); the runner supersedes the old version only if it is unchanged, else it regenerates. A Postgres advisory lock lets one batch build drafts at a time, so two servers never open two auto drafts for one baseline.
- `plans.superseded_by` links a superseded version to the newer one. `GET /staffing-drafts/{id}` returns `auto`, `triggered_by` and `superseded_by`.
- A failed batch rolls back, then its events get the attempt, back-off and `last_error` in a separate write.
- `REPLAN_DEBOUNCE` defaults to 5m; `.env.example` sets 30s for local demos.

## Client

- `/programs/:id`: a re-plan banner when `/events?programId=` returns an event with status `draft_created`. It links to the draft.
- `/events`: change history with status, linked draft and `failed` events.
- `/staffing-drafts/:id`: the existing draft review, plus "triggered by" with the linked events. On `409 DRAFT_SUPERSEDED` it reloads the latest version.
- Feature folders: `features/replan-banner`, `features/change-history`; `features/draft-review` gains the triggers.
- As built: `shared/events` holds the API and refetches every 15 s, since changes are handled in the background. `EventSummary` (shared) writes one sentence per change. A superseded draft shows a link to the newer version rather than redirecting on its own.
- As built, `ProjectStatusChanged`, `DemandChanged`, `PersonCapacityChanged` and `TeamChanged` also record change events in the same transaction, and only the events that open projects or free capacity widen the re-plan scope (server README section 6).

## Acceptance criteria

1. Adding approved leave that overloads someone produces a draft linked to that event within `REPLAN_DEBOUNCE` + `REPLAN_POLL_INTERVAL`. The program page shows the banner.
2. Leave that causes no conflict ends as `no_impact`, with no draft.
3. If saving the absence fails, no change event exists (outbox).
4. Two runner instances never process the same event (integration test with two concurrent claims).
5. Two leaves added within `REPLAN_DEBOUNCE` give one draft linked to both.
6. A third leave while that auto draft is open gives a new draft version linked to all three, with the Planner's locks kept. Editing the old version gets `409 DRAFT_SUPERSEDED`.
7. An event that keeps failing ends as `failed` after `REPLAN_MAX_ATTEMPTS` and shows in the change history.
8. Moving a phase end later shifts the later phases and records one `PhaseSlipped` event in the same transaction.

## Tasks

**Server**
- [x] `change_event` model with runner columns; `events.Record` used by absences and phases inside their transactions
- [x] Runner: FX lifecycle, poll loop, `SKIP LOCKED` claim, retries, `failed`
- [x] Impact assessment and `replan` call; draft linked to events
- [x] Debounce and merge; `DRAFT_SUPERSEDED` on writes to a superseded version
- [x] Settings `REPLAN_POLL_INTERVAL`, `REPLAN_BATCH_SIZE`, `REPLAN_MAX_ATTEMPTS`, `REPLAN_RETRY_BACKOFF`; migration; Swagger

**Client**
- [x] Re-plan banner, change history, "triggered by" on the draft page, 409 reload

**Tests**
- [x] Outbox rollback; concurrent claims (integration); retry and `failed`
- [x] Debounce batch; merge keeps locks; `DRAFT_SUPERSEDED`
