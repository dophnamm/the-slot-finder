# What-if chat

**Goal:** a planner asks Ask AI a what-if in plain language, e.g. "what if Checkout Revamp slips 3 weeks?" or "what if Quinn is away from 2026-11-02 to 2026-11-13?". Ask AI saves a scenario (a what-if on a copy of the baseline, never on the baseline) and answers with the result: people changed, hours moved, unmet hours and the biggest changes. The AI design is in the [root README §15](../../README.md#15-ai-assistant) and [server README §9](../../server/README.md#9-ai-assistant).

**Roles:** planner (scenarios are planner-only). Ask AI itself stays planner and manager.
**Target:** 17–18 Oct · **Depends on:** [scenarios](scenarios.md), [ask-ai](ask-ai.md) · **Branch:** `feat/whatif-chat` · **Status:** Planned

## In scope

- Three new Ask AI tools: `find_projects` and `find_people` resolve names to ids, `create_scenario` saves the scenario and returns its result.
- The model is offered only the tools the caller's role allows. A manager gets no `create_scenario`, so the model says it cannot run a what-if.
- Every tool runs with the caller's rights. The service's own `authz.Require` check still applies; a refused tool returns `{error}` to the model, the same 403 the route gives.
- At most one scenario per question. A second `create_scenario` call in the same Ask returns `{error: "one scenario per question"}`.
- A scenario made from chat is named `What-if (AI): <first 60 chars of the question>`. The prefix marks it as made by chat.
- The Ask AI system prompt describes what-if and the three tools.

## Not now

- Write tools for staffing drafts (`create_staffing_draft`).
- More assumption types (new project, planned hire, demand change).
- A clickable link from the answer to the scenario. The answer names the scenario and says it is on the Scenarios page.
- An audit field `via` (`ai_assistant`). The audit note in the root README stays later.

## Server

**Tools** (names are constants in `server/internal/constants/ai.go`; see [server/README §9](../../server/README.md#9-ai-assistant))

| Tool | Arguments | Result | Roles |
|---|---|---|---|
| `find_projects` | `query` (text) | up to 5 `{id, name, program, deadline}`; name match, case-insensitive | planner, manager (as `ProjectService`) |
| `find_people` | `query` (text) | up to 5 `{id, name, primary_expertise}` | planner, manager (as `PersonService.List`) |
| `create_scenario` | `assumptions`: array of `{type, project_id, weeks, person_id, from_date, to_date, hours_per_day}` | `{scenario_id, name, people_changed, hours_moved, unmet_hours, top_changes}`; `top_changes` is up to 5 `{person, demand, before, after}` | `constants.ScenarioRoles` (planner) |

- `type` is `shift_project` (`constants.AssumptionShiftProject`) or `add_absence` (`constants.AssumptionAddAbsence`), as in the Scenarios page.
- `create_scenario` calls `ScenarioService.Create`. `people_changed`, `hours_moved` and `unmet_hours` come from `Plan.Summary`; `top_changes` comes from the scenario's changes, largest hours first.
- A request with more than `ScenarioMaxAssumptions` assumptions is refused by the service, as today.
- A bad project or person id gives `{error}`; the model is told to resolve names with `find_*` first and never to invent ids.

**Declarations** `aiToolDeclarations(role)` returns only the tools the role may use. Manager gets `find_projects` and `find_people`; planner gets all three.

**Loop** `AIService.Ask` keeps the existing rules: `LLM_MAX_TOOL_ROUNDS`, and tool results cut at `LLM_MAX_TOOL_RESULT_BYTES`. The once-per-question guard lives in the Ask call, not in package state.

**System prompt** adds: what-if and the three tools; resolve names with `find_*` first; never invent ids; after `create_scenario`, give the scenario name and say it is on the Scenarios page; show numbers from the tool result only; weeks as `2026-W45`; dates as `YYYY-MM-DD`.

**Data policy.** Ask AI already sends names and data through its tools to the LLM provider. This feature follows that existing policy: the tool results for `find_people` and `create_scenario` contain person names and demand names. The rate limit does not change: the Ask route keeps the heavy rate limit.

**Route:** none new. Swagger is unchanged.

**Tables:** none new. The scenario is saved through the existing scenario tables.

## Acceptance criteria

1. A planner asking "what if Checkout Revamp slips 3 weeks?" gets a saved scenario named `What-if (AI): …` and an answer with its people changed, hours moved and unmet hours.
2. The scenario is a copy: the baseline and its allocations are unchanged.
3. A manager's Ask gets no `create_scenario` in the tools offered to the model; a forced call returns `{error}` with the same 403 the route gives.
4. `find_projects` and `find_people` return at most 5 matches; no match gives an empty list, not an error.
5. A second `create_scenario` in the same question returns `{error: "one scenario per question"}` and saves nothing.
6. A bad project or person id returns `{error}` to the model; no scenario is saved.
7. A question longer than 60 characters gives a scenario name of the prefix plus 60 characters of the question.
8. The tool result cut at `LLM_MAX_TOOL_RESULT_BYTES` still applies to the new tools.
9. Route, RBAC and Swagger are unchanged.

## Client

- No new UI. The Ask AI panel already renders the answer as markdown; the scenario name is plain text in it.
- No new endpoint, hook or i18n key.

## Tasks

**Server**
- [ ] `find_projects` and `find_people` tools
- [ ] Role-filtered `aiToolDeclarations(role)` and the `create_scenario` tool with the once-per-question guard
- [ ] System prompt and the `Ask` loop with the what-if tools

**Client**
- [ ] None (the answer is plain markdown, no UI change)

**Tests**
- [ ] `find_projects` and `find_people`: match, no match, max 5, rights
- [ ] `create_scenario`: valid shift, valid absence, bad project id gives `{error}`, manager refused, second call refused, result fields
- [ ] `aiToolDeclarations`: planner has all three, manager has no `create_scenario`
- [ ] `Ask` loop with the fake `llm.Client`: `find_projects`, then `create_scenario`, then final text; tool result cut still applies
- [ ] Handler and RBAC unchanged

**Docs** (same PR)
- [ ] [README.md](../../README.md): §15 says what-if chat is built via Ask AI tools
- [ ] [server/README.md](../../server/README.md): §9 tools table; the three tools move from Later to built, with the role filter and the one-scenario-per-question rule; `get_scenario_diff` stays folded into `create_scenario`'s result; "what-if chat" leaves the Later list
- [ ] [client/README.md](../../client/README.md): the AI chat panel row
- [ ] [docs/features/README.md](README.md): row 22 → Done; remove "what-if chat" from the Later AI item
- [ ] [docs/features/scenarios.md](scenarios.md): the "AI what-if chat → Later" bullet
- [ ] [user-manual](../user-manual.md): Ask AI section gets example what-if questions, planner only
- [ ] Drawio and README AI node: "what-if chat" removed from the later list
- [ ] `README.md:3` and `CLAUDE.md:9`: twenty-two features built
