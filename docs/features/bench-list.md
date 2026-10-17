# Bench list

**Goal:** show who is idle for the coming weeks, so a Planner or Manager can find work for them. This covers the bench list in the [root README §7](../../README.md#7-metrics).

**Roles:** planner, manager (Manager: the people of their teams).
**Target:** 17–18 Oct · **Depends on:** [team-heatmap](team-heatmap.md) · **Branch:** `feat/bench-list` · **Status:** Done

## In scope

- A person is on the bench when, in **each** of the next `BENCH_WEEKS` weeks starting with the current week, their capacity is above 0 and their utilization is below `UTILIZATION_UNDERUSED_BELOW`. Both come from `clock.Clock` and the settings (default 4 weeks, default 0.5).
- Free hours: Σ over those weeks of (capacity − baseline allocated hours). Utilization and hours come from the same source as the heatmap (`server/internal/services/insights.go`, `heatmapRow` at `:230`, `utilizationBand` at `:288`).
- `GET /insights/bench`, a list sorted by free hours.
- The **Bench** list on the dashboard's Risks card.

## Not now

- The suggested actions from the old design: a stretch assignment, training, internal work. See the bench row in [root README §7](../../README.md#7-metrics). They are a Planner action, not part of this list.
- A separate `/insights` page.

## Server

**Endpoint** (see [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/insights/bench` | People on the bench, most free hours first | planner, manager (team) |

**Tables:** none new.

**Rules**
- The rule is a pure function, `metrics.Bench`, in `server/internal/metrics`, next to `SustainedOverload` (`metrics.go:189`). It has table tests.
- The window is the current week and the next `BENCH_WEEKS - 1` weeks.
- Utilization is below the band: at exactly `UTILIZATION_UNDERUSED_BELOW` a person is off the bench.
- A week with 0 capacity (leave, or outside the person's start and end dates) keeps the person off the bench.
- Capacity and load per week are exactly the ones the heatmap computes for the caller: same baseline, same capacity function, same people. Whatever the heatmap counts as load, the bench counts. `metrics.Bench` takes those per-week numbers; it does not query anything.
- Only people the caller may see are listed (`AuthzService.VisiblePeople`), the same rows as the heatmap. A manager sees only the people of their programs' teams.
- `primary_expertise` is the name of the person's primary expertise. Every person has exactly one.
- Sorted by free hours, most first, then by name.
- Route path, role list and messages are constants. Add Swagger annotations and regenerate Swagger.
- Config: `BENCH_WEEKS`, default 4. Add it to the config struct, `cmd/server/.env.example` and [server/README §8](../../server/README.md#8-settings).

**Response** (`data`)

```json
{
  "weeks": 4,
  "people": [
    {"person_id": "uuid", "name": "Quinn", "primary_expertise": "QA", "free_hours": 104}
  ]
}
```

`weeks` is `BENCH_WEEKS`, so the client can say "{hours} h free in {weeks} weeks" without its own copy of the setting.

## Acceptance criteria

1. A person with 36 h capacity and 10 h load in each of the next four weeks is on the bench, with 104 h free.
2. One week at or above `UTILIZATION_UNDERUSED_BELOW` takes the person off the bench.
3. A week with 0 capacity takes the person off the bench.
4. Utilization exactly at `UTILIZATION_UNDERUSED_BELOW` is not on the bench.
5. The list is sorted by free hours, most first, then by name.
6. A manager sees only the people of their teams; a specialist gets `403`.
7. At the end of the Risks card, "Bench" shows each person's name, primary expertise and free hours. A name opens that person's page. With nobody on the bench, the card shows an empty state.

## Client

- The **Bench** list sits at the end of the Risks card (`client/src/features/risks/component.tsx`), after the risk groups.
- Each row: name (a link to the person's page), primary expertise, and "{hours} h free in {weeks} weeks".
- Empty state when nobody is on the bench.
- Endpoint in `src/resources/endpoints/insights.ts`. TanStack Query hook inside the risks feature. i18n keys in `LanguageKey`.

## Tasks

**Server**
- [x] `metrics.Bench` with table tests
- [x] `InsightsService.Bench`, handler, DTO, route, constants, Swagger
- [x] Setting `BENCH_WEEKS`

**Client**
- [x] Endpoint, hook and types in the risks feature
- [x] Bench list and empty state in the Risks card

**Tests**
- [x] Metrics: on the bench; one week at or above the threshold makes it off; a 0-capacity week makes it off; the threshold boundary
- [x] Handler and RBAC: specialist `403`; manager sees only their teams' people
- [x] Client: the list, the name link and the empty state

**Docs** (same PR)
- [x] [README.md](../../README.md): the Bench list row in §7 is built, with no suggested actions
- [x] [server/README.md](../../server/README.md): move `GET /insights/bench` out of the Later routes; update the "Team signals" row; add `BENCH_WEEKS` to the config table and remove it from the "Later settings" list
- [x] [client/README.md](../../client/README.md): the Risks card text; remove "`/insights`: bench list" from Later
- [x] [team-heatmap](team-heatmap.md): the bench bullet under Not now
- [x] [user-manual](../user-manual.md): the dashboard section
