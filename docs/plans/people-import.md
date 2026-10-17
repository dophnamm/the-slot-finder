# People import Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A planner adds or updates many people at once from a CSV file: new emails are created, known ones updated, all or nothing, with a downloadable template.

**Architecture:** A new server module `peopleimport` (one file per layer: constants, service, DTO, handler) copies the timesheet import's shape: read the file under a size limit, parse and check every row into `RowProblem`s, and only when there is no problem save everything in one `TxManager.Do`. Saving goes through the existing `PeopleService.Create`, `Update` and `ReplaceSkills`, whose own `tx.Do` joins the outer transaction, so `PersonCapacityChanged` and `SkillChanged` are recorded exactly as for single edits. The client gets a `people-import` feature (button + dialog) on the People page, and the timesheet's problem table moves to `shared/components` so both imports use it.

**Tech Stack:** Go (Echo v5, GORM, go-playground validator, testify), React + TypeScript (TanStack Query, Vitest, Testing Library).

**Spec:** [docs/features/people-import.md](../features/people-import.md)

## Global Constraints

- Planner only: routes use `constants.PeopleManageRoles`; the service calls `s.authz.Require(ctx, CanManagePeople)`.
- Columns, read by header name in any order. Required: `name`, `email`, `weekly_hours`, `primary_expertise`, `primary_level`. Optional: `role`, `non_project_hours`, `start_date`, `end_date` (`YYYY-MM-DD`), `other_skills` (`QA:2;Analyst:1`).
- Row checks = the person form's rules: name required, ≤ 200; email valid, ≤ 320; role one of `constants.Roles`; `weekly_hours` > 0 and ≤ 80; `non_project_hours` 0–80 and allowed against weekly hours by the existing `checkNonProjectHours`; valid dates and end ≥ start; levels 1–5; exactly one primary; no expertise twice (the primary included); expertise names matched case-insensitively; an email twice in the file; a missing required column; over `PEOPLE_IMPORT_MAX_BYTES` (1 MiB) or `PEOPLE_IMPORT_MAX_ROWS` (1000) → `413 PEOPLE_IMPORT_TOO_LARGE`.
- All or nothing: any problem → `400 PEOPLE_IMPORT_INVALID` with `data.problems` = `[{ line, column, code }]` sorted by line; nothing saved.
- New email (lower-cased) → create; empty `role` on create = `specialist`; `non_project_hours` empty on create = the existing default. Existing email → update: an empty optional cell keeps the stored value; skills are always replaced by the file's list (`primary_expertise` + `other_skills`), so an empty `other_skills` removes secondary skills; a row equal to what is stored writes nothing (skills compared with `sameSkills`). Inactive people are updated like anyone; the import never changes `active` and cannot clear an end date. People missing from the file are left alone. No Keycloak accounts.
- Response `{ created, updated, unchanged }`.
- No literals: paths, roles, messages, error codes, column names, problem codes, settings and CSV separators are constants or config. Client endpoints in `src/resources/endpoints`, permission `people:import` in `src/resources/rbac`, text through `t()` with keys sorted alphabetically in `language-key.ts` (case-insensitive, lowercase first on a tie).
- Swagger regenerated (`make swagger-load`, or `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init` with the Makefile flags), never hand-edited.
- Load the matching skills: `structuring-server-modules`, `writing-go-server-code`, `testing-go-server-code` (server); `structuring-client-components`, `translating-client-text`, `organizing-client-shared-code`, `styling-client-components` (client).
- Commits: local only on `feat/people-import`, never push. Server and client never in the same commit. Run `make check-server` / `make check-client` once before each task's last commit. Conventional subject; the message ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

1. Re-importing the same file gives `created 0, updated 0` and records no event. Test: Task 2 case "same file twice".
2. An update row whose `end_date` is before the person's stored `start_date` (start column empty) is a row problem, not a 500 at save time. Test: Task 2 case "end before the stored start".
3. A file with one bad row among good ones saves nothing. Test: Task 2 case "one bad row saves nothing".
4. Header names with extra spaces or another case (`Email`, ` weekly_hours `) are still found, and an extra unknown column is ignored. Test: Task 1 cases "header case and spaces" and "unknown column ignored".
5. A manager or specialist cannot import or download the template. Test: Task 3 RBAC rows.

---

### Task 1: Constants, settings and the row parser

**Files:**
- Create: `server/internal/constants/peopleimport.go`
- Modify: `server/internal/constants/error_codes.go` (next to the timesheet codes ~85-89), `server/internal/constants/messages.go` (next to the timesheet messages ~185-194), `server/internal/constants/routes.go` (next to `PathTimesheetImport` ~68-69), `server/internal/constants/timesheet.go` (point `TimesheetContentType` at the new shared `CSVContentType`)
- Modify: `server/internal/config/config.go` (next to `TimesheetMaxBytes`/`TimesheetMaxRows` ~202-208 and ~366-370), `server/cmd/server/.env.example` (next to the timesheet settings ~111-117)
- Create: `server/internal/services/peopleimport.go` (parser part only in this task)
- Test: `server/internal/services/peopleimport_test.go` (new `TestParsePeopleCSV_Rows`, `TestParsePeopleCSV_Problems`, `TestParsePeopleCSV_TooManyRows`)

**Interfaces:**
- Produces constants (`server/internal/constants/peopleimport.go`):

```go
const (
	PeopleImportColumnName            = "name"
	PeopleImportColumnEmail           = "email"
	PeopleImportColumnRole            = "role"
	PeopleImportColumnWeeklyHours     = "weekly_hours"
	PeopleImportColumnNonProjectHours = "non_project_hours"
	PeopleImportColumnStartDate       = "start_date"
	PeopleImportColumnEndDate         = "end_date"
	PeopleImportColumnPrimaryExpertise = "primary_expertise"
	PeopleImportColumnPrimaryLevel    = "primary_level"
	PeopleImportColumnOtherSkills     = "other_skills"
)
// PeopleImportColumns is the template's column order; PeopleImportRequiredColumns must be present.
var PeopleImportColumns = []string{ /* the ten above, in that order */ }
var PeopleImportRequiredColumns = []string{PeopleImportColumnName, PeopleImportColumnEmail, PeopleImportColumnWeeklyHours, PeopleImportColumnPrimaryExpertise, PeopleImportColumnPrimaryLevel}
const (
	PeopleImportSkillSeparator = ";"
	PeopleImportLevelSeparator = ":"
)
// Row problem codes.
const (
	PeopleImportColumnMissing    = "COLUMN_MISSING"
	PeopleImportCSVMalformed     = "CSV_MALFORMED"
	PeopleImportRowDuplicate     = "ROW_DUPLICATE"
	PeopleImportNameInvalid      = "NAME_INVALID"
	PeopleImportEmailInvalid     = "EMAIL_INVALID"
	PeopleImportRoleInvalid      = "ROLE_INVALID"
	PeopleImportHoursInvalid     = "HOURS_INVALID"
	PeopleImportDateInvalid      = "DATE_INVALID"
	PeopleImportExpertiseUnknown = "EXPERTISE_UNKNOWN"
	PeopleImportLevelInvalid     = "LEVEL_INVALID"
	PeopleImportSkillsInvalid    = "SKILLS_INVALID"
	PeopleImportSkillDuplicate   = "SKILL_DUPLICATE"
)
const (
	PeopleImportProblemsKey = "problems"
	PeopleImportFormField   = "file"
	PeopleImportFileName    = "people.csv"
	CSVContentType          = "text/csv; charset=utf-8"
)
```

- Produces error codes `PeopleImportInvalid = "PEOPLE_IMPORT_INVALID"`, `PeopleImportTooLarge = "PEOPLE_IMPORT_TOO_LARGE"`; messages `MsgPeopleImportInvalid`, `MsgPeopleImportTooLargeFormat` (bytes and rows, like `MsgTimesheetTooLargeFormat`), `MsgPeopleImported`, `MsgPeopleImportTemplate`, `MsgPeopleImportFileMissing`, `MsgPeopleImportUnreadable`; paths `PathPeopleImport = "/people/import"`, `PathPeopleImportTemplate = "/people/import-template"`.
- Produces settings `config.Config.PeopleImportMaxBytes int64` (env `PEOPLE_IMPORT_MAX_BYTES`, default `1<<20`) and `PeopleImportMaxRows int` (env `PEOPLE_IMPORT_MAX_ROWS`, default 1000), documented in `.env.example`.
- Produces (in `server/internal/services/peopleimport.go`):

```go
// peopleRow is one checked CSV row. Optional cells that were empty stay nil.
type peopleRow struct {
	Line            int
	Name            string
	Email           string // lower-cased
	Role            *string
	WeeklyHours     float64
	NonProjectHours *float64
	StartDate       *time.Time
	EndDate         *time.Time
	Skills          []SkillInput // primary first
}

type peopleImportLimits struct{ MaxRows int }

// parsePeopleCSV reads the header by name (trimmed, case-insensitive; unknown columns ignored) and checks every row
// on its own. expertises maps a lower-cased expertise name to its id. It returns errTooManyPeopleRows past MaxRows.
func parsePeopleCSV(data []byte, limits peopleImportLimits, expertises map[string]string) ([]peopleRow, []RowProblem, error)
```

Reuse the existing `RowProblem{Line, Column, Code}`. Line numbers are file lines (header = 1). A missing required column is one problem per column at line 1. Email format uses the same rule as the person DTO (`email`, max 320): validate with a `*validator.Validate` or `net/mail`, builder's choice, as long as `"not-an-email"` fails and `"a@b.co"` passes. Role must be in `constants.Roles` when present. Dates use `constants.DateLayout`. End ≥ start is checked here only when both cells are filled.

- [ ] **Step 1: Write the failing tests**
  - `TestParsePeopleCSV_Rows`: a full row (`Quinn,QUINN@x.io,manager,36,4,2026-11-02,2027-06-30,qa,3,Engineer:2;Analyst:1`) → `peopleRow{Line: 2, Name: "Quinn", Email: "quinn@x.io", Role: "manager", WeeklyHours: 36, NonProjectHours: 4, dates set, Skills: [{qa-id,3,true},{eng-id,2,false},{an-id,1,false}]}`; a row with only required cells → optional fields nil and one skill; "header case and spaces" (`Name, EMAIL ,Weekly_Hours,...`) still parses; "unknown column ignored" (an extra `team` column) parses.
  - `TestParsePeopleCSV_Problems` (table, each one problem with its line and column): missing `email` column → `{1, email, COLUMN_MISSING}`; empty name; name of 201 chars; `not-an-email`; email twice in the file (second line → `ROW_DUPLICATE`); role `boss`; `weekly_hours` `0`, `81`, `abc`; `non_project_hours` `-1`; `start_date` `2026-13-01`; end before start; `primary_expertise` `Pilot` (`EXPERTISE_UNKNOWN`); `primary_level` `6`; `other_skills` `QA` (no level → `SKILLS_INVALID`); `other_skills` `QA:9` (`LEVEL_INVALID`); `other_skills` `Pilot:2` (`EXPERTISE_UNKNOWN`); `other_skills` naming the primary (`SKILL_DUPLICATE`); a malformed CSV line (`CSV_MALFORMED`).
  - `TestParsePeopleCSV_TooManyRows`: MaxRows 2, three data rows → `errTooManyPeopleRows`.
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/services -run TestParsePeopleCSV -v`
- [ ] **Step 3: Implement** the constants, settings and `parsePeopleCSV` (with `encoding/csv`, like `parseTimesheet`).
- [ ] **Step 4: Run, expect PASS** — `cd server && go test ./internal/services ./internal/config`, then `make check-server`
- [ ] **Step 5: Commit** `feat(server): parse and check a people CSV`

### Task 2: `PeopleImportService` (import and template)

**Files:**
- Modify: `server/internal/services/peopleimport.go`
- Modify: `server/cmd/server/main.go` (FX provider, like `ProvideTimesheetService`)
- Modify: `server/internal/models/` (add `PeopleImportResult{Created, Updated, Unchanged int}` next to `ImportResult` in `models/timesheet.go`, or a new `models/peopleimport.go` — follow the module rule)
- Test: `server/internal/services/peopleimport_test.go` (new `TestPeopleImportService_Import`, `TestPeopleImportService_Template`)

**Interfaces:**
- Consumes: `parsePeopleCSV`, `peopleRow`, constants and settings (Task 1); `PeopleService.Create(ctx, CreatePersonInput)`, `Update(ctx, id, UpdatePersonInput)`, `ReplaceSkills(ctx, id, []SkillInput)`, `ListExpertises(ctx)`; `PersonRepository.ListByEmails(ctx, emails)` (returns people with skills? if not, load skills with `FindWithSkills` per matched person); `sameSkills`, `toSkillModels`, `checkNonProjectHours`, `checkDates` semantics; `db.TxManager` (`Do` joins an open transaction); `AuthzService.Require(ctx, CanManagePeople)`; `ImportInput{File io.Reader; Size int64}` (reuse the timesheet's).
- Produces:

```go
type PeopleImportService interface {
	Import(ctx context.Context, in ImportInput) (*models.PeopleImportResult, error)
	Template(ctx context.Context) ([]byte, error)
}
func ProvidePeopleImportService(people PeopleService, repo repositories.PersonRepository, authz AuthzService,
	tx db.TxManager, cfg *config.Config) PeopleImportService
```

- [ ] **Step 1: Write the failing tests** — `TestPeopleImportService_Import` (subtests; mocks as in `timesheet_test.go` and `person_test.go`; a fake or mock `PeopleService` that records calls):
  - "creates a new person": new email → `Create` called once with role `specialist` (empty cell), default non-project hours (nil), the row's skills → `{created: 1}`
  - "updates a known person": stored Quinn 40 h, file 36 h → `Update` with `WeeklyHours` 36 only (other cells empty → nil) → `{updated: 1}`
  - "a changed level replaces the skills": same fields, `primary_level` 3 → 4 → `ReplaceSkills` called, `Update` not called → `{updated: 1}`
  - "empty other_skills removes secondary skills": stored QA primary + Engineer secondary, file `other_skills` empty → `ReplaceSkills` with only the primary
  - "same file twice": every cell equals the stored person → neither `Update` nor `ReplaceSkills` nor `Create` → `{unchanged: 1}`
  - "empty end_date keeps the stored end date": `UpdatePersonInput.EndDate.Set` is false
  - "end before the stored start": stored start 2026-11-02, row `end_date` 2026-10-01, `start_date` empty → 400 `PEOPLE_IMPORT_INVALID` with `{line 2, end_date, DATE_INVALID}`, nothing saved
  - "non-project hours above the row's weekly hours": the same rule as `checkNonProjectHours` → a row problem, not a save error
  - "one bad row saves nothing": two good rows and one bad → 400, no `Create`/`Update`/`ReplaceSkills` call
  - "a file over the size limit": declared `Size` > `PeopleImportMaxBytes`, and a body larger than declared → 413 `PEOPLE_IMPORT_TOO_LARGE`
  - "too many rows" → 413
  - "a specialist is refused": `Require` fails → error, nothing read
  - "a save error rolls back": `Update` returns an error → the error is returned (a real rollback is `TxManager`'s contract)

  `TestPeopleImportService_Template`: the CSV's first row is `constants.PeopleImportColumns`; the second row is an example whose `primary_expertise` is the first expertise of `ListExpertises` and whose other cells are valid for `parsePeopleCSV` (parse the template back and expect no problem).
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/services -run TestPeopleImportService -v`
- [ ] **Step 3: Implement**
  - `Import`: `Require`; read the file under `PeopleImportMaxBytes` exactly like the timesheet's `readFile` (declared size, then `io.LimitReader(max+1)`); parse; any parse error → too large; load expertises (lower-cased name → id); load stored people by the file's emails; for update rows check the merged values (row cell or stored value) with the same rules as `Update` (`checkDates`, `checkNonProjectHours`) and turn a failure into a row problem on that line; problems → 400 sorted by line with context key `PeopleImportProblemsKey`; then one `s.tx.Do`: create, update (only the fields given and different from the stored value; `DateChange{Set: true}` only for filled date cells) and replace skills (only when `!sameSkills`), counting created / updated / unchanged.
  - `Template`: `Require`; header `PeopleImportColumns`, one example row (name `Example Person`, email `example.person@example.com`, role `specialist`, weekly 40, non-project 4, leave dates empty (the current week's Monday is not needed), primary the first expertise name, level 3, other skills empty), written with `csv.NewWriter`. Example values are constants in `constants/peopleimport.go`.
- [ ] **Step 4: Run, expect PASS** — `cd server && go test ./internal/services`, then `make check-server`
- [ ] **Step 5: Commit** `feat(server): import people from a CSV in one transaction`

### Task 3: `POST /people/import` and `GET /people/import-template`

**Files:**
- Create: `server/internal/handlers/peopleimport.go`, `server/internal/dtos/peopleimport.go`
- Modify: `server/cmd/server/routes/router.go` (next to the people routes ~216-221: `v1.GET(constants.PathPeopleImportTemplate, h.Template, authenticated, manage)` and `v1.POST(constants.PathPeopleImport, h.Import, authenticated, manage, heavy)` with `manage := middlewares.RequireRole(constants.PeopleManageRoles...)`; register them so the static paths are not shadowed by `/people/:id`), `server/cmd/server/main.go` (handler provider)
- Regenerate: `server/docs`
- Test: `server/internal/handlers/peopleimport_test.go` (new `TestPeopleImportHandler_Import`, `TestPeopleImportHandler_Template`, using the timesheet handler test's `multipartBody` style), `server/cmd/server/routes/router_test.go` (RBAC rows in `TestRouter_People_RBAC` ~244)

**Interfaces:**
- Consumes: `PeopleImportService` (Task 2).
- Produces (client relies on these): `PeopleImportResultResponse{Created int json:"created"; Updated int json:"updated"; Unchanged int json:"unchanged"}`; the 400 body `data.problems` = `[{line, column, code}]`; the template as `text/csv` with `Content-Disposition: attachment; filename="people.csv"`.

- [ ] **Step 1: Write the failing tests** — handler: "imports the uploaded file" (fake service gets the bytes, 200 with the counts); "no file" → validation error, as the timesheet handler; "row problems come back in the data" (400 with `data.problems`); template: 200, content type `constants.CSVContentType`, the attachment header. RBAC rows: manager → 403 and specialist → 403 on both routes; no token → 401.
- [ ] **Step 2: Run, expect FAIL** — `cd server && go test ./internal/handlers -run TestPeopleImportHandler -v && go test ./cmd/server/routes -run TestRouter_People_RBAC -v`
- [ ] **Step 3: Implement** — copy `TimesheetHandler.Import` / `Template` (Swagger annotations in the same style: `@Accept multipart/form-data`, `@Param file formData file true "Filled-in people template (CSV)"`, `@Success 200 {object} object{meta=dtos.Meta,data=dtos.PeopleImportResultResponse}`, `@Failure 400 {object} object{meta=dtos.Meta,data=object{problems=[]services.RowProblem}}`, `@Failure 403`, `@Failure 413`, `@Failure 429`; template `@Produce text/csv`, `@Success 200 {file} file`); routes; regenerate Swagger and check `git diff --stat server/docs` only adds the two paths and the new DTO.
- [ ] **Step 4: Run, expect PASS** — `make check-server`
- [ ] **Step 5: Commit** `feat(server): add the people import and template routes`

### Task 4: Share the problem table (client)

**Files:**
- Move: `client/src/features/timesheet-import/components/problem-table/` → `client/src/shared/components/problem-table/`
- Modify: `client/src/features/timesheet-import/` (`component.tsx`, `types.ts`, `utils.ts`, tests) to use the shared component

**Interfaces:**
- Produces: `ProblemTable` in `@/shared/components/problem-table` with props `{ problems: RowProblem[]; labels: Partial<Record<string, PlainKey>> }` (a code with no label shows the raw code, as today); `RowProblem { line: number; column: string; code: string }` exported from the shared component's `types.ts`; the timesheet feature passes its `ROW_PROBLEM_LABELS`.

- [ ] **Step 1: Move the folder and adapt the props.** The timesheet feature imports it from `@/shared/components/problem-table` and its `RowProblem` type from there (keep `RowProblemCode` in the timesheet feature).
- [ ] **Step 2: Run** — `cd client && npx vitest run src/features/timesheet-import && npm run lint && npx tsc -b`; all existing timesheet tests pass unchanged in behaviour.
- [ ] **Step 3: Commit** `refactor(client): share the import problem table`

### Task 5: Import people dialog (client)

**Files:**
- Modify: `client/src/resources/endpoints/people.ts` (`import: '/people/import'`, `importTemplate: '/people/import-template'` in `PEOPLE_ENDPOINTS`)
- Modify: `client/src/resources/rbac/permissions.ts` (`'people:import': [ROLES.PLANNER]`, next to `'timesheets:import'`)
- Create: `client/src/features/people-import/` (`component.tsx`, `index.ts`, `api.ts`, `use-people-import.ts`, `constants.ts`, `types.ts`, `utils.ts`, `component.test.tsx`)
- Modify: `client/src/pages/people/component.tsx` (the `actions` slot shows both imports a planner may use)
- Modify: `client/src/shared/i18n/language-key.ts`

**Interfaces:**
- Consumes: `POST /people/import` (multipart field `file`) → `{ created, updated, unchanged }`; 400 `PEOPLE_IMPORT_INVALID` with `details.problems`; `GET /people/import-template` → CSV blob; `ProblemTable` (Task 4); `PEOPLE_QUERY_KEY` (`@/shared/people`); `saveFile` as the timesheet uses it.
- Produces: `PeopleImport` component; `useImportPeople()` (`meta: { quietErrorCodes: [PEOPLE_IMPORT_INVALID] }`, on success invalidates `PEOPLE_QUERY_KEY` and shows the toast), `useDownloadPeopleTemplate()`.

- [ ] **Step 1: Write the failing test** (`component.test.tsx`, mocking `./use-people-import` like the timesheet tests mock theirs):
  - the button "Import people" opens the dialog with "Download template", a file picker and "Import"
  - on a `PEOPLE_IMPORT_INVALID` error the problems table shows line, column and the readable label of `EMAIL_INVALID`
  - on success the toast text is `Created 3, updated 2, unchanged 14`
- [ ] **Step 2: Run, expect FAIL** — `cd client && npx vitest run src/features/people-import`
- [ ] **Step 3: Implement** — copy the timesheet import's dialog shape without its week range; keys `'Import people'`, `'Created {created}, updated {updated}, unchanged {unchanged}'` and one readable label per problem code (`ROW_PROBLEM_LABELS: Record<PeopleImportProblemCode, PlainKey>`, reusing an existing key's text where the timesheet already has the same meaning); template file name from a constant; accept `.csv,text/csv`.
- [ ] **Step 4: Run, expect PASS** — `cd client && npx vitest run src/features/people-import src/features/timesheet-import && npm run lint && npx tsc -b`, then `make check-client`
- [ ] **Step 5: Commit** `feat(client): import people from a CSV`

### Task 6: Docs

Written by the docs agent from the diff, reviewed before commit.

- [ ] `README.md`: the integrations node and any `/imports` or import-job text: people import is built and synchronous (`POST /people/import`); HR import stays a CSV export from HR, no live HR connection
- [ ] `server/README.md`: the routes table gets both routes (planner); remove the Later `POST /imports/*` bullets and the "Later · import" job row, or reword them to point at `POST /people/import`; the config table gets `PEOPLE_IMPORT_MAX_BYTES` and `PEOPLE_IMPORT_MAX_ROWS`; the integrations row ("HR import")
- [ ] `client/README.md`: the People screen gets the Import people dialog
- [ ] `docs/features/people-and-skills.md`: drop the "CSV and HR import → Later" bullet
- [ ] `docs/features/README.md`: row 20 → Done; remove "CSV import" from the Later list
- [ ] `docs/features/people-import.md`: status Done, tasks ticked
- [ ] `docs/user-manual.md`: the People section explains the import (template, columns, all or nothing, update by email, delete the example row before importing)
- [ ] `README.md:3` and `CLAUDE.md:9`: twenty features built (sign-in through people import)
- [ ] Commit `docs: describe the people import`

### Task 7: Final check

- [ ] `make check` from the repo root passes
- [ ] If Docker runs: `cd server && go test -tags=integration ./internal/services -run 'Timesheet|People'` passes
- [ ] `git log --oneline feat/bench-list..HEAD` shows only this feature's commits; nothing pushed
