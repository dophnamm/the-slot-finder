# Program teams

**Goal:** a planner gives each program its team, and the program's manager staffs the program's projects from that team only. A manager sees the people of the programs they manage, not everyone in the company.

**Roles:** planner (adds and removes team members, staffs anything), manager (the `lead` role, shown as **Manager**: reads and staffs the programs they manage), specialist (unchanged).
**Target:** 14–15 Oct · **Depends on:** [programs-and-demand](programs-and-demand.md), [staffing-draft](staffing-draft.md), [member-calendar](member-calendar.md) · **Branch:** `feat/program-teams` · **Status:** Done

## In scope

- **Manager.** The realm role stays `lead`; the app shows it as **Manager**. A manager manages a program when they are its lead (`programs.lead_id`).
- **Team.** A program has a team: the people a planner added to it. A person may be in several teams. Their weekly capacity is shared: the engine already counts every baseline hour, whatever the program.
  - The planner adds and removes members on the program's **Team** tab, which lists each member with their expertise and the hours they have in this program's baseline.
  - Removing a member keeps their baseline hours; the next draft for the program plans without them.
- **Staffing from the team.** The engine staffs a program's demand only with members of that program's team. This holds for Propose squad, re-plan drafts and scenarios.
  - When nobody in the team can take the demand, the draft reports it as unmet with the reason **No one in the program team** (`NO_ONE_IN_TEAM`), so the planner knows to add people.
  - On a staffing draft, Add person lists the team of the demand's program only. The server refuses a person outside the team (`422 PERSON_NOT_IN_TEAM`).
- **Manager's view.** A manager sees:
  - on People and the dashboard heatmap: the members of the teams of the programs they manage, and themselves;
  - a person page: only for those people (else `403`);
  - the overloaded people in Risks: only those people.
  - Programs, projects and events stay readable as before.
- **Manager's staffing.** A manager may propose, edit, accept and reject a staffing draft when every project in it belongs to a program they manage. Any other draft stays with the planner.

## Not in this feature

- A manager asking the planner for more people (a request flow). The unmet reason tells the planner instead.
- Hiding another program's project names from a manager on a team member's calendar.
- Renaming the `lead` realm role in Keycloak.

## Server

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/programs/{id}/team` | The team: each member with their expertise and their hours, weeks and projects in this program's baseline | planner, the program's manager |
| POST | `/programs/{id}/team` | Add a member: `{ person_id }`; `409 TEAM_MEMBER_EXISTS` when already in | planner |
| DELETE | `/programs/{id}/team/{personId}` | Remove a member | planner |

- **Data:** `program_team_members (program_id, person_id, created_at)`, unique on the pair, both foreign keys cascade on delete.
- **Visibility:** `AuthzService.VisiblePeople(ctx)` answers "everyone" for a planner, and for a manager their own id plus the person ids of the teams of the programs they read (they lead it or lead one of its projects). `GET /people`, `GET /people/{id}`, `GET /people/{id}/schedule`, the heatmap and the risks use it.
- **Programs a manager reads:** `AuthzService.ProgramReader(ctx)` answers "every program" for a planner, and for a manager the programs they lead or lead a project in. `GET /programs`, the project list (intake) and the risks leave the other programs out; `GET /programs/{id}`, its overview, members and projects, and `GET /projects/{id}` answer `403` for them.
- **Drafts:** the draft routes also admit `lead`; the staffing service lets a manager act only when every project of the draft belongs to a program they lead (else `403`).
- **Engine:** `engine.Demand.Candidates` (a set of person ids; `nil` means anyone). The staffing snapshot fills it with the team of the demand's program. A person outside the set is never proposed, and their unlocked hours on that demand are dropped; locked hours stay.
- **Seed:** each demo program gets a team: the people the demo baseline uses, plus one or two more.

## Client

- The role label **Lead** becomes **Manager** everywhere.
- Program page: the **Members** tab becomes **Team**: member, expertise, hours in this program, weeks, projects; planners get **Add member** (a dialog to pick a person) and **Remove** on each row.
- People and the dashboard show what the server returns; a manager's People list is read-only.
- A staffing draft's Add person lists the demand's program team.
- Propose squad and the draft actions show for a manager on the programs they manage.

## Acceptance criteria

1. A planner adds Bao Tran to the team of Acme Commerce Platform; the Team tab lists him with 0 h until a draft gives him hours.
2. Propose squad for an Acme project proposes only Acme team members. A demand whose expertise nobody in the team has is unmet with **No one in the program team**.
3. The manager of Acme (Lee Lead) sees only Acme's team on People and the heatmap; opening someone outside it gives `403`.
4. Lee proposes, edits and accepts a draft for an Acme project. Lee gets `403` on a draft that also covers a Beta project.
5. A planner still sees everyone and staffs any program.

## Tasks

**Server**
- [x] Model, migration, repository for team members; constants (paths, roles, messages, error codes)
- [x] Team service and handlers (list with baseline hours, add, remove), routes, Swagger
- [x] `VisiblePeople` in `AuthzService`; apply it to people list and detail, schedule, heatmap and risks
- [x] Engine: `Demand.Candidates`, `NO_ONE_IN_TEAM`; snapshot fills candidates from the team; draft edit refuses a person outside the team
- [x] Draft routes admit managers; the staffing service checks every project's program
- [x] Seed teams for the demo programs
- [x] Tests: engine candidates, team service, visibility, draft rights, RBAC through the router, an integration test of a manager's draft

**Client**
- [x] Manager label; Team tab with Add member and Remove; draft Add person from the team; manager draft actions
- [x] Tests: Team tab, add member dialog, People read-only for a manager

**Estimate:** about 1.5 days (server 1, client 0.5).

## As built

- **Draft data for the client:** each draft project carries `program_lead_id`, and each draft demand its program team as `team_ids` (`null` puts no limit). Project lists that load the program (intake, a program's projects) carry `program_lead_id` too. The client shows Propose squad and the draft actions with `canStaff(user, programLeadIds)`: a planner always, a manager when they manage every program involved.
- **Empty team:** a program with no members gets an empty candidate set, so every demand is unmet with `NO_ONE_IN_TEAM` until a planner adds people.
- **Person page:** a manager opening someone outside their teams sees **You can’t see this person**; the person query keeps the `403` quiet (queries can list `quietErrorCodes` like mutations), and the calendar loads only once the person does. Queries no longer retry a `403`.
- **Wording:** the program's lead is shown as **Manager** in the programs list, the program header and the program form; a project's lead keeps the name **Lead**.
- **Members preview:** the Programs list preview keeps its Members tab (`/programs/{id}/members`, everyone with baseline hours); the program page shows the Team tab instead.
