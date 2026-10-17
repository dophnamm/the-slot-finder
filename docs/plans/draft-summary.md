# Draft summary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A planner or manager asks for a short summary of a staffing draft. The text is written by the LLM from a small facts object, checked against the draft's numbers, and falls back to a template built from the same facts.

**Architecture:** A new file `server/internal/services/aisummary.go` holds the pure functions: `buildDraftFacts` turns the `DraftView` into `draftFacts`, `templateSummary` writes the fallback text, and `numbersKnown` checks the model's numbers. `AIService.SummarizeDraft` loads the draft through `StaffingService.Get`, calls the LLM with no tools, retries once on an invented number and falls back to the template. `POST /staffing-drafts/{id}/summary` returns `{ summary, source }`. The client's draft-review feature gets a `Summary` card child with a `useSummarizeDraft` mutation.

**Tech Stack:** Go (Echo v5, FX, testify), React + TypeScript (TanStack Query, Vitest, Testing Library).

**Spec:** [docs/features/draft-summary.md](../features/draft-summary.md)

## Global Constraints

- No literals: route path (`PathStaffingDraftSummary = "/staffing-drafts/:id/summary"`), roles (`constants.InsightRoles`), messages (`MsgDraftSummarized`), sources (`constants.SummarySourceAI = "ai"`, `SummarySourceTemplate = "template"`) and the prompt (`aiSummaryPrompt`) are constants. Client endpoint in `src/resources/endpoints`, text through `t()` with keys in `client/src/shared/i18n/language-key.ts` sorted alphabetically (compare case-insensitively, lowercase before uppercase on a tie).
- Heavy rate limit: `middlewares.HeavyRateLimit` on the route.
- Data policy: no person name, no email and no customer or client name in the facts. Project names are allowed. Tests assert this.
- The summary route never returns `503`, `502` or `429` for a provider problem; it returns the template with `source: "template"`.
- Swagger is regenerated (`make swagger-load`, or `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init` with the flags of the `swagger-load` target in `server/Makefile`), never edited by hand.
- Client component rules from the `structuring-client-components` skill (`component.tsx` holds only imports and one component, no inline function props; children in `components/<child>/`); no `style={...}`.
- Load the matching skills: `writing-go-server-code`, `testing-go-server-code`, `structuring-server-modules` (server); `translating-client-text`, `structuring-client-components`, `organizing-client-shared-code`, `styling-client-components` (client).
- Commits: local only on `feat/draft-summary`, never push. Server and client never in the same commit. `git add` only the files of the task. Run `make check-server` / `make check-client` once before each task's last commit. Conventional subject; the message ends with the trailer `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>`.

## Review Focus

1. An invented number is accepted by the check. Test: Task 1 cases "invented number", "21 when only 20 is known"; Task 2 case "invented twice falls back".
2. A person name or email leaks into the facts. Test: Task 1 case "no names in the JSON"; Task 2 case "facts passed to the LLM have no person name".
3. The LLM being unavailable returns an error instead of the template. Test: Task 2 cases "no LLM key" and "provider error"; Task 3 handler test.
4. A second click while the summary is pending sends a second request. Test: Task 4 case "button disabled while pending".
5. The draft changes after a summary and the card shows stale text. Test: Task 4 case "text clears when the draft version changes".

---

### Task 1: Facts, template and number check

**Files:**
- Create: `server/internal/services/aisummary.go`
- Test: `server/internal/services/aisummary_test.go` (create)

**Interfaces:**
- Consumes: `DraftView` (`server/internal/services/staffing.go:70-103`): assignments, per-demand `unmet_reason` and `assigned_hours`, `Changes`, `auto` flag.
- Produces:
```go
// draftFacts is the only input the LLM sees. No names, emails or customer names.
type draftFacts struct {
	People          int           `json:"people"`
	Assignments     int           `json:"assignments"`
	Projects        int           `json:"projects"`
	PlannedHours    float64       `json:"planned_hours"`
	HoursAdded      float64       `json:"hours_added"`
	HoursRemoved    float64       `json:"hours_removed"`
	IsReplan        bool          `json:"is_replan"`
	Unmet           []unmetFact   `json:"unmet"`
}

// unmetFact is one demand with unmet hours.
type unmetFact struct {
	Expertise  string  `json:"expertise"`
	Project    string  `json:"project"`
	Hours      float64 `json:"hours"`
	FirstWeek  string  `json:"first_week"`
	LastWeek   string  `json:"last_week"`
}

func buildDraftFacts(view *DraftView) draftFacts
func templateSummary(f draftFacts) string
func numbersIn(text string) []string          // regex `\d+(?:\.\d+)?`
func numbersKnown(text string, factsJSON string) bool
```
- Produces: `aiSummaryPrompt` (constant in `services/ai.go`, next to the existing system prompt): "Summarize a staffing draft in at most 3 sentences of plain English. Use only the numbers given in the facts; do not invent or round numbers. Do not mention ids. Plain text, no lists."

- [ ] **Step 1: Write the failing tests** — table tests in `aisummary_test.go`:
  - `TestBuildDraftFacts`: "re-plan flag from auto"; "unmet per demand with project name, expertise, hours and first and last week"; "hours added and removed from Changes"; "no names in the JSON": marshal the facts and assert the fixture's person names and emails are not in the string
  - `TestTemplateSummary`: the example from the spec ("4 people on 6 assignments across 2 projects, 312 planned hours. 20 h of QA is still uncovered on Checkout Revamp (2026-W45 to 2026-W47).") and a draft with no unmet hours
  - `TestNumbersKnown`: "accepts digits inside week labels" (`2026-W45` known from facts); "accepts 20.5 when 20.5 is known"; "rejects an invented number"; "rejects 21 when only 20 is known"
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/services -run 'TestBuildDraftFacts|TestTemplateSummary|TestNumbersKnown' -v`
- [ ] **Step 3: Implement** the four functions and the prompt constant. `numbersKnown` takes the facts JSON as a string and checks that each number in the text appears as a number token in it, so `2026-W45` matches the week label string in the facts.
- [ ] **Step 4: Run, expect PASS** — `cd server && go test ./internal/services`
- [ ] **Step 5: Commit** `feat(server): build draft facts and the template summary`

### Task 2: `AIService.SummarizeDraft`

**Files:**
- Modify: `server/internal/services/ai.go` (interface, `ProvideAIService` unchanged, new method next to `Ask`; reuse `upstreamError`)
- Modify: `server/internal/models/ai.go` (`DraftSummary{Summary string; Source string}`)
- Test: the existing AI service tests in `server/internal/services/` (add `TestAIService_SummarizeDraft`; reuse the fake `llm.Client`)

**Interfaces:**
- Consumes: `StaffingService.Get` (draft and `DraftView`), `buildDraftFacts`, `templateSummary`, `numbersKnown` (Task 1), `llm.Client.Generate(ctx, llm.Request{SystemInstruction, Contents, GenerationConfig})` with no `Tools`, `config.Config.LLMAPIKey`.
- Produces:
```go
// on AIService
SummarizeDraft(ctx context.Context, draftID string) (*DraftSummary, error)
```
- Flow: `Get` the draft (its errors pass through as `AppError`). Build facts. If `cfg.LLMAPIKey == ""`, return the template. Otherwise call the LLM with the facts JSON as the content. If the provider fails (`upstreamError` class: `AI_UPSTREAM_FAILED` or rate limited), return the template. If the text's numbers are all known, return it with `constants.SummarySourceAI`. If not, call once more; if that also fails the check, return the template with `constants.SummarySourceTemplate`.

- [ ] **Step 1: Write the failing tests** — `TestAIService_SummarizeDraft` (fake `llm.Client` returning a fixed text, a fake `StaffingService`):
  - "known numbers return ai": text "4 people on 6 assignments." → `source ai`
  - "invented twice falls back": two replies with `99` → `source template`, two `Generate` calls
  - "invented once then known": first reply invented, second known → `source ai`, two calls
  - "no LLM key": `LLMAPIKey` empty → `source template`, zero `Generate` calls
  - "provider error": `Generate` returns `AI_UPSTREAM_FAILED` → `source template`, no error
  - "facts passed to the LLM have no person name": the request's content does not contain any fixture person name or email
  - "draft not found passes the error": `Get` returns `NOT_FOUND` `AppError` → same error
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/services -run TestAIService_SummarizeDraft -v`
- [ ] **Step 3: Implement** `SummarizeDraft` as in the flow above; the prompt is `aiSummaryPrompt`. Add a `GenerationConfig` max-output field only if the llm package already has one; otherwise rely on the prompt.
- [ ] **Step 4: Run, expect PASS** — `cd server && go test ./internal/services`
- [ ] **Step 5: Commit** `feat(server): summarize a staffing draft with the LLM and a template fallback`

### Task 3: Route, handler, DTO, Swagger and RBAC

**Files:**
- Modify: `server/internal/constants/routes.go` (`PathStaffingDraftSummary = "/staffing-drafts/:id/summary"`)
- Modify: `server/internal/constants/messages.go` (`MsgDraftSummarized = "Draft summarized"`)
- Modify: `server/internal/constants/ai.go` or the existing constants file (`SummarySourceAI = "ai"`, `SummarySourceTemplate = "template"`)
- Modify: `server/internal/dtos/ai.go` (`DraftSummaryResponse{Summary string \`json:"summary"\`; Source string \`json:"source"\`}` and `NewDraftSummaryResponse(models.DraftSummary)`)
- Modify: `server/internal/handlers/ai.go` (`SummarizeDraft` handler, next to `Ask`; godoc with `@Summary Summarize a staffing draft`, `@Tags AI`, `@Success 200 {object} object{meta=dtos.Meta,data=dtos.DraftSummaryResponse}`, `@Failure 403`, `@Failure 404`, `@Router /staffing-drafts/{id}/summary [post]`)
- Modify: `server/cmd/server/routes/router.go` (next to `v1.POST(constants.PathAIAsk, ...)` ~187: `v1.POST(constants.PathStaffingDraftSummary, aiHandler.SummarizeDraft, authenticated, roles, heavy)` with `roles` = `RequireRole(constants.InsightRoles...)`)
- Regenerate: `server/docs`
- Test: `server/internal/dtos/ai_test.go` (`TestNewDraftSummaryResponse`), `server/internal/handlers/ai_test.go` (`httptest` cases), `server/cmd/server/routes/router_test.go` (RBAC rows next to the Ask AI rows)

**Interfaces:**
- Consumes: `AIService.SummarizeDraft` (Task 2).
- Produces: `POST /api/v1/staffing-drafts/{id}/summary` → `data: { summary, source }`. Json keys are `summary` and `source`; the client relies on them.

- [ ] **Step 1: Write the failing tests** — `TestNewDraftSummaryResponse` (keys and values); handler: `200` with `source` set; RBAC: specialist `POST` → `403`; manager and planner → not `403`; no-role token → `403`.
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/dtos -run TestNewDraftSummaryResponse -v && go test ./internal/handlers ./cmd/server/routes -run 'DraftSummary|RBAC' -v`
- [ ] **Step 3: Implement** the constants, DTO, handler (no request body; errors through `HandleError`), route and Swagger. Check `git diff --stat server/docs` shows only the new path and types.
- [ ] **Step 4: Run, expect PASS** — `make check-server` from the repo root
- [ ] **Step 5: Commit** `feat(server): add POST /staffing-drafts/{id}/summary`

### Task 4: Summary card on the draft review page (client)

**Files:**
- Modify: `client/src/resources/endpoints/staffing.ts` (`summary: (id: string) => \`/staffing-drafts/${id}/summary\``)
- Modify: `client/src/features/draft-review/api.ts` (`summarizeDraft(id)` via `request<DraftSummary>()`), `types.ts` (`DraftSummary = { summary: string; source: 'ai' | 'template' }`)
- Create: `client/src/features/draft-review/use-summarize-draft.ts` (`useSummarizeDraft(draftId)`, `useMutation`)
- Create: `client/src/features/draft-review/components/draft-summary/` (`component.tsx`, `index.ts`, `types.ts` for the props)
- Modify: `client/src/features/draft-review/component.tsx` (render `<DraftSummary />` between `DraftHeader` and `DemandCoverage`, ~37-38)
- Modify: `client/src/shared/i18n/language-key.ts` (`'Summarize this draft'`, `'Summary'`, `'Template summary'`, `'Written by AI'`, sorted)
- Test: `client/src/features/draft-review/components/draft-summary/component.test.tsx` (create)

**Interfaces:**
- Consumes: `POST /staffing-drafts/{id}/summary` → `{ summary, source }` (Task 3).
- Produces: the card shows the button; after success the text and the label (`Written by AI` for `ai`, `Template summary` for `template`). The mutation is cleared when the draft's version changes (the parent passes the version; the card resets its state when it changes). Errors use the shared toast from `shared/api/query-client.ts`.

- [ ] **Step 1: Write the failing tests** (`vi.mock` of `../../use-summarize-draft`, wrapped like the other draft-review tests):
  - "button shows the text and the AI label": click `Summarize this draft` → text from the mock result and `Written by AI`
  - "template label": `source: 'template'` → `Template summary`
  - "button disabled while pending": the button is disabled and shows busy state; a second click does not call `mutate` again
  - "text clears when the draft version changes": rerender with a new version → the text is gone
- [ ] **Step 2: Run, expect FAIL** — `cd client && npx vitest run src/features/draft-review`
- [ ] **Step 3: Implement** the endpoint, api function, hook, card and keys. Use `Card` and `Button` from `shared/ui`; Tailwind classes only.
- [ ] **Step 4: Run, expect PASS** — `cd client && npx vitest run src/features/draft-review && npm run lint && npx tsc -b`, then `make check-client` from the repo root
- [ ] **Step 5: Commit** `feat(client): summarize a staffing draft on the draft review page`

### Task 5: Docs

Written from the diff of Tasks 1–4, reviewed before commit. Server and root docs in one commit; a separate `docs(client): ...` commit only if `client/README.md` changes.

- [ ] `README.md` §15 and the AI node: draft summary is built
- [ ] `server/README.md` §9: draft summary built (route, no tools, number check, template fallback, data sent); state that customer names and emails are never sent
- [ ] `client/README.md`: the draft review screen has the Summary card
- [ ] `docs/features/README.md`: row 21 → Done; remove "draft summaries" from the Later AI item (commit only those lines if the file holds another feature's work)
- [ ] `docs/features/draft-summary.md`: status Done, tasks ticked
- [ ] `docs/user-manual.md`: the draft review section describes Summarize
- [ ] `README.md:3` and `CLAUDE.md:9`: twenty-one features built
- [ ] Drawio: "AI assistant (later)" and "LLM API (later)" drop "(later)"; intake extraction and what-if chat stay as later only
- [ ] Commit `docs: describe the draft summary`

### Task 6: Final check

- [ ] `make check` from the repo root passes
- [ ] Optional: `cd server && go test -tags=integration ./...` with Docker running
- [ ] `git log --oneline` on `feat/draft-summary` shows only this feature's commits; nothing pushed
