# Timesheet import

**Goal:** record the hours people actually worked, so burn shows real hours instead of planned ones. A planner downloads a CSV template built from the baseline, fills in the actual hours, and uploads it. This replaces the "work logging" line in the [root README](../../README.md#1-problem-and-scope).

**Roles:** planner.
**Target:** 14–15 Oct · **Depends on:** [program-metrics](program-metrics.md), [member-calendar](member-calendar.md) · **Branch:** `feat/timesheet-import` · **Status:** Done (PR #33)

## In scope

- **Template.** For a range of past weeks, a CSV with one row per person × project × week:
  - every combination with baseline hours in the range, plus every one that already has actual hours
  - columns: `person_email`, `person_name`, `program`, `project`, `project_id`, `week`, `planned_hours`, `actual_hours`
  - `actual_hours` is filled in where it is already recorded, and empty otherwise
- **Import.** The planner uploads the filled-in file:
  1. The server reads the whole file and checks every row before it saves anything.
  2. One bad row means nothing is saved. The answer lists every problem by line and column.
  3. Otherwise each row with a value is saved for its person, project and week, in one transaction. A row with an empty `actual_hours` is skipped; `0` is saved as zero hours.
  4. Importing again overwrites the same person, project and week. Nothing is duplicated.
- **Metrics** ([root README §7](../../README.md#7-metrics)):
  - projects and programs get **Actual hours** (Σ actual hours) and **Burn %** (actual hours ÷ budget hours)
  - Burn % is `null` until the project has actual hours. Until then the client keeps showing Planned burn.
- **Member calendar:** a week with actual hours shows "worked N h" next to its load.
- **Seed:** `make seed` fills the template for Data Migration (15 % over plan) and Checkout Revamp (10 % under) and imports it through the same service, so Burn % has numbers for the demo. Patient App is left empty for a live import. Running the seed again adds nothing.

## Not in this feature

- Files exported by other timesheet tools, matched by name.
- Hours per day.
- People entering their own hours in the app.
- Actual hours per demand or expertise. They are recorded per project.

**As built**
- `updated_by` holds the planner's Keycloak subject, not a person id: a planner may have no person record.
- `ImportResult` lives in `models`, so `dtos` can map it.
- The server names the download `timesheet.csv`; the client saves it as `timesheet-<from>-<to>.csv`.
- Text cells that start with `= + - @` are written with a leading `'` in the template, so a spreadsheet shows them as text.
- A file over `TIMESHEET_MAX_BYTES` or `TIMESHEET_MAX_ROWS` gets `413 TIMESHEET_TOO_LARGE`.
- Files with `;` separators or `7,5` decimals are refused with clear problems (`COLUMN_MISSING`, `HOURS_INVALID`), not read.
- Blank rows, which a spreadsheet saves as a line of commas, are skipped and do not count toward `TIMESHEET_MAX_ROWS`.
- After a refused import the file field is emptied, so the fixed file (same name) can be picked again.
- The seed fills the template rows of Data Migration and Checkout Revamp that have no actual hours and imports them, so a second run adds nothing.
- Checked on the seeded stack: the program shows Burn 10 % (89 of 920 h), a file with problems on lines 5 and 7 saves nothing, the same file twice gives 0 added and 4 updated, a lead and a specialist get `403`, and the calendar shows "Worked 27.5 h".

## Server

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/timesheets/template?from=&to=` | The CSV template for a range of ISO weeks (`text/csv`, downloaded as a file) | planner |
| POST | `/timesheets/import` | Upload a filled-in template (`multipart/form-data`, field `file`). Returns how many rows were created, updated and skipped | planner |

**Table:** `actual_hours`: `person_id`, `project_id`, `week_start`, `hours`, `updated_by`, timestamps. Unique on person, project and week. One Atlas migration: `make migrate-generate name=actual_hours`.

**Row checks** (every failure is reported as `{ line, column, code }`)
- The header holds the required columns: `person_email`, `project_id`, `week`, `actual_hours`. Their order is free and the other columns are ignored, because they are there for people to read.
- `person_email` belongs to a person.
- `project_id` is a project that exists.
- `week` is an ISO week (`2026-W41`) no later than the current week.
- `actual_hours` is empty, or a number from 0 to `TIMESHEET_MAX_WEEK_HOURS`.
- The same person, project and week appear only once in the file.
- A person may have actual hours on a project without baseline hours there. Helping out is allowed.

**Rules**
- A file with any failed row gets `400` with `TIMESHEET_INVALID` and the list of problems in `data`.
- Limits: `TIMESHEET_MAX_BYTES` (default 1 MB) and `TIMESHEET_MAX_ROWS` (default 5,000). The template range is at most `TIMESHEET_MAX_WEEKS` (default 12), and without `from` and `to` it covers the last `TIMESHEET_TEMPLATE_WEEKS` (default 4) up to the current week.
- Importing records no change event. Actual hours describe the past and never trigger a re-plan.
- The insights module reads actual hours for the metrics. The schedule endpoint adds `actual_hours` to each week.
- Settings go in `cmd/server/.env.example` and [server/README §8](../../server/README.md#8-settings).

## Client

- People page: an **Import timesheet** button (planner) opens a dialog in two steps:
  1. Choose the from and to weeks (`WeekPicker`, default the last 4 weeks), then **Download template**.
  2. Choose the filled-in file, then **Import**. Success shows a toast with the counts. Failure shows a bordered table of line, column and problem inside the dialog.
- Program page: the header shows **Burn** (actual) when it exists, else Planned burn. Cost figures add **Actual hours**.
- Projects table: the burn column shows actual burn when it exists, with the same rule.
- Member calendar: "worked N h" on weeks with actual hours.

## Acceptance criteria

1. The template for W41–W42 has one row per person × project × week with baseline hours, with their planned hours.
2. After importing actual hours, the program shows Actual hours and Burn % = actual ÷ budget. Its projects do too.
3. A file with an unknown email on line 5 and a future week on line 7 saves nothing and lists both problems.
4. Importing the same file twice leaves the same data: rows are updated, not duplicated.
5. A lead or a specialist gets `403` on both routes.
6. The member calendar shows the worked hours for the imported weeks.

## Tasks

**Server**
- [x] Model and migration, constants (paths, roles, messages, error codes, CSV columns), settings
- [x] Repository: upsert by person, project and week; sums per project and per person × week
- [x] Service: build the template; parse and check the file; save in one transaction
- [x] Handlers (CSV download, multipart upload), routes, Swagger
- [x] Metrics: Actual hours and Burn % per project and program; `actual_hours` per week in the schedule
- [x] Tests: each row check, all-or-nothing import against a real DB (integration), metrics, RBAC
- [x] Seed actual hours for Data Migration and Checkout Revamp through the import

**Client**
- [x] Endpoints, types and hooks (file download and upload through `shared/api/http.ts`)
- [x] Import timesheet dialog with the error table
- [x] Burn and Actual hours on the program page and projects table; worked hours on the calendar
- [x] Tests: the dialog shows the counts or the error table; burn falls back to Planned burn

**Estimate:** about 9 hours (server 6, client 3).
