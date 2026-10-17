# Draft summary

**Goal:** a planner or manager opens a staffing draft and asks for a short summary of it, e.g. "4 people assigned, 20 h of QA still uncovered in 2026-W45". The text is written by the LLM and checked against the draft's numbers, with a template fallback. The AI design is in the [root README §15](../../README.md#15-ai-assistant) and [server README §9](../../server/README.md#9-ai-assistant).

**Roles:** planner, manager (the draft view: whoever can open the draft).
**Target:** 17–18 Oct · **Depends on:** [staffing-draft](staffing-draft.md), [ask-ai](ask-ai.md) · **Branch:** `feat/draft-summary` · **Status:** Done

## In scope

- A **Summarize this draft** action on the draft review page. It returns a summary of the draft as it is now.
- The model gets no tools. Its input is a small facts object built by the server from the draft.
- Every number in the model's text must appear in the facts. If one does not, the server asks once more; if it still fails, the template summary is returned.
- A template summary, plain English from the same facts. It is used when the LLM is not configured, when the provider fails, and when the number check fails twice.
- The summary is not stored and not cached. The client asks on click.

## Not now

- Summaries for scenarios.
- Streaming.
- Caching or storing the summary in the plan.
- Other languages.

## Server

**Endpoint** (see [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| POST | `/staffing-drafts/{id}/summary` | Short summary of a staffing draft | planner, manager |

No request body. The route has the heavy rate limit (`middlewares.HeavyRateLimit`) and the roles `constants.InsightRoles`. The service loads the draft through `StaffingService.Get`, so the draft access rules apply.

**Tables:** none new.

**Facts sent to the model**
- Counts: people assigned, assignments and projects.
- Total planned hours.
- Hours added and removed versus the baseline, from the draft's `Changes`.
- Per demand with unmet hours: expertise name, project name, unmet hours, first and last week.
- Whether the draft is a re-plan (`auto`).
- **No person names, no emails and no customer or client names.** Project names are allowed. Customer names and emails are never sent to the LLM.

**Rules**
- Pure functions in `server/internal/services/aisummary.go`: `buildDraftFacts`, `templateSummary` and `numbersKnown`. They have table tests.
- Number check: every number in the text (integers and decimals, including the digits in week labels such as 2026-W45) must also appear as a number in the facts JSON.
- The model is asked for at most 3 sentences, plain text, English, with no ids. The limit is set in the system prompt, and in `GenerationConfig` if the llm package has a max-token field.
- The route never returns `503`, `502` or `429` for a provider problem. It falls back to the template with `source: "template"`.
- Route path, role list, messages, source values and the prompt are constants.
- Add Swagger annotations and regenerate Swagger.

**Response** (`data`)

```json
{
  "summary": "4 people on 6 assignments across 2 projects, 312 planned hours. 20 h of QA is still uncovered on Checkout Revamp (2026-W45 to 2026-W47).",
  "source": "template"
}
```

`source` is `ai` or `template`.

## Acceptance criteria

1. The summary of a draft with known numbers returns `source: "ai"`.
2. An invented number in the model's text is retried once; if the retry is also invented, the response is the template with `source: "template"`.
3. With no LLM key configured, the response is the template with `source: "template"`, not an error.
4. A provider failure (`AI_UPSTREAM_FAILED` or rate limited) returns the template with `source: "template"`, not an error.
5. The facts sent to the model contain no person name and no email of the draft's people.
6. The check accepts digits inside week labels (2026-W45) and decimals (20.5); it rejects a number the facts do not contain, and rejects 21 when only 20 is known.
7. A specialist gets `403`.
8. The **Summary** card shows the text and a label: "Written by AI" or "Template summary". A failure shows the shared error toast.

## Client

- The **Summary** card sits on the draft review page (`client/src/features/draft-review/component.tsx`), between `DraftHeader` and `DemandCoverage`. It is a child `components/draft-summary/` of the draft-review feature.
- A **Summarize this draft** button. While the request is pending the button is disabled and shows busy state, so a second click does nothing.
- After the result: the text, and the label from `source`.
- The result stays until the draft changes; it is cleared when the draft changes (a new draft object).
- Hook `useSummarizeDraft(draftId)` (`useMutation`) and `summarizeDraft(id)` in `features/draft-review/api.ts`. The endpoint is in `client/src/resources/endpoints` (the staffing endpoints).
- No new permission: whoever can open the draft can summarize it.
- i18n keys in `LanguageKey`: `Summarize this draft`, `Summary`, `Written by AI`, `Template summary`.

## Tasks

**Server**
- [x] Facts builder, number check and template summary as pure functions
- [x] `AIService.SummarizeDraft`: prompt, number check with one retry, template fallback
- [x] Handler, DTO, route, constants, Swagger

**Client**
- [x] Endpoint, `summarizeDraft` and `useSummarizeDraft` in the draft-review feature
- [x] Summary card with the button, busy state, text and label

**Tests**
- [x] Facts: re-plan flag, unmet hours per demand, no names or emails in the marshalled facts
- [x] `templateSummary` and `numbersKnown`: the cases in acceptance criterion 6
- [x] `SummarizeDraft`: AI text with known numbers; invented number then template; no LLM key then template; provider error then template; no person name in the facts passed to the LLM
- [x] Handler and RBAC: specialist `403`
- [x] Client: button shows the text and the AI label; template label

**Docs** (same PR)
- [x] [README.md](../../README.md): §15 and the AI node show draft summary as built
- [x] [server/README.md](../../server/README.md): §9 describes the route, no tools, the number check, the template fallback and the data sent; state that customer names and emails are never sent
- [x] [client/README.md](../../client/README.md): the draft review screen
- [x] [docs/features/README.md](README.md): row 21 → Done; remove "draft summaries" from the Later AI item
- [x] [user-manual](../user-manual.md): the draft review section describes Summarize
- [x] `README.md:3` and `CLAUDE.md:9`: twenty-one features built
- [x] Drawio: the "AI assistant (later)" and "LLM API (later)" labels drop "(later)"; list intake extraction and what-if chat as later only
