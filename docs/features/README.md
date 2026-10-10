# Feature specs

One spec per feature, in build order. Each spec holds the goal, scope, server and client work, acceptance criteria and a task checklist. The design behind them is in the [root README](../README.md), [`server/README.md`](../server/README.md) and [`client/README.md`](../client/README.md).

**Demo:** Sunday 18 Oct 2026.

| # | Feature | What it delivers | Target | Depends on | Status |
|---|---|---|---|---|---|
| 1 | [Sign-in](sign-in.md) | Keycloak realm import, PKCE login in the client, local JWT verification, roles, `/me` | 10–11 Oct | | In progress |
| 2 | [People and skills](people-and-skills.md) | People, expertise, levels, absences (company-wide holidays only), `capacity(person, week)`, seed data | 11 Oct | sign-in | Not started |
| 3 | [Programs and demand](programs-and-demand.md) | Programs, projects, phases with slip propagation, demand (profile `flat` only) | 12 Oct | people-and-skills | Not started |
| 4 | [Staffing draft](staffing-draft.md) | Greedy engine, draft generation, review, edit and lock, atomic accept into the baseline | 12–13 Oct | programs-and-demand | Not started |
| 5 | [Team heatmap](team-heatmap.md) | Person × week heatmap, utilization bands | 14 Oct | staffing-draft | Not started |
| 6 | [Re-plan on change](replan-on-change.md) | Absence or phase slip → ChangeEvent → outbox runner → re-plan draft with diff, banner on the program page | 15 Oct | staffing-draft | Not started |
| 7 | [Program metrics](program-metrics.md) | Time, Scope, Planned burn, Staffing with gap, risks list | 16 Oct | staffing-draft | Not started |

## Schedule

- **Checkpoint, 14 Oct:** sign-in through team-heatmap work end to end on seed data. Sign in, see people and programs, generate and accept a staffing draft, see the heatmap change.
- **If late at the checkpoint:**
  - re-plan-on-change ships without debounce or merge: one event gives one draft
  - program-metrics keeps only Staffing % with its gap
- **17 Oct, freeze:** bug fixes, seed polish, demo script, rehearsal. No new features.
- **18 Oct:** demo.

## Later

Designed in the root README, not built for the demo:

- Skill tags (required and preferred) and the preferred person
- Capacity changes (effective-dated weekly hours) and holidays per location
- What-if scenarios
- CSV import
- Email and Slack notifications
- Capacity forecast by expertise
- Bench list
- Plan churn
- EAC hours
- Program assistants and milestones
- AI assistant

## Conventions for these specs

- A spec is named after its feature, and so is its branch (`feat/<feature>`).
- "Not now" links to the design section that describes the deferred part.
- When a feature ships, set its status to **Done** here and tick its tasks.
