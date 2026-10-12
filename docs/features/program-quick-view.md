# Program quick view

**Goal:** from the Programs list, a planner or lead opens any program in place and sees its key figures, its projects and who works on it, without leaving the list. The program page's Members tab shows the same people in full.

**Roles:** planner, lead.
**Target:** 16–17 Oct · **Depends on:** [program-metrics](program-metrics.md) · **Branch:** `feat/program-quick-view` · **Status:** Done (PR #21)

## In scope

- **Programs list:** an arrow at the start of each row opens a preview under it; a second click closes it. A click elsewhere on the row still opens the program page.
- **Preview**, laid out like the program page and loaded only when opened:
  - a line of figures: budget, planned and gap hours (gap red when understaffed), and the Staffing bar
  - a **Projects (n)** tab: a table of name (link), status and warnings, deadline week, Staffing bar
  - a **Members (n)** tab: the same members table as the program page
- **Program page, Members tab:** a bordered table of the same members (person, expertise, hours, weeks, project links), in place of the placeholder.
- **Server:** `GET /programs/{id}/members`.

## Not now

- Each member's overall utilization across programs, and assistants → [Later](README.md#later).
- Editing members here; that stays in staffing drafts.

## Server

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/programs/{id}/members` | Everyone with baseline hours on the program's projects, most hours first | planner, lead |

- Computed from the published baseline: assignments whose demand belongs to one of the program's projects, summed per person.
- Each member: `person`, `total_hours`, `from_week` and `to_week` (first and last week with hours), `expertises` (hours per expertise and assignment kind, most hours first) and `projects` (by name).
- Assignments without hours are left out. An unknown program is `404 PROGRAM_NOT_FOUND`.
- Lives with the other program read models in `InsightsService` (`ProgramMembers`).

## Client

- `shared/programs`: `ProgramMember` types, `useProgramMembers(id)` under the program's query key, so accepting a draft refreshes it.
- `features/programs`: the arrow and the preview row (`program-preview` with its Projects table in `preview-projects`). The preview reuses the overview, project and member queries of the program page.
- `shared/components/members-table`: the members table, used by the preview and the Members tab.
- `features/program-members`: the Members tab.

## Acceptance criteria

1. On the Programs list, the arrow opens and closes a preview under the row; the rest of the row still opens the program.
2. The preview shows budget, planned, gap and Staffing, then Projects and Members tabs: each project with status, warnings, deadline and Staffing, and each member with hours per expertise, total, weeks and projects.
3. Nothing is fetched for a program until its preview is opened.
4. A member's hours count only the program's projects; hours on other programs are left out.
5. The Members tab lists the same members, with links to their projects.
6. A specialist gets `403` from `/programs/{id}/members`; an unknown program gets `404`.

## Tasks

**Server**
- [x] `ProgramMembers` in the insights service, DTOs, handler, route, Swagger
- [x] Unit tests for the sums (other programs, empty assignments, order) and the service (404, 403); RBAC case in the router test

**Client**
- [x] Members types, API call and hook
- [x] Expandable preview on the Programs list, with Projects and Members tabs
- [x] Members tab on the program page
- [x] Component tests for the preview toggle and the Members table

**Estimate:** about three hours.
