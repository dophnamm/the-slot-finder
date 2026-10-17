# People import

**Goal:** a planner adds or updates many people at once from a CSV file exported from HR. This is the CSV import deferred in [people-and-skills](people-and-skills.md#not-now).

**Roles:** planner.
**Target:** 17–18 Oct · **Depends on:** [people-and-skills](people-and-skills.md), [skill-and-program-events](skill-and-program-events.md) (for `SkillChanged` on update) · **Branch:** `feat/people-import` · **Status:** Planned

## In scope

- **Template.** `GET /people/import-template` returns a CSV with the header row and one example row. The column names come from the same constants the import reads, so there is one source of truth.
- **Import.** The planner uploads the filled-in file (`multipart/form-data`, field `file`, as in the [timesheet import](timesheet-import.md)). The import runs synchronously in the request. There is no background job and no import table.
- **Columns.** Read by header name, in any order. Constants live in `internal/constants`.
  - Required: `name`, `email`, `weekly_hours`, `primary_expertise`, `primary_level` (1–5).
  - Optional: `role` (planner, manager or specialist), `non_project_hours`, `start_date` and `end_date` (YYYY-MM-DD), `other_skills` (for example `QA:2;Analyst:1`: expertise name, colon, level, separated by semicolons).
- **Checks.** Every row is checked before anything is saved. The same rules as creating a person apply (`CreatePersonInput` at `server/internal/services/person.go:51-60`, `CreatePersonRequest` at `server/internal/dtos/person.go:13-27`, `Create` at `person.go:142-175`):
  - name at most 200 characters; a valid email of at most 320 characters
  - a known role
  - `weekly_hours` above 0 and at most 80; `non_project_hours` from 0 to 80
  - valid dates, and end on or after start
  - levels from 1 to 5; exactly one primary; no expertise twice, the primary included
  - expertise names matched to existing expertises, case-insensitively
  - also: a missing required column, an email that appears twice in the file, and a file over the size or row limit
- **All or nothing.** Any problem means nothing is saved. Every problem is reported.
- **Save.** One `TxManager.Do` for the whole file:
  - a new email (lower-cased) creates the person
  - an existing email updates that person
  - on update, an empty optional cell keeps the current value (`role`, `non_project_hours`, `start_date`, `end_date`)
  - the skills are always replaced by the file's list (`primary_expertise` and `other_skills`), so an empty `other_skills` removes the secondary skills
  - a row that changes nothing writes nothing. A row is unchanged when every filled-in field equals the stored value and the skill set is equal under the comparison `SkillChanged` uses ([skill-and-program-events](skill-and-program-events.md)): order does not matter.
  - an inactive person is updated like anyone else. The import never changes `active` (there is no column for it), and it cannot clear an end date (an empty cell keeps it); both are done on the person form.
  - `role` follows the same rules as the person update service. Access comes from the Keycloak token roles, so the column grants no access by itself. A new manager joins no team until a planner adds them, as today.
- **Events.** Updates record `PersonCapacityChanged` and `SkillChanged` in the same transaction, exactly as the single-person paths do. The builder may split the create, update and skills logic into inner methods that run inside the caller's transaction, as long as the checks and events stay the same.
- **Keycloak.** The import creates no Keycloak accounts. A person is linked on first sign-in by verified email, as today (`person.go:301-310`).
- **Limits.** `PEOPLE_IMPORT_MAX_BYTES` (default 1 MiB) and `PEOPLE_IMPORT_MAX_ROWS` (default 1000).

## Not now

- Excel files.
- A live connection to an HR system.
- Deactivating people who are missing from the file. People missing from the file are left alone.
- Creating Keycloak accounts.

## Server

**Endpoints** (see [server/README §3](../../server/README.md#3-api-restjson))

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/people/import-template` | The CSV template (`text/csv`, downloaded as a file) | planner |
| POST | `/people/import` | Upload a filled-in file. Returns the counts of created, updated and unchanged people | planner |

**Tables:** none new.

**Rules**
- Each problem is `{ line, column, code }`. Codes are constants in `internal/constants`. The builder names them.
- A file with any problem gets `400` with error code `PEOPLE_IMPORT_INVALID` and the list of problems in `data.problems`. Nothing is saved.
- A file over `PEOPLE_IMPORT_MAX_BYTES` or `PEOPLE_IMPORT_MAX_ROWS` gets `413 PEOPLE_IMPORT_TOO_LARGE`, as the timesheet import does.
- An empty `role` on a row that creates a person means `specialist`. On a row that updates a person it keeps the current role.
- Settings go in `cmd/server/.env.example` and [server/README §8](../../server/README.md#8-settings).

**Response** (`data`)

```json
{"created": 3, "updated": 2, "unchanged": 14}
```

## Acceptance criteria

1. A file with a new email creates that person with the file's skills and dates.
2. A file that changes one existing person's level updates them and records `SkillChanged`. Their unchanged rows write nothing.
3. A file with an unknown expertise on line 5 and a duplicate email on line 7 saves nothing and lists both problems.
4. An empty `other_skills` removes the person's secondary skills.
5. An empty `end_date` on update keeps the person's current end date.
6. Importing the same file twice gives `created: 0` and `updated: 0` the second time, and records no events.
7. A manager or a specialist gets `403` on both routes.
8. A file over `PEOPLE_IMPORT_MAX_ROWS` is refused.
9. The import creates no Keycloak account. A person who signs in with the same verified email is linked to their row.

## Client

- People page, planner only (`people:import` in `src/resources/rbac/permissions.ts`): an **Import people** button opens a dialog like **Import timesheet**.
  - **Download template** link, a file picker, and an **Import** button.
  - On `400`, the dialog shows the problems table: line, column, and a readable message for each code.
  - On success, a toast: "Created {created}, updated {updated}, unchanged {unchanged}". The people queries are invalidated.
- The problems table lives in `client/src/features/timesheet-import/components`. It cannot be imported from another feature. The builder decides between moving it to `src/shared/components` and using it from both features, or copying it. Moving is preferred.
- Endpoints in `src/resources/endpoints/people.ts`. i18n keys in `LanguageKey`.

## Tasks

**Server**
- [ ] Constants: paths, role list, messages, error codes, CSV column names, settings
- [ ] CSV parser with a row check per problem code
- [ ] Service: build the template; check every row; save in one `TxManager.Do` through the existing person update and create logic
- [ ] Handlers (CSV download, multipart upload), routes, Swagger
- [ ] Settings `PEOPLE_IMPORT_MAX_BYTES` and `PEOPLE_IMPORT_MAX_ROWS`

**Client**
- [ ] Endpoints, types and hooks through `shared/api/http.ts`
- [ ] Import people dialog with the template link, file picker and problems table
- [ ] Problems table moved to `src/shared/components`, or a copy, as the builder decides

**Tests**
- [ ] Parser table tests: each problem code
- [ ] Service: create, update, unchanged, one bad row saves nothing, events recorded on update
- [ ] Handler and RBAC: manager and specialist `403`; size limit `413`
- [ ] Client: the problems table on `400`; the success toast

**Docs** (same PR)
- [ ] [README.md](../../README.md): the integrations node (importers now synchronous `POST /people/import`, not later); any `/imports` or import job text
- [ ] [server/README.md](../../server/README.md): the routes table; remove the Later `POST /imports/*` bullet and the "Later · import" job row, or reword them; the config table; the integrations row ("HR import")
- [ ] [client/README.md](../../client/README.md): the People screen
- [ ] [people-and-skills](people-and-skills.md): the "CSV and HR import" bullet under Not now
- [ ] [user-manual](../user-manual.md): the People section
