# Skill and program events Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Record `SkillChanged` when a person's skills change (and re-plan) and `ProgramSlipped` when a program's end date moves (change history only).

**Architecture:** Both events go through the existing outbox: the service calls `s.events.Record(ctx, ChangeInput{...})` inside the same `s.tx.Do` as the save. The re-plan runner treats `SkillChanged` like `PersonCapacityChanged`. The engine already drops unlocked hours of someone who no longer has the skill (`keepFeasible`, `server/internal/engine/greedy.go:119`), so no engine change. The client's shared `EventSummary` gets one line per new type.

**Tech Stack:** Go (Echo v5, GORM, testify), React + TypeScript (Vitest, Testing Library).

**Spec:** [docs/features/skill-and-program-events.md](../features/skill-and-program-events.md)

## Global Constraints

- No literals: event types, payload json keys and messages live in `server/internal/constants`; client text goes through `t()` with keys in `client/src/shared/i18n/language-key.ts` (sorted alphabetically).
- Events are recorded inside the same `s.tx.Do` as the save. A failed `Record` fails the save and nothing is written.
- Error flow: repository errors are wrapped by the service (`errors.DatabaseError(...)`), handlers call `HandleError`.
- Swagger is regenerated (`make swagger-load`, or `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init` with the flags of the `swagger-load` target in `server/Makefile`), never edited by hand.
- Client dates only through `@/utils/functions` (`toIsoWeek`, `formatDate`); component rules from the `structuring-client-components` skill.
- Load the matching skill before each side: `writing-go-server-code`, `testing-go-server-code`, `structuring-server-modules` (server); `translating-client-text`, `structuring-client-components` (client).
- Commits: local only on `feat/skill-and-program-events`, never push. Server and client never in the same commit. Every commit passes its side's checks (`make check-server` or `make check-client`). Conventional subject, and the message ends with the trailer `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

1. The same skills saved in another order record no `SkillChanged` (the comparison ignores order). Test: Task 1, case "same skills, other order".
2. A skills save rejected by `checkSkills` (two primaries, unknown expertise) records nothing and saves nothing. Test: Task 1, case "rejected skills record nothing".
3. A program edit that changes only the name or the start date records no `ProgramSlipped`; an end date moved earlier does record it (same as `ProjectSlipped`). Test: Task 3 cases.
4. `SkillChanged` for a person with no baseline hours ends `no_impact` unless some project is understaffed; it never fails the runner. Test: Task 2, eventScope case "skill change, nothing understaffed".
5. An event row whose payload misses a field (older rows) still renders without crashing: every payload read uses `?? ''`. Test: Task 4, case "missing payload fields".

---

### Task 1: `SkillChanged` recorded by `ReplaceSkills`

**Files:**
- Modify: `server/internal/constants/events.go` (add the type)
- Modify: `server/internal/services/event.go` (payload types, next to `PersonCapacityChange` at ~57-61)
- Modify: `server/internal/services/person.go` (`ReplaceSkills` at ~236-249; new helpers near `capacityFieldsChanged` at ~406)
- Test: `server/internal/services/event_test.go` (new `TestPeopleService_RecordsSkillChange`, next to `TestPeopleService_RecordsCapacityChange` at ~297)
- Test: `server/internal/services/person_test.go` (new `TestSameSkills`; create the file if missing)

**Interfaces:**
- Produces: `constants.EventSkillChanged = "SkillChanged"`
- Produces:

```go
// SkillState is one skill in a SkillChange payload.
type SkillState struct {
	Expertise string `json:"expertise"`
	Level     int    `json:"level"`
	Primary   bool   `json:"primary"`
}

// SkillChange is the payload of SkillChanged.
type SkillChange struct {
	PersonID   string       `json:"person_id"`
	PersonName string       `json:"person_name"`
	Before     []SkillState `json:"before"`
	After      []SkillState `json:"after"`
}
```

- Produces: `func sameSkills(a, b []models.PersonSkill) bool` — true when both have the same set of `ExpertiseID`, each with the same `Level` and `IsPrimary`; order ignored. people-import reuses it.
- Produces: `func skillStates(skills []models.PersonSkill) []SkillState` — uses the preloaded `Expertise.Name`, sorted by expertise name so the payload is stable.
- Consumes: `repo.FindWithSkills(ctx, id)`, `repo.ReplaceSkills(ctx, personID, []models.PersonSkill)`, `toSkillModels(id, skills)`, `s.checkSkills(ctx, skills, opReplaceSkills)`, `s.events.Record`.

- [ ] **Step 1: Write the failing tests**

`TestSameSkills` (table): both empty → true; same skills same order → true; same skills other order → true; one level differs → false; primary moved to another expertise → false; one extra expertise → false.

`TestPeopleService_RecordsSkillChange` (table, same style and mocks as `TestPeopleService_RecordsCapacityChange`: `MockPersonRepository`, `passthroughTx{}`, `&recordedEvents{}`, `ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), events)`; mock whatever `checkSkills` reads as the existing ReplaceSkills tests do). `FindWithSkills` returns the person with the "before" skills first, then (`.Once()` order) the "after" skills. Expertises: Engineer (`eng`) and QA (`qa`).
  - "level changed": before `[{eng,3,primary}]`, input `[{eng,4,primary}]` → one event, `Type == constants.EventSkillChanged`, payload `SkillChange{PersonID: personID, PersonName: "Alex", Before: [{Engineer,3,true}], After: [{Engineer,4,true}]}`
  - "expertise removed": before `[{eng,3,true},{qa,2,false}]`, input `[{eng,3,true}]` → one event
  - "primary moved": before `[{eng,3,true},{qa,2,false}]`, input `[{eng,3,false},{qa,2,true}]` → one event
  - "same skills, other order": before `[{eng,3,true},{qa,2,false}]`, input `[{qa,2,false},{eng,3,true}]` → no event
  - "rejected skills record nothing": input with two primaries → error, `repo.ReplaceSkills` not called, no event
  - "a failed event fails the save": `events.fail = errors.New("boom")` → `ReplaceSkills` returns an error

- [ ] **Step 2: Run them, expect FAIL**

Run: `cd server && go test ./internal/services -run 'TestSameSkills|TestPeopleService_RecordsSkillChange' -v`
Expected: FAIL (undefined `sameSkills`, `SkillChange`, `constants.EventSkillChanged`).

- [ ] **Step 3: Implement**

Add the constant, the two payload types and the two helpers. In `ReplaceSkills`: keep `before, err := s.Get(ctx, id)` (it loads skills with expertise), keep `checkSkills`, then inside one `s.tx.Do`: `repo.ReplaceSkills`; if `sameSkills(before.Skills, toSkillModels(id, skills))` return nil; else load `after` with `repo.FindWithSkills(ctx, id)` and `s.events.Record(ctx, ChangeInput{Type: constants.EventSkillChanged, Payload: SkillChange{PersonID: id, PersonName: before.Name, Before: skillStates(before.Skills), After: skillStates(after.Skills)}})`. Keep the existing error wrapping (`errors.DatabaseError(constants.MsgPersonUpdateFailed, err)...`).

- [ ] **Step 4: Run them, expect PASS**

Run: `cd server && go test ./internal/services -run 'TestSameSkills|TestPeopleService_RecordsSkillChange|TestPeopleService' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

`feat(server): record SkillChanged when a person's skills change`

### Task 2: Re-plan scope for `SkillChanged`

**Files:**
- Modify: `server/internal/services/replan.go` (`touchedProjects` ~407-440, `(*draftBuilder).eventScope` ~471-506)
- Test: `server/internal/services/replan_test.go` (`TestTouchedProjects` ~53, `TestReplanService_RunOnce_EventScope` ~217)
- Test: `server/internal/engine/engine_test.go` (new `TestPropose_ReplanDropsHoursOfSomeoneWhoLostTheSkill`, next to `TestPropose_ReplanKeepsWhatStillFits` ~237)

**Interfaces:**
- Consumes: `constants.EventSkillChanged` and the `person_id` payload key (Task 1), `projectsOfPersonFrom`, `isoweek.Of(e.OccurredAt)`, `payloadPersonID`.

- [ ] **Step 1: Write the failing tests**
  - `TestTouchedProjects`: a `SkillChanged` event for person P with no `ProjectIDs` touches P's projects from the event week on (same expectation as the existing `PersonCapacityChanged` case).
  - `TestReplanService_RunOnce_EventScope`: case "skill change opens understaffed projects" (an understaffed project not named by the event is planned, as for `PersonCapacityChanged`); case "skill change, nothing understaffed" → event status `constants.EventStatusNoImpact`, no draft.
  - `TestPropose_ReplanDropsHoursOfSomeoneWhoLostTheSkill` (fixture style of `TestPropose_ReplanKeepsWhatStillFits`): demand QA min level 2; `Existing` holds person P with unlocked hours on it plus one locked week; P's skills in the request no longer include QA; person Q has QA 3 and free hours. Assert: P keeps only the locked week; the unlocked hours go to Q.

- [ ] **Step 2: Run, expect FAIL** (the engine test may already PASS: that confirms the spec; keep it as a regression test)

Run: `cd server && go test ./internal/services -run 'TestTouchedProjects|TestReplanService_RunOnce_EventScope' -v && go test ./internal/engine -run TestPropose_ReplanDropsHoursOfSomeoneWhoLostTheSkill -v`

- [ ] **Step 3: Implement**

In `touchedProjects`, handle `constants.EventSkillChanged` exactly like `constants.EventPersonCapacityChanged`. In `eventScope`, `case constants.EventPersonCapacityChanged, constants.EventSkillChanged: freed = true`.

- [ ] **Step 4: Run, expect PASS**

Run: `cd server && go test ./internal/services ./internal/engine`

- [ ] **Step 5: Commit**

`feat(server): re-plan on SkillChanged`

### Task 3: `ProgramSlipped` and the full event type enum

**Files:**
- Modify: `server/internal/constants/events.go`
- Modify: `server/internal/services/event.go` (payload next to `DeadlineChange` ~83-88)
- Modify: `server/internal/services/program.go` (`Update` ~170-199)
- Modify: `server/internal/dtos/event.go:31` (enum), then regenerate `server/docs`
- Test: `server/internal/services/event_test.go` (new `TestProgramService_RecordsEndDateChange`)
- Test: `server/internal/services/replan_test.go` (`TestReplanService_RunOnce_EventScope`, one case)

**Interfaces:**
- Produces: `constants.EventProgramSlipped = "ProgramSlipped"`
- Produces:

```go
// ProgramEndChange is the payload of ProgramSlipped.
type ProgramEndChange struct {
	ProgramID  string `json:"program_id"`
	Program    string `json:"program"`
	OldEndDate string `json:"old_end_date"`
	NewEndDate string `json:"new_end_date"`
}
```

Dates use `constants.DateLayout` (`2006-01-02`), as `DeadlineChange` does.

- Produces (client relies on it): the API enum lists `AbsenceAdded, AbsenceRemoved, PhaseSlipped, ProjectSlipped, ProjectStatusChanged, DemandChanged, PersonCapacityChanged, TeamChanged, SkillChanged, ProgramSlipped`.

- [ ] **Step 1: Write the failing tests**

`TestProgramService_RecordsEndDateChange` (table; mocks as in the existing program service tests, `passthroughTx{}`, `&recordedEvents{}`). Program "Checkout Revamp", end date 2026-10-30:
  - "end moved later" (2026-11-13) → one event `constants.EventProgramSlipped`, payload `ProgramEndChange{ProgramID, Program: "Checkout Revamp", OldEndDate: "2026-10-30", NewEndDate: "2026-11-13"}`, `ProjectIDs` empty
  - "end moved earlier" (2026-10-23) → one event
  - "same end date" → no event
  - "only the name changes" → no event
  - "only the start date changes" → no event
  - "a failed event fails the update" → `Update` returns an error

`TestReplanService_RunOnce_EventScope`: case "program slip opens nothing" → status `constants.EventStatusNoImpact`, no draft.

- [ ] **Step 2: Run, expect FAIL**

Run: `cd server && go test ./internal/services -run 'TestProgramService_RecordsEndDateChange|TestReplanService_RunOnce_EventScope' -v`

- [ ] **Step 3: Implement**

In `programService.Update`: keep `oldEnd := program.EndDate` before `applyProgramChanges`; wrap `programs.Save` and the record in one `s.tx.Do`; record only when `!program.EndDate.Equal(oldEnd)`. Keep the current error mapping for a failed save. No `eventScope` case for `ProgramSlipped`.

- [ ] **Step 4: Run, expect PASS**

Run: `cd server && go test ./internal/services`

- [ ] **Step 5: Commit** `feat(server): record ProgramSlipped when a program's end date moves`

- [ ] **Step 6: Enum and Swagger**

Set the `enums:` tag at `dtos/event.go:31` to the ten types above; run `cd server && make swagger-load`; check `git diff --stat server/docs` changes only the event type enum. Run `make check-server` from the repo root; expect it to pass.

- [ ] **Step 7: Commit** `chore(server): list every change event type in the API docs`

### Task 4: Change history lines (client)

**Files:**
- Modify: `client/src/shared/events/types.ts` (`ChangeEventType` line 2, `ChangeEventPayload` 6-17)
- Modify: `client/src/shared/components/event-summary/component.tsx` (switch at 6-50)
- Modify: `client/src/shared/i18n/language-key.ts` (two keys, sorted)
- Test: `client/src/shared/components/event-summary/component.test.tsx` (create)

**Interfaces:**
- Consumes: payload keys from Tasks 1 and 3: `person_name`; `program`, `old_end_date`, `new_end_date`.
- Produces: `ChangeEventType` includes `'SkillChanged' | 'ProgramSlipped'`; `ChangeEventPayload` gains `program?: string`, `old_end_date?: string`, `new_end_date?: string`.

- [ ] **Step 1: Write the failing test** (`component.test.tsx`, Vitest + Testing Library)
  - "skill change": `{ type: 'SkillChanged', payload: { person_name: 'Quinn' } }` → text `Quinn: skills changed`
  - "program slip": `{ type: 'ProgramSlipped', payload: { program: 'Checkout Revamp', old_end_date: '2026-10-30', new_end_date: '2026-11-13' } }` → text `Checkout Revamp: end moved from 2026-W44 to 2026-W46`
  - "missing payload fields": `{ type: 'ProgramSlipped', payload: {} }` renders without throwing

- [ ] **Step 2: Run, expect FAIL**

Run: `cd client && npx vitest run src/shared/components/event-summary/component.test.tsx`

- [ ] **Step 3: Implement**

Keys: `'{person}: skills changed'` and `'{program}: end moved from {from} to {to}'`. `SkillChanged` → `t('{person}: skills changed', { person: payload.person_name ?? '' })`. `ProgramSlipped` → `t('{program}: end moved from {from} to {to}', { program: payload.program ?? '', from: payload.old_end_date ? toIsoWeek(payload.old_end_date) : '', to: payload.new_end_date ? toIsoWeek(payload.new_end_date) : '' })`, with `toIsoWeek` from `@/utils/functions`. Follow the `translating-client-text` skill for any other locale file it requires.

- [ ] **Step 4: Run, expect PASS**

Run: `cd client && npx vitest run src/shared/components/event-summary && npm run lint && npx tsc -b`

- [ ] **Step 5: Commit** `feat(client): show SkillChanged and ProgramSlipped in the change history`

### Task 5: Docs

Written by the docs agent from the diff of Tasks 1–4, reviewed before commit. Server-side docs ride in a server commit; client-side docs in a client commit.

- [ ] `README.md`: the domain-events edge label in the module diagram and the `CHANGE_EVENT.type` column list `SkillChanged` and `ProgramSlipped` as built; remove `PersonLeft`
- [ ] `server/README.md`: the people row (`PersonJoined`, `PersonLeft`, `SkillChanged`: later) → `SkillChanged` built, `PersonJoined` and `PersonLeft` gone; `PATCH /programs/{id}` records `ProgramSlipped`; `PUT /people/{id}/skills` records `SkillChanged`
- [ ] `docs/features/replan-on-change.md`: replace the `PersonLeft` bullet with one line: a person leaving is an `end_date` or `active` change, which records `PersonCapacityChanged`
- [ ] `docs/architecture.drawio`: event labels (XML stays valid)
- [ ] `docs/user-manual.md`: only if it lists what triggers a re-plan
- [ ] `docs/features/README.md` row 18 → Done; tick the spec's tasks
- [ ] Commit `docs: describe SkillChanged and ProgramSlipped` (server and root docs); a separate `docs(client): ...` commit only if `client/README.md` changes

### Task 6: Final check

- [ ] `make check` from the repo root passes (lint, test, build, both sides)
- [ ] If Docker runs: `cd server && go test -tags=integration ./internal/services -run 'Replan|Outbox'` passes
- [ ] `git log --oneline main..HEAD` shows only this feature's commits on top of the docs cleanup; nothing pushed
