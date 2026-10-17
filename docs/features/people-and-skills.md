# People and skills

**Goal:** the system knows who can do what, at which level, and how many hours each person has free per week. This is the supply side for staffing (G1) and the heatmap (G2) in the [root README](../../README.md#1-problem-and-scope).

**Roles:** planner (manages people), manager (reads), specialist (own absences).
**Target:** 11 Oct · **Depends on:** [sign-in](sign-in.md) · **Branch:** `feat/people-and-skills` · **Status:** Done (PR #5)

## In scope

- PERSON: name, email, `keycloak_id`, role, `weekly_hours`, `non_project_hours`, `start_date`, `end_date`, `active`.
- EXPERTISE catalogue (PM, Analyst, Engineer, QA) and PERSON_SKILL (level 1–5, exactly one primary).
- ABSENCE: personal leave, and **company-wide holidays only** (`person_id` null).
- `capacity(person, week)` as in the [root README §6](../../README.md#6-data-model), without location holidays.
- `isoweek.Week`: the Go type for weeks (`2026-W42`). Capacity is the first code that needs it.
- `/me` gains `person_id`, linked by `keycloak_id`, falling back to email on the first sign-in.
- Seed data: 12 people, 4 expertises, 3 leave periods, company-wide holidays, and the three demo users as people.

## Not now

- Holidays per location → [Later](README.md#later).
- `location` and `timezone` on PERSON: columns can wait until holidays per location.

## Server

**Endpoints** (subset of [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET, POST | `/people` | List or create people | GET planner, manager · POST planner |
| PATCH | `/people/{id}` | Edit a person (hours, dates, active) | planner |
| PUT | `/people/{id}/skills` | Replace expertise and levels | planner |
| GET, POST | `/people/{id}/absences` | Leave calendar | planner, the person themselves |
| DELETE | `/absences/{id}` | Remove an absence | planner, the person themselves |
| GET | `/expertises` | Expertise catalogue | everyone |
| GET, POST | `/holidays` | Company-wide holidays | GET everyone · POST planner |

**Tables and migration:** `person`, `expertise`, `person_skill`, `absence`. One Atlas migration generated from the GORM models (`make migrate-generate name=people_and_skills`).

**Rules**
- `email` and `keycloak_id` are unique, enforced by the database. A duplicate returns `409`.
- Exactly one skill per person has `is_primary`. Levels are 1–5.
- Absence hours in a week = working days of the absence inside that week × `hours_per_day`. Working days are config (`WORKING_WEEKDAYS`, e.g. Monday to Friday).
- `approved` absence always counts. `planned` absence counts when `COUNT_PLANNED_ABSENCE` is on.
- Capacity is 0 before `start_date` and after `end_date`, and never below 0.
- `authz` gains `CanManagePeople` and `CanEditAbsence(personID)` (planner, or the person themselves).
- Capacity is a pure function (`internal/capacity`), so it is tested without a database. When a person starts or leaves mid-week, the base hours shrink to the working days they are employed; overlapping absences count once per day. No endpoint returns capacity yet; team-heatmap is its first user.

## Client

- `/people`:
  - a list with an expertise × level matrix and weekly hours
  - a person form (create and edit) with a skills editor
  - an absence list per person
- `/me`: the signed-in specialist's own absences, with "report leave".
- Feature folders: `features/people` (list, person dialog, skills editor), `features/absences`, `features/holidays`. The People page composes them and opens a person's absences in a dialog.
- Holidays: a small list with "add holiday" for planners, on the people page.

## Acceptance criteria

1. A planner creates a person with two skills, one of them primary, and gets `201`. With zero or two primary skills, the request gets a validation error.
2. Creating a person with an existing email gets `409`.
3. For a person with 40 h, 4 h non-project and a 2-day approved leave at 8 h per day in a week, capacity that week is 20 h.
4. Capacity is 0 for a week before `start_date`.
5. A company-wide holiday lowers everyone's capacity that week.
6. With `COUNT_PLANNED_ABSENCE` false, planned leave does not lower capacity. Approved leave still does.
7. A specialist can add and remove their own absence, and gets `403` for someone else's.
8. After sign-in, `/me` returns the demo user's `person_id`.
9. `make seed` loads the seed data into an empty database, and running it again adds no duplicates.

## Tasks

**Server**
- [x] Models, repositories, services, handlers and DTOs for people, expertise, skills, absences and holidays
- [x] `isoweek` package: parse, format, Monday date, add weeks, range
- [x] Capacity function and its config (`WORKING_WEEKDAYS`, `DEFAULT_NON_PROJECT_HOURS`, `COUNT_PLANNED_ABSENCE`)
- [x] `/me` links PERSON by `keycloak_id`, then by email
- [x] Migration, route constants, role lists, messages, error codes, Swagger
- [x] `cmd/seed` and `make seed`, idempotent

**Client**
- [x] People list, person form, skills editor, absences, holidays
- [x] `/me` absences with "report leave"
- [x] Endpoints in `resources/endpoints`, routes and RBAC in `resources`

**Tests**
- [x] Capacity table tests (absence overlap, start and end dates, holidays, planned vs approved)
- [x] `isoweek` tests across year boundaries (2026 has 53 weeks: 2026-W53 runs 28 Dec 2026 – 3 Jan 2027, and 2027-W01 starts 4 Jan 2027)
- [x] Service tests for uniqueness and the primary-skill rule
- [x] RBAC tests for absences (self vs other)
