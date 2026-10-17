# Skill and program events

**Goal:** two more change events. A change to a person's skills triggers an automatic re-plan. A move of a program's end date is recorded in the change history. Both use the outbox from [replan-on-change](replan-on-change.md).

**Roles:** planner (skills); planner and manager of the program (end date).
**Target:** 17–18 Oct · **Depends on:** [people-and-skills](people-and-skills.md), [replan-on-change](replan-on-change.md) · **Branch:** `feat/skill-and-program-events` · **Status:** Planned

## In scope

- **`SkillChanged`.** `PersonService.ReplaceSkills` (`server/internal/services/person.go:236-249`, `PUT /people/{id}/skills`) records it in the same `TxManager.Do` as the save.
  - Recorded only when the skill set really changes: an expertise added or removed, a level changed, or the primary changed. Saving the same skills records nothing.
  - Payload: person id, person name, skills before and after (expertise name, level, primary).
  - Re-plan scope: the person's projects from the event week on, as `PersonCapacityChanged` does in `touchedProjects` (`server/internal/services/replan.go:407-440`). Plus `freed = true` in `eventScope` (`replan.go:471-506`), so every understaffed project is reconsidered. The person may now qualify for it.
- **`ProgramSlipped`.** The program update (`server/internal/services/program.go:170-199`, `PATCH /programs/{id}`) records it only when `end_date` changes.
  - Recorded in the same transaction as the save. Wrap the save in `TxManager.Do` if it is not already.
  - Payload: program id, program name, old end date, new end date.
- **Event type list.** The enum in `server/internal/dtos/event.go:31` lists only four types today. List every type, then regenerate Swagger.
- **Change history.** The client shows a readable line for both new types, for example "Skills changed: Quinn" and "Program end moved: Checkout Revamp 2026-W44 → 2026-W46". The builder finds the component that renders event types and adds the lines, with i18n keys in `LanguageKey`.

## Not now

- `PersonJoined`.
- A re-plan on a program end date. No planning reads the program end date today: project windows come from each project's own start and deadline (`server/internal/services/project.go:676-696`). `eventScope` gets no case for `ProgramSlipped`, so the event only shows in the change history. The runner may still build a draft if real conflicts already exist, as it does for any event.
- `PersonLeft`: not built, and not needed. A person leaving is an `end_date` or `active` change. That already records `PersonCapacityChanged` and re-plans (`person.go:406-424`, `capacityFieldsChanged`).

## Server

**Endpoints** (see [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| PUT | `/people/{id}/skills` | Replace a person's skills; records `SkillChanged` when they change | planner |
| PATCH | `/programs/{id}` | Edit a program; records `ProgramSlipped` when the end date changes | planner, manager (own) |

**Tables:** none new. Both events go in the existing change event table.

**Rules**
- Both events are recorded through `events.Record` inside the same `TxManager.Do` as the save. A failed save records nothing.
- `SkillChanged` compares the before and after skill sets with one function: order does not matter, and two sets are equal when every expertise has the same level and the same primary flag. [people-import](people-import.md) reuses this function to decide "unchanged".
- `eventScope` opens every understaffed project for `SkillChanged`, as it does for `PersonCapacityChanged`. Only the person's projects from the event week on are named.
- The engine already drops unlocked hours for a person who no longer qualifies: `keepFeasible` (`server/internal/engine/greedy.go:119`) skips an existing allocation when the person lacks the expertise at `min_level` in the snapshot, and the snapshot is built from current skills. No engine change is expected; a test proves acceptance criterion 3.
- `ProgramSlipped` is recorded whenever the `end_date` value changes, even inside the same week. The payload stores dates (YYYY-MM-DD); the client shows them as week labels with the week helpers in `@/utils/functions`.
- `ProgramSlipped` has no case in `eventScope`, so it opens no projects.
- The event type enum in `dtos/event.go` lists all types. Regenerate Swagger with `make swagger-load`. Never edit `docs/` by hand.

## Client

- The change history shows one readable line for `SkillChanged` and one for `ProgramSlipped`.
- The new strings are in `LanguageKey`.

## Acceptance criteria

1. Changing one person's level on one expertise records one `SkillChanged` with the skills before and after. Saving the same skills again records nothing.
2. Moving a program's end date records one `ProgramSlipped` with the old and new end dates. Saving the same end date records nothing.
3. A person loses an expertise, or drops below a demand's `min_level`, while holding unlocked baseline hours on that demand. The auto re-plan draft no longer gives them those hours: they go to someone else or to `unmet`. Locked allocations stay as they are.
4. A person gains an expertise. An understaffed project that needs it is reconsidered in the draft.
5. A program end-date move shows in the change history. It opens no project, so the runner ends it as `no_impact`, unless real conflicts exist.

## Tasks

**Server**
- [ ] Event constants and payload types for `SkillChanged` and `ProgramSlipped`
- [ ] `ReplaceSkills` diffs the skill set and records `SkillChanged` in `TxManager.Do`
- [ ] Program update wraps the save in `TxManager.Do` and records `ProgramSlipped` on an end-date change
- [ ] `touchedProjects` and `eventScope` handle `SkillChanged`
- [ ] One skill-set comparison function, shared with people-import
- [ ] Event type enum in `dtos/event.go` lists every type; regenerate Swagger

**Client**
- [ ] Change history lines for both types, with i18n keys in `LanguageKey`

**Tests**
- [ ] Service: `SkillChanged` recorded on a real change, nothing on an identical save
- [ ] Service: `ProgramSlipped` recorded on an end-date change, nothing on an identical save
- [ ] `eventScope` test for `SkillChanged`
- [ ] Re-plan test: a lost expertise moves the unlocked hours
- [ ] Client: the history lines for both types

**Docs** (same PR)
- [ ] [README.md](../../README.md): the event list in the module diagram edge label and the CHANGE_EVENT `type` column. Move `SkillChanged` and `ProgramSlipped` to built; remove `PersonLeft`
- [ ] [server/README.md](../../server/README.md): the people row (`PersonJoined`, `PersonLeft`, `SkillChanged` later) and `PATCH /programs/{id}` (`ProgramSlipped` later)
- [ ] [replan-on-change](replan-on-change.md): the `PersonLeft` bullet under Not now
- [ ] `docs/architecture.drawio`: the event labels
- [ ] [user-manual](../user-manual.md): only if it lists what triggers a re-plan
