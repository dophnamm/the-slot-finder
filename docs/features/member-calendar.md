# Member calendar

**Goal:** a planner opens one person and sees their month: the days they are away (leave, training, holidays), the hours they carry each week against their capacity, and the projects those hours go to. A specialist sees the same view of themselves on **My schedule**. Actual hours ([timesheet-import](timesheet-import.md)) and leave read from their own calendar ([calendar-sync](calendar-sync.md)) show up here too.

**Roles:** planner (any person), manager (the people of their program teams, and themselves), specialist (themselves only).
**Target:** 13–14 Oct · **Depends on:** [people-and-skills](people-and-skills.md), [team-heatmap](team-heatmap.md) · **Branch:** `feat/member-calendar` · **Status:** Done (PR #32)

## In scope

- **Person page** `/people/:id`:
  - a page header: name, primary expertise, weekly hours, non-project hours
  - the calendar
  - the person's absences (the existing absence list). Planners and the person themselves see it; a planner can **Approve** a planned absence there.
- **The calendar** is a month grid:
  - one row per week, one column per working weekday (from the server), and a last column with the week's load and capacity
  - a day cell shows absences (leave, training) and holidays. Planned absence is marked tentative.
  - the week column shows `load / capacity h` in the heatmap's band colour, with a warning when overloaded
  - selecting a week lists its hours per project (program, project, expertise, hours). Each project links to its page.
  - previous and next month, and a **Today** button. Today's cell is outlined.
- **My schedule** (`/me`) shows the same calendar for the signed-in person, instead of the placeholder. The absences section stays.
- **People list:** each row opens the person page (`LinkRow`).
- **Breadcrumbs:** People › person name. "People" is a link for planners and managers; a manager sees only the people of their program teams and themselves.
- **Approve.** A planner turns a planned absence into an approved one. Until now the status could only be chosen when the absence was created. [calendar-sync](calendar-sync.md) depends on this, because leave read from a calendar arrives as planned.

## Not in this feature

- Editing hours from the calendar. Hours change through a staffing draft.
- Hours per day. Hours stay per week ([ADR-3](../../README.md#14-architecture-decisions)).
- Several people in one calendar. The heatmap does that.

**As built**
- The absence list on the person page is for planners. The person sees and reports their own on My schedule. Managers see absences on the calendar only, because the absence routes let a non-planner read only their own.
- The schedule types live in `models/schedule.go`, with aliases in `services`, because `dtos` cannot import `services`.
- Approving records `AbsenceAdded` with the new status; there is no separate event type. Approving twice, or approving a holiday, gets `409 ABSENCE_NOT_PLANNED`.
- The band classes and labels moved to `shared/utilization`. The person and absence types, their labels and `PLANNED_ABSENCE_STATUS` moved to `shared/people`.
- The breadcrumb `Trail` shows a middle step without a link as plain text. It is used for "People" when the user cannot open the list.
- Names on the dashboard heatmap open the person page. This is how managers reach it from the dashboard.
- The week detail names projects as plain text for viewers who cannot open project pages (a specialist on My schedule).
- Approving only changes an absence that is still planned, in the same statement. Of two approvals at the same time, one wins and the other gets `409`. The Approve button is disabled while it runs, and the list refreshes after a refusal.
- Checked on the local stack as planner, manager and specialist. The load of every week equals the heatmap's, and November 2026 has six rows. Planned leave shows as tentative. Approve updates the calendar without a reload and records `AbsenceAdded`. A specialist gets `403` on someone else's schedule.

## Server

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/people/{id}/schedule?from=&to=` | One person's weeks and days in a range of weeks: see the response below | planner, manager, the person themselves |
| GET | `/people/{id}` | One person, for the page header | planner, manager |
| POST | `/absences/{id}/approve` | Approve a planned absence. Records an `AbsenceAdded` change event with the new status, like any absence change | planner |

**Response**
- `working_weekdays`: the days that count, from `WORKING_WEEKDAYS`, so the client never hardcodes Monday to Friday.
- `weeks[]`: `week`, `capacity`, `load`, `utilization`, `band`, `tentative`, and `projects[]` (`program_id`, `program_name`, `project_id`, `project_name`, `expertise`, `hours`).
- `days[]`: only the days with something on them: `date`, `absences[]` (`id`, `type`, `status`, `hours_per_day`) and `holidays[]` (`id`). Holidays have no name; the client labels them "Holiday".

**Rules**
- `from` and `to` are weeks. At most `SCHEDULE_MAX_WEEKS` weeks (default 8, enough for a month grid); more gets `400`.
- Capacity, load, band and the tentative flag come from the same code as the heatmap, so the two never disagree.
- Load comes from the baseline only. Draft and scenario hours are not shown.
- Access goes through `AuthzService`: a planner may read anyone, a manager only the people of their program teams and themselves, and any other caller only their own person. Others get `403`.
- No new tables.
- Approving an absence that is already approved, or a holiday, gets `409`.

## Client

- Route `ROUTES.personDetail = '/people/:id'`, guarded by a new permission `people:view` (planner, manager).
- `features/person-calendar`: the month grid, the week detail and the month navigation. The person page and My schedule both use it.
- `pages/person-detail`: header, calendar, absences.
- `pages/my-schedule`: calendar for `user.person_id`, then the absences section.
- The People list rows become `LinkRow`s.

## Acceptance criteria

1. A planner opens Alex: every week shows the same load and capacity as the heatmap for that week.
2. Approved leave shows on its days. Planned leave shows as tentative. A company-wide holiday shows on everyone's calendar.
3. Selecting a week lists hours per project that add up to the week's load. Each project opens its page.
4. A specialist sees their own calendar on `/me` and gets `403` for anyone else's schedule.
5. A manager opens a person in their program teams and sees the calendar with its absences and holidays, without the absence list or Approve.
6. A person without baseline hours shows `0 / capacity` for each week, with their absences.
7. A planner approves a planned absence on the person page: it shows as approved on the calendar and records a change event. A manager or a specialist gets `403`.

## Tasks

**Server**
- [x] Constants (path, message, setting `SCHEDULE_MAX_WEEKS`), DTOs, handler, route with roles, Swagger
- [x] Schedule service: reuse the heatmap's capacity and load per week; hours per project; absences and holidays per day
- [x] `GET /people/{id}`; `POST /absences/{id}/approve` with its change event
- [x] Tests: service table tests (absences, holidays, tentative, range limit), RBAC through the router (planner, manager, self → 200; another specialist → 403)

**Client**
- [x] Endpoint, types and query hook in `shared/people`; route, permission and breadcrumb
- [x] `features/person-calendar`: month grid, week detail, month navigation
- [x] Person page, My schedule, People rows as `LinkRow`; Approve on planned absences
- [x] Tests: the grid builds the right weeks for a month, selecting a week lists its projects, tentative leave is marked

**Estimate:** about 8 hours (server 4, client 4).
