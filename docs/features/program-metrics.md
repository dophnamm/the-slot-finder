# Program metrics

**Goal:** each program shows how far along it is and whether it is understaffed, and the team sees one list of risks. This covers G1 and G2 in the [root README](../../README.md#1-problem-and-scope).

**Roles:** planner, lead.
**Target:** 16 Oct · **Depends on:** [staffing-draft](staffing-draft.md) · **Branch:** `feat/program-metrics` · **Status:** In review

## In scope

- Per program and per project, with formulas from [root README §7](../../README.md#7-metrics):
  - Time %
  - Scope % and the **behind on delivery** warning (`SCOPE_LAG_THRESHOLD` percentage points)
  - Budget, planned and forecast hours (forecast = planned)
  - Gap hours with Staffing %, returned side by side
  - Planned burn % (no warning until actual hours exist)
  - At risk: a gap above 0 with the deadline within `ALERT_HORIZON_WEEKS`, or the last phase past the deadline
- Staffing by expertise: budget, planned, gap and Staffing % per expertise.
- A risks list: understaffed and at-risk projects, projects behind on delivery, and overloaded people.

**Fallback** if late at the 14 Oct checkpoint: only Staffing % with its gap.

## Not now

- EAC hours, forecast override, contribution %, cross-expertise share → [Later](README.md#later), design in [root README §7](../../README.md#7-metrics).
- Next milestone → [Later](README.md#later) (milestones).
- Sustained overload → [Later](README.md#later).

## Server

**Endpoints** (subset of [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/programs/{id}/overview` | Header metrics, scope and cost figures, staffing by expertise | planner, lead |
| GET | `/programs/{id}/projects` | Projects with Time / Scope / Planned burn / Staffing and at-risk flags | planner, lead |
| GET | `/insights/risks` | Understaffed, at-risk and behind-on-delivery projects; overloaded people | planner, lead |

**Tables:** none new.

**Rules**
- Metrics are pure functions in the insights module. "Now" comes from the Clock, so tests pin it.
- A ratio with a zero denominator returns `null`.
- Program values sum over projects that are not cancelled.
- Warnings are flags in the response (`behind_on_delivery`, `understaffed`, `at_risk`, `over_planned`). The Client shows them and never compares numbers itself.

**As built**
- The formulas live in `internal/metrics` as pure functions with "today" as a parameter. Percentages are 0–100; Time % is 0 before the start and may pass 100, which sets `past_end` while the work is not closed.
- `at_risk` also covers a gap with the deadline already past. A program is at risk when any of its projects is.
- A project's Time % runs from `start_earliest` to `deadline`; a program's from its start to end date.
- The overview sums over projects that are not cancelled, done ones included. The risks list leaves out done and cancelled projects.
- Overloaded people are those over 100 % in the next `HEATMAP_DEFAULT_WEEKS` weeks from the current week, the same range the dashboard heatmap shows.
- `GET /programs/{id}/projects` adds a `metrics` object to each project.
- An incoming project without accepted assignments shows its whole budget as the gap, so it is understaffed until a staffing draft covers it.

## Client

- `/programs/:id`:
  - header with Time, Scope, Planned burn and Staffing bars and their warnings
  - Cost figures (budget, planned, planned to date, gap)
  - staffing-by-expertise table
  - Projects tab rows with the same bars
- `/`: the risks list next to the heatmap.
- Feature folders: `features/program-overview` and `features/program-projects` move from mock data to API data; `features/risks` is new.
- A `null` metric is shown as "—".
- As built: `MetricBar` (shared) draws each percentage with its bar, red on a warning. The dashboard puts the risks list next to the heatmap on wide screens and above it otherwise.

## Acceptance criteria

With "now" pinned in tests:
1. A program from 1 Oct to 31 Dec, on 15 Nov, has Time % ≈ 49 %.
2. 30 of 100 points done gives Scope % 30 %. With Time % at 55 % and the threshold at 20, the program is behind on delivery. With Time % at 45 %, it is not.
3. Budget 400 h and planned 320 h give Staffing 80 %, a gap of 40 h, and `understaffed`.
4. With the deadline within `ALERT_HORIZON_WEEKS` and a gap above 0, the project is at risk.
5. Planned burn = planned hours up to and including the current week ÷ budget. It never raises a warning.
6. A project with budget 0 returns `null` for Staffing %, and the Client shows "—".
7. The risks list shows every understaffed project, every at-risk project and every person over 100 % in the coming weeks.

## Tasks

**Server**
- [x] Metric functions with a Clock; overview, projects and risks endpoints
- [x] Settings used: `SCOPE_LAG_THRESHOLD`, `ALERT_HORIZON_WEEKS`; constants; Swagger

**Client**
- [x] Program header bars and warnings, cost figures, staffing table, project rows from the API
- [x] Risks list on the dashboard

**Tests**
- [x] One table test per formula, including `null` on a zero denominator
- [x] Component tests for the bars and the risks list
