# Team heatmap

**Goal:** see weeks ahead who is overloaded or underused, and which weeks will crunch. This covers G2 in the [root README](../../README.md#1-problem-and-scope).

**Roles:** planner, lead.
**Target:** 14 Oct (checkpoint) · **Depends on:** [staffing-draft](staffing-draft.md) · **Branch:** `feat/team-heatmap` · **Status:** Done (PR #10)

## In scope

- A person × week matrix from the baseline.
  - Each cell holds load, capacity, utilization, its band, and hours per expertise.
  - Planned absence is marked tentative.
- Utilization bands from config: underused, healthy, near full, overloaded (over 100 %).
- Filters: week range and expertise. The expertise filter keeps everyone who holds it in the range.
- A cell's detail: that person's assignments in that week.
- The dashboard at `/`.

## Not now

- Bench list, capacity forecast → [Later](README.md#later), design in [root README §7](../../README.md#7-metrics).
- Comparing a draft or scenario on the heatmap → [Later](README.md#later).

## Server

**Endpoints** (subset of [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/insights/heatmap?from=&to=&expertise=` | Person × week cells | planner, lead |
| GET | `/people/{id}/assignments?from=&to=` | A cell's detail | planner, lead, the person themselves |

**Tables:** none new. The insights module reads the baseline through planning and capacity through availability.

**Rules**
- `utilization = load ÷ capacity`.
  - With capacity 0 and no load, utilization is `null` and the cell is unavailable.
  - With capacity 0 and some load, the cell is overloaded.
- Bands use `UTILIZATION_UNDERUSED_BELOW` and `UTILIZATION_NEAR_FULL_FROM` ([server/README §8](../../server/README.md#8-settings)). The Server returns the band, so the Client never compares numbers.
- `from` and `to` default to the current ISO week (from the Clock) and the next `HEATMAP_DEFAULT_WEEKS` weeks. The range is capped at `HEATMAP_MAX_WEEKS`.
- Rows are ordered by primary expertise, then name.

**As built**
- Bands: below `UTILIZATION_UNDERUSED_BELOW` underused, below `UTILIZATION_NEAR_FULL_FROM` healthy, up to and including 100 % near full, above it overloaded. No capacity and no load is `unavailable`.
- A week is tentative when the person has planned leave on a working day they are employed, whether or not `COUNT_PLANNED_ABSENCE` lets it lower capacity.
- `expertise=` takes the expertise id. It keeps people who hold the skill or have hours on it in the range.
- Rows follow the expertise catalogue order (`sequence`), then name.
- Capacity uses the same inputs as staffing drafts: the person's own absence plus company holidays.

## Client

- `/`: a heatmap grid with rows of people grouped by primary expertise and columns of ISO weeks with their date ranges.
  - Each cell's colour comes from the band returned by the Server.
  - A cell shows a small split of its hours by expertise. Tentative absence is hatched.
  - Clicking a cell opens a popover with that week's assignments.
- Filters for week range and expertise are UI state (Zustand), not server data.
- Feature folder: `features/heatmap`.
- As built: a cell shows utilization, load of capacity and any cross-expertise hours as text (`+QA 8`), since sized bars would need inline styles. The filters default to the current week and 12 weeks; the server still caps the range.
- The client gets its test runner here: Vitest with Testing Library in jsdom (`npm test`, part of `make check`).

## Acceptance criteria

1. After accepting a draft, the heatmap shows the new hours without a manual reload.
2. A person at 44 h load with 40 h capacity shows as overloaded. At 30 h they are healthy, at 36 h near full, at 10 h underused, with the default bands.
3. A person on a company-wide holiday week shows reduced capacity. Planned leave is hatched.
4. Filtering by QA keeps an Engineer who holds QA cover that week, marked as non-primary.
5. Clicking a cell lists that week's assignments with project, expertise and hours.
6. A range wider than `HEATMAP_MAX_WEEKS` gets a validation error.

## Tasks

**Server**
- [x] Heatmap query in insights (baseline allocations + capacity), bands, expertise filter
- [x] Settings `HEATMAP_DEFAULT_WEEKS`, `HEATMAP_MAX_WEEKS`; constants; Swagger

**Client**
- [x] Heatmap grid, cell split, tentative marking, popover
- [x] Filters in a Zustand store; query invalidated after a draft is accepted

**Tests**
- [x] Band boundaries, zero capacity, expertise filter (service tests)
- [x] Heatmap grid component test
