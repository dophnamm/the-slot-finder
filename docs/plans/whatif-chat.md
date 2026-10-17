# What-if chat Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A planner asks Ask AI a what-if in plain language. Ask AI resolves project and person names, saves a scenario on a copy of the baseline, and answers with the result. A manager gets no scenario tools.

**Architecture:** Three tools are added to the Ask AI loop in `server/internal/services/ai.go`: `find_projects` and `find_people` (read tools, caller's rights) and `create_scenario` (calls `ScenarioService.Create`, planner only). `aiToolDeclarations(role)` offers the model only the tools the caller may use. `dispatch` re-checks each call through the service. A per-question guard allows one `create_scenario` per Ask. The system prompt describes what-if. The client has no change: the answer is markdown in the existing Ask AI panel.

**Tech Stack:** Go (Echo v5, FX, testify).

**Spec:** [docs/features/whatif-chat.md](../features/whatif-chat.md)

## Global Constraints

- No literals: tool names (`constants.ToolFindProjects = "find_projects"`, `ToolFindPeople = "find_people"`, `ToolCreateScenario = "create_scenario"`), the scenario name prefix (`aiScenarioNamePrefix = "What-if (AI): "`), the once-per-question error (`MsgOneScenarioPerQuestion = "one scenario per question"`), the assumption types (`constants.AssumptionShiftProject`, `constants.AssumptionAddAbsence`), roles (`constants.ScenarioRoles`) and the system prompt text are constants.
- Tools run with the caller's rights. `dispatch` never bypasses the service's `authz.Require`; a refused tool returns `{error}` to the model.
- Data policy: Ask AI already sends names and data through tools to the LLM provider. This plan keeps that policy and adds no new fields beyond the tool results listed in the spec.
- Heavy rate limit stays on the Ask route. No route change, so Swagger is unchanged.
- Load the matching skills: `writing-go-server-code`, `testing-go-server-code`, `structuring-server-modules`. There is no client task.
- Commits: local only on `feat/whatif-chat`, never push. `git add` only the files of the task. Run `make check-server` once before each task's last commit. Conventional subject; the message ends with the trailer `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>`.

## Review Focus

1. A manager obtains a scenario tool. Test: Task 2 case "manager has no create_scenario" and "manager refused"; Task 3 RBAC through the loop.
2. The model invents ids. Test: Task 2 case "bad project id gives {error}", and Task 3 loop with a bad id.
3. Two scenarios from one question. Test: Task 2 case "second call refused"; Task 3 loop with two `create_scenario` calls.
4. Tool results with person names go back to the provider. This is the existing Ask AI policy, stated in the spec. Test: none new; Task 3 checks that the tool result cut still applies.
5. A very long question in the scenario name. Test: Task 2 case "name cut to 60 characters".

---

### Task 1: `find_projects` and `find_people` tools

**Files:**
- Modify: `server/internal/constants/ai.go` (`ToolFindProjects`, `ToolFindPeople`, the declaration names next to the existing tool names ~46-51)
- Modify: `server/internal/services/ai.go` (declarations in `aiToolDeclarations`, dispatch cases in `dispatch` ~181)
- Test: `server/internal/services/ai_test.go` (add cases to the existing tool tests)

**Interfaces:**
- Consumes: `ProjectService.List` and `ProjectService.Get` (existing), `PersonService.List` (existing; the name filter is added at service level only if it is missing).
- Produces:
```go
// tool results, JSON
// find_projects: {"projects": [{"id", "name", "program", "deadline"}]}   max 5
// find_people:   {"people":   [{"id", "name", "primary_expertise"}]}      max 5
```

- [ ] **Step 1: Write the failing tests** — table tests in `ai_test.go`:
  - `find_projects`: "name match, case-insensitive"; "no match gives an empty list, not an error"; "at most 5 matches"; "manager allowed, specialist refused with {error}"
  - `find_people`: "name match"; "no match gives an empty list"; "at most 5 matches"; "planner and manager allowed"
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/services -run 'FindProjects|FindPeople' -v`
- [ ] **Step 3: Implement** the two constants, the two declarations and the two dispatch cases. Each dispatch case calls the existing service, filters by `query` and caps the result at 5 items.
- [ ] **Step 4: Run, expect PASS** — `cd server && go test ./internal/services`
- [ ] **Step 5: Commit** `feat(server): add find_projects and find_people to Ask AI`

### Task 2: Role-filtered declarations and `create_scenario`

**Files:**
- Modify: `server/internal/constants/ai.go` (`ToolCreateScenario`, `MsgOneScenarioPerQuestion`, `aiScenarioNamePrefix` in the services file)
- Modify: `server/internal/services/ai.go` (`aiToolDeclarations(role)` filters by role; `create_scenario` dispatch case with the per-Ask guard; builds the result from `ScenarioView`)
- Test: `server/internal/services/ai_test.go` (scenario cases), `server/internal/services/ai_declarations_test.go` (create)

**Interfaces:**
- Consumes: `ScenarioService.Create(ctx, CreateScenarioInput{Name, Assumptions})` returning `ScenarioView{Plan, Changes}` with `Plan.Summary` (`PeopleChanged`, `HoursMoved`, `UnmetHours`), `constants.ScenarioRoles`, `constants.AssumptionShiftProject`, `constants.AssumptionAddAbsence`, `ScenarioMaxAssumptions`.
- Produces:
```go
// aiToolDeclarations(role string) []llm.ToolDeclaration  // only tools the role may use
// create_scenario result, JSON:
// {"scenario_id", "name", "people_changed", "hours_moved", "unmet_hours",
//  "top_changes": [{"person", "demand", "before", "after"}]}   top_changes max 5, largest hours first
// refusal or bad input: {"error": "..."}
```
- Guard: a per-Ask counter (a local variable in the Ask call, passed to `dispatch`), not package state. The second `create_scenario` in the same Ask returns `{"error": MsgOneScenarioPerQuestion}` and saves nothing.
- Name: `aiScenarioNamePrefix` + the question cut to 60 characters.

- [ ] **Step 1: Write the failing tests**
  - `aiToolDeclarations`: "planner has all three tools"; "manager has no create_scenario"
  - `create_scenario`: "valid shift saves and returns result fields"; "valid absence saves"; "bad project id gives {error} and saves nothing"; "manager refused with {error}"; "second call in the same Ask refused"; "name is the prefix plus the first 60 characters of the question"; "top_changes capped at 5, largest first"
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/services -run 'CreateScenario|Declarations' -v`
- [ ] **Step 3: Implement** the filter in `aiToolDeclarations(role)`, the `create_scenario` case, the guard and the result builder. Pass the role from the caller's `Principal` into `Ask`.
- [ ] **Step 4: Run, expect PASS** — `cd server && go test ./internal/services`
- [ ] **Step 5: Commit** `feat(server): let Ask AI save a what-if scenario for planners`

### Task 3: System prompt and the `Ask` loop test

**Files:**
- Modify: `server/internal/services/ai.go` (system prompt constant gets the what-if paragraph and the rules listed in the spec)
- Test: `server/internal/services/ai_test.go` (add `TestAIService_AskWhatIf`)

**Interfaces:**
- Consumes: the fake `llm.Client` (existing test helper), the tools from Tasks 1 and 2.
- Produces: the system prompt text as a constant; no new exported symbols.

- [ ] **Step 1: Write the failing tests** — `TestAIService_AskWhatIf` with a scripted fake `llm.Client`:
  - "find_projects, then create_scenario, then final text": the answer is the final text; one scenario saved
  - "tool result cut still applies": a result longer than `LLM_MAX_TOOL_RESULT_BYTES` is cut before it goes back to the model
  - "two create_scenario calls in one Ask": the second returns `{error}`; one scenario saved
  - "manager Ask has no create_scenario in the tools sent to the model"
  - "system prompt contains the what-if rules" (assert the constant, not a literal copy)
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/services -run TestAIService_AskWhatIf -v`
- [ ] **Step 3: Implement** the prompt paragraph and rules. Keep the existing prompt text unchanged except for the additions.
- [ ] **Step 4: Run, expect PASS** — `make check-server` from the repo root
- [ ] **Step 5: Commit** `feat(server): teach Ask AI the what-if tools`

### Task 4: Docs

Written from the diff of Tasks 1–3, reviewed before commit. Server and root docs in one commit.

- [ ] `README.md` §15: what-if chat is built via Ask AI tools
- [ ] `server/README.md` §9: tools table with the three tools as built, the role filter and the one-scenario-per-question rule; `get_scenario_diff` stays folded into `create_scenario`'s result; "what-if chat" leaves the Later list
- [ ] `client/README.md`: the AI chat panel row (no UI change)
- [ ] `docs/features/README.md`: row 22 → Done; remove "what-if chat" from the Later AI item (commit only those lines if the file holds another feature's work)
- [ ] `docs/features/whatif-chat.md`: status Done, tasks ticked
- [ ] `docs/features/scenarios.md`: the "AI what-if chat → Later" bullet
- [ ] `docs/user-manual.md`: Ask AI section gets example what-if questions, planner only
- [ ] `README.md:3` and `CLAUDE.md:9`: twenty-two features built
- [ ] Drawio and the README AI node: "what-if chat" removed from the later list
- [ ] Commit `docs: describe what-if chat`

### Task 5: Final check

- [ ] `make check` from the repo root passes
- [ ] Optional: `cd server && go test -tags=integration ./...` with Docker running
- [ ] `git log --oneline` on `feat/whatif-chat` shows only this feature's commits; nothing pushed
