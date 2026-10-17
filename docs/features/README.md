# Feature specs

One spec per feature, in build order. Each spec holds the goal, scope, server and client work, acceptance criteria and a task checklist. The design behind them is in the [root README](../../README.md), [`server/README.md`](../../server/README.md) and [`client/README.md`](../../client/README.md).

**Demo:** Sunday 18 Oct 2026.

| # | Feature | What it delivers | Target | Depends on | Status |
|---|---|---|---|---|---|
| 1 | [Sign-in](sign-in.md) | Keycloak realm import, PKCE login in the client, local JWT verification, roles, `/me` | 10–11 Oct | | Done |
| 2 | [People and skills](people-and-skills.md) | People, expertise, levels, absences (company-wide holidays only), `capacity(person, week)`, seed data | 11 Oct | sign-in | Done |
| 3 | [Programs and demand](programs-and-demand.md) | Programs, projects, phases with slip propagation, demand (profile `flat` only) | 12 Oct | people-and-skills | Done |
| 4 | [Staffing draft](staffing-draft.md) | Greedy engine, draft generation, review, edit and lock, atomic accept into the baseline | 12–13 Oct | programs-and-demand | Done |
| 5 | [Team heatmap](team-heatmap.md) | Person × week heatmap, utilization bands | 14 Oct | staffing-draft | Done |
| 6 | [Re-plan on change](replan-on-change.md) | Absence or phase slip → ChangeEvent → outbox runner → re-plan draft with diff, banner on the program page | 15 Oct | staffing-draft | Done |
| 7 | [Program metrics](program-metrics.md) | Time, Scope, Planned burn, Staffing with gap, risks list | 16 Oct | staffing-draft | Done |
| 8 | [Sign-in theme](sign-in-theme.md) | Keycloak login theme in the app's look, no extra sign-in click, sign-out goes straight back to sign-in | before 17 Oct | sign-in | Done |
| 9 | [Program quick view](program-quick-view.md) | Expand a program in the list to see its figures, projects and members; Members tab | 16–17 Oct | program-metrics | Done |
| 10 | [Scenarios](scenarios.md) | What-if: move a project or add leave, see the changes, churn and unmet demand without touching the baseline | 12–13 Oct | staffing-draft, replan-on-change | Done |
| 11 | [Member calendar](member-calendar.md) | A person page and My schedule with a month calendar: absences, holidays, weekly load against capacity, hours per project | 13–14 Oct | people-and-skills, team-heatmap | Done |
| 12 | [Timesheet import](timesheet-import.md) | CSV template from the baseline, import of actual hours, Actual hours and Burn % on programs and projects | 14–15 Oct | program-metrics, member-calendar | Done |
| 13 | [Calendar sync](calendar-sync.md) | Secret ICS link from Google or Outlook; out-of-office events become planned leave and feed the re-plan flow | 15–16 Oct | member-calendar, replan-on-change | Done |
| 14 | [Program teams](program-teams.md) | A planner gives each program its team; the program's manager sees and staffs only that team, and the engine staffs a program from its team | 14–15 Oct | programs-and-demand, staffing-draft, member-calendar | Done |
| 15 | [Ask AI](ask-ai.md) | Floating assistant for planners and Managers: plain-language questions answered only from read-only tools that run with the caller's rights (Gemini, Groq or OpenRouter, key from env) | 16 Oct | staffing-draft, team-heatmap, program-metrics, program-teams | Done |
| 16 | [Guided tour](guided-tour.md) | First-time tour of the sidebar, one step per link the role can open and one on the account menu; started again from the account menu | 17 Oct | sign-in | Done |
| 17 | [Capacity forecast](capacity-forecast.md) | By-expertise view of the dashboard heatmap, per expertise × week: free capacity, unmet confirmed demand, pipeline demand × the global pipeline confidence, overload and sustained overload | 17 Oct | team-heatmap, program-metrics, program-teams | Done |
| 18 | [Skill and program events](skill-and-program-events.md) | A skill change re-plans the person's projects; a program end-date move is recorded in the change history | 17–18 Oct | people-and-skills, replan-on-change | Done |
| 19 | [Bench list](bench-list.md) | People idle for the next weeks, shown on the Risks card | 17–18 Oct | team-heatmap | Planned |

## Schedule

- **Checkpoint, 14 Oct:** sign-in through team-heatmap work end to end on seed data. Sign in, see people and programs, generate and accept a staffing draft, see the heatmap change.
- **If late at the checkpoint:**
  - re-plan-on-change ships without debounce or merge: one event gives one draft
  - program-metrics keeps only Staffing % with its gap
- **17 Oct, freeze:** bug fixes, seed polish, demo script, rehearsal. No new features.
- **18 Oct:** demo.

## Later

Designed in the root README, not built for the demo:

- Holidays per location
- More scenario assumptions (new project, planned hire, demand change) and turning a scenario into a draft
- CSV import
- Email and Slack notifications
- Bench list
- Plan churn
- EAC hours, hours per point, contribution %, cross-expertise share
- Forecast overrides, per-project confidence and the `front` and `back` profiles
- AI assistant beyond Ask AI: write actions, intake extraction, what-if chat, draft summaries and more tools (server README section 9)

## Conventions for these specs

- A spec is named after its feature, and so is its branch (`feat/<feature>`).
- "Not now" links to the design section that describes the deferred part.
- When a feature ships, set its status to **Done** here and tick its tasks.
