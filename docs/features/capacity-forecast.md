# Capacity forecast

**Goal:** see per expertise and week whether the team has the capacity for the work ahead, confirmed and expected, so a Planner can hire or bring in contractors before a crunch. This covers the capacity forecast in the [root README §7](../../README.md#7-metrics).

**Roles:** planner, manager (Manager: their teams' capacity and the demand of the programs they may read).
**Target:** 17 Oct · **Depends on:** [team-heatmap](team-heatmap.md), [program-metrics](program-metrics.md), [program-teams](program-teams.md) · **Branch:** `feat/capacity-forecast` · **Status:** Done

## In scope

- `GET /insights/capacity-forecast`: per expertise × week, from the baseline and the open demand.
- Capacity, load and free hours of the people whose **primary** expertise it is; free hours never count an overloaded person against a free one.
- Unmet confirmed demand: hours of `planned` and `active` projects the baseline does not cover yet.
- Pipeline demand: the same for `incoming` projects, times the global `PIPELINE_CONFIDENCE_PERCENT`.
- An overload flag per cell, and the first sustained overload run per expertise.
- The people behind a cell: who is in the expertise that week and how each stands, shown in a popover on click.
- The Ask AI tool `get_capacity_forecast`.

## Not now

- A confidence on each project (`confidence`, [root README §6](../../README.md#6-data-model)): the field is not built, so one setting, `PIPELINE_CONFIDENCE_PERCENT`, weights every incoming project.
- FTE shortfall signals: the forecast shows hours only.
- Demand profiles other than `flat`: the engine only lays out `flat` today.
- Splitting a cross-expertise person's capacity across expertises: all of it counts for their primary expertise.
- Filters by expertise or program in the by-expertise view.

## Server

**Endpoint** (see [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/insights/capacity-forecast?from=&to=` | Expertise × week figures | planner, manager (team) |

**Tables:** none new.

**Rules**
- `from` and `to` default and are capped like the heatmap's (`HEATMAP_DEFAULT_WEEKS`, `HEATMAP_MAX_WEEKS`).
- Capacity and load come from the heatmap's cells, summed over people by primary expertise. `free_hours = Σ max(0, capacity − load)`.
- Remaining demand of a demand row = `total_hours` − its baseline hours, spread flat over its window's weeks from the current week on. Past weeks never count.
- `unmet_hours` come from `planned` and `active` projects, `pipeline_hours` from `incoming` ones, times `PIPELINE_CONFIDENCE_PERCENT`. On hold, done and cancelled projects add nothing.
- `shortfall_hours = max(0, unmet + pipeline − free)`. `overloaded` is load above capacity.
- `sustained_overload` is the first run of `SUSTAINED_OVERLOAD_WEEKS` or more overloaded weeks, found with `metrics.SustainedOverload`; `null` when none.
- A manager sees capacity of the people in the teams of the programs they lead, and demand of the programs they may read. Expertises are ordered by catalogue `sequence`, then name.
- `people` of a cell are the primary-expertise people from the heatmap's cells, highest load first (then name), with the heatmap's `utilization` and `band`; people with no capacity and no load are left out. `metrics.RankForecastPeople` derives `free_hours` and `overloaded` and sorts.
- The computation is `metrics.CapacityForecast`, a pure function; the insights service gathers its inputs.

**Response** (`data`)

```json
{
  "weeks": ["2026-W42", "2026-W43"],
  "expertises": [
    {
      "expertise_id": "uuid",
      "expertise": "QA",
      "weeks": [
        {"week": "2026-W42", "capacity": 72, "load": 60, "free_hours": 12, "unmet_hours": 8,
         "pipeline_hours": 10, "shortfall_hours": 6, "overloaded": false,
         "people": [{"person_id": "uuid", "name": "Quinn", "capacity": 36, "load": 32, "free_hours": 4,
                     "utilization": 0.89, "band": "near_full", "overloaded": false}]}
      ],
      "sustained_overload": {"from_week": "2026-W44", "weeks": 3}
    }
  ]
}
```

## Acceptance criteria

1. Two QA people with 36 h capacity each and 60 h load in a week show `free_hours` 12 and `overloaded` false.
2. A QA person at 40 h load on 36 h capacity shows `overloaded` true, and `free_hours` 0 for that person.
3. An active project needing 80 h QA with 40 h in the baseline shows 10 h `unmet_hours` in each of its four remaining weeks.
4. An incoming project needing 40 h PM over four weeks shows 5 h `pipeline_hours` a week at 50 %, and no load for anyone.
5. Overloaded weeks in a row, at least `SUSTAINED_OVERLOAD_WEEKS`, fill `sustained_overload`; fewer leave it `null`.
6. A manager's figures hold only their teams' people and the demand of programs they may read; a specialist gets `403`.
7. Each cell lists its people, highest load first; a person with no capacity and no load is not listed, and one at 40 h load on 36 h capacity has `overloaded` true and `band` `overloaded`.
8. A range wider than `HEATMAP_MAX_WEEKS` gets `WEEK_RANGE_TOO_LONG`.

## Tasks

**Server**
- [x] `metrics.CapacityForecast` with table tests
- [x] `people` per cell, with `metrics.RankForecastPeople`
- [x] `InsightsService.CapacityForecast`, handler, DTO, route, constants, Swagger
- [x] Setting `PIPELINE_CONFIDENCE_PERCENT`
- [x] Ask AI tool `get_capacity_forecast`

**Client**
- [x] A forecast view (expertise rows, week columns, shortfall and overload marks) as the **By expertise** view of the dashboard's heatmap card, switched with a **By person** / **By expertise** toggle; both views share one week range (`shared/week-range`)
- [x] A cell opens a popover with its people (name link, utilization badge, `load/capacity hrs`, free hours)

**Tests**
- [x] Metrics: shortfall, expected pipeline, overload, sustained run
- [x] Service: capacity, unmet and pipeline demand, ordering, forbidden role
- [x] Client: tone and rounding helpers, the hook, the grid (cells, shortfall, sustained overload badge), the dashboard toggle (default view, switching, range kept)
