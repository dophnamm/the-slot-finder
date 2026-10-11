# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

The Slot Finder assigns incoming customer projects to cross-functional squads (PM, Analyst, Engineer, QA), forecasts who is overloaded or idle per ISO week, and re-plans when projects slip or people take leave. `README.md` is the architecture source of truth: goals, data model, metrics formulas, flows and ADRs. `docs/architecture.drawio` mirrors its diagrams. `server/README.md` and `client/README.md` hold the per-side details, API table and screens. `docs/features/` holds one spec per feature in build order (scope, acceptance criteria, tasks); read the feature's spec before building it.

Status: sign-in, people-and-skills, programs-and-demand and staffing-draft are built; `docs/features/README.md` tracks the rest (heatmap, re-plan, metrics).

## Repo-wide rules

- All docs, diagrams, code comments and UI text are in **English**.
- **Never hardcode**, in client or server. URLs, routes, roles, thresholds and formats each have one named definition (see the conventions below).
- When the user adds a convention, record it in the matching skill under `.claude/skills/`, and enforce it with lint where possible.

## Commands

Run `make` at the repo root to list the shortcuts:

| Task | Command |
|---|---|
| First run: `.env` files and dependencies | `make setup` |
| Whole stack in Docker: Postgres, Redis, Keycloak, OTel, Jaeger, migrations, API on `:3000` | `make up` · stop with `make down` · logs with `make logs` |
| Infra in Docker, server and client on the host | `make dev` (or `make infra`, then `make server` / `make client`) |
| Lint, test and build both sides (CI runs it on every pull request: `.github/workflows/check.yml`) | `make check` |
| Load demo data (expertises, 12 people with skills, leave, holidays, 2 programs with 6 projects; safe to run again) | `make seed` |
| Wipe local data (DB and Keycloak volumes) | `make reset-data` |

Compose lives in `server/docker-compose.yml` and reads `server/cmd/server/.env`. The `server` service overrides the hosts so they point at containers. The `migrate` service runs `atlas migrate apply` before the API starts.

### Server (`cd server`)

| Task | Command |
|---|---|
| First setup | `cp cmd/server/.env.example cmd/server/.env` |
| Local infra (Postgres, Redis, Keycloak, OTel collector, Jaeger) | `make container-up` |
| Run API on `:3000` (Swagger at `/swagger/index.html`, non-prod) | `make up` (runs from `cmd/server`, so `.env` there is loaded) |
| Build / vet | `go build ./...` · `go vet ./...` |
| Unit tests | `go test ./...` · CI-like: `make tests` (race, coverage, count=1) |
| Single test | `go test ./internal/services -run TestName -v` |
| Integration tests (testcontainers, needs Docker) | `go test -tags=integration ./...` |
| Lint | `make lint` (golangci-lint v2 via `go run`) |
| Regenerate Swagger after changing handler annotations | `make swagger-load` (needs `swag`), or `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init` with the same flags as the Makefile |
| New migration from GORM models (needs Atlas CLI + Docker) | `make migrate-generate name=<name>`, then `make migrate-up` |

### Client (`cd client`)

| Task | Command |
|---|---|
| Dev server on `:5173` (proxies `/api` to `:3000`) | `npm run dev` |
| Type-check | `npx tsc -b` |
| Lint (ESLint, then `scripts/check-structure.mjs`) | `npm run lint` |
| Production build | `npm run build` |
| Add a shadcn component (lands in `src/shared/ui`) | `npx shadcn@latest add <name>` |

There is no client test runner yet.

## Architecture

### Domain flow

A Program contains Projects. A Project is split into ordered **Phases** (week ranges) and has Demand: budget hours per expertise and minimum level, placed in one phase or across the whole project. A phase ending later shifts every later phase (`PhaseSlipped`) and triggers a re-plan; this is how "QA waits for Engineering" is modelled. Ownership is an `ASSIGNMENT` (person × demand, kind `primary` / `cross_expertise` / `stretch`) with weekly `ALLOCATION`s. A person can hold several expertises in the same week, against one shared capacity; the engine prefers primary-expertise people and never proposes stretch. The pure **Planning Engine**, `propose(snapshot, policy)`, turns people (skills, levels, tags, capacity) and demand (profile, required and preferred tags) into allocations. Capacity is weekly hours (effective-dated) minus non-project hours and absence, and 0 outside a person's start and end dates.
- A result is saved as a versioned `PLAN` of kind `draft`, shown in the UI as a "staffing draft". The word "proposal" is avoided because it means a business proposal.
- A Planner accepts the draft with `baseVersion` and it becomes the new `baseline`. A stale version gets `409`.
- A ChangeEvent (leave, slipped deadline) triggers a re-plan. Events are debounced, and there is at most one open auto re-plan draft per baseline: later events are merged into it, keeping the Planner's edits and locks.
- All metrics (utilization, gap, staffing %, Time/Scope/Planned burn %, EAC hours, capacity forecast) are computed on the server. The formulas are in `README.md` section 7. Thresholds and policy defaults are server config (`server/README.md` section 8).

### Server (Go, Echo v5, not Gin)

- Built from a template, layer-based:
  - `internal/handlers` → `internal/services` → `internal/repositories` → `internal/models`
  - `internal/dtos` only at the handler edge
  - a logical module is a set of same-named files across the layers
- `cmd/server/main.go` wires providers with **Uber FX**. `cmd/server/routes/router.go` holds the middleware chain and routes under `/api/v1`, protected by `AuthMiddleware` + `RequireRole`.
- Auth uses Keycloak JWTs. Roles `planner`, `lead` and `specialist` are in `internal/constants/roles.go`. CSRF is on: unsafe requests need `X-CSRF-Token`, which the server exposes in a response header.
  - `AuthMiddleware` verifies tokens locally against the realm's cached JWKS (`internal/integration/auth/verifier.go`), with no Keycloak call per request. It puts a `request.Principal` in the request context.
  - Routes add `RequireRole(constants.<X>Roles...)`. Services ask `AuthzService` for named checks (`Require(ctx, CanViewOwnProfile)`).
  - Keycloak imports `deploy/keycloak/realm.json` on a fresh start (placeholders from `.env`). The demo users `planner`, `lead` and `specialist` are for development only.
- "Now" comes from `clock.Clock` (FX, organization time zone from `TIMEZONE`). Nothing sets `time.Local`.
- Every response is the envelope `{ meta: { error_code, message, code, page, page_size, total }, data }`, built by `internal/errors` (`AppError`, `ErrorHandler`) and `BaseHandler`.
- `docs/` is generated Swagger (code-first, Swagger 2.0). Do not edit it by hand.
- Migrations are Atlas files in `cmd/migrations/sql`, generated from GORM models.
- `internal/engine` imports the stdlib only: `engine.Propose(request)` with the greedy strategy. The staffing service builds the snapshot (capacity, demand windows, the baseline) and saves the result as a draft.
- Planned, not built yet:
  - in-process background jobs (re-plan, import, notify)
- `cmd/server/.env` is read by the app, the Makefile and docker compose. `deploy/postgres/init` creates the `keycloak` database only when the volume is first created.

### Client (React + TS, Vite, Tailwind v4, shadcn/ui on Base UI)

- Layers:
  - `pages/` composes `features/`
  - features never import each other
  - shared code lives in `shared/` (api, i18n, ui, components, stores, and domain modules such as `programs` and `expertises`), `utils/` (shared constants and functions, including the dayjs ISO-week helpers) and `resources/` (endpoints, routes, RBAC)
- HTTP goes only through `shared/api/http.ts`:
  - `request<T>()` returns the unwrapped `data`; `requestWithMeta<T>()` returns `{ data, meta }`.
  - Every failure becomes an `ApiError`.
  - JWT and CSRF headers are added automatically.
- Server state lives in TanStack Query: invalidate after mutations, and `shared/api/query-client.ts` shows one shared toast on error (a `401` ends the session; a mutation lists codes it shows itself in `meta.quietErrorCodes`). Zustand holds UI and session state only.
- Sign-in: `oidc-client-ts` (PKCE) in `shared/auth/`. The session token is in `shared/stores/auth-store.ts`; the role comes from `/me` via `useCurrentUser`. `app/router.tsx` guards routes with `RequireAuth` and `RequirePermission` (permissions in `resources/rbac`).
- shadcn here uses **Base UI**: components take a `render` prop, not `asChild`. Files in `src/shared/ui` are project code; add variants there rather than overriding them from outside.

## Conventions (load the skill before touching that area)

Skills live in `.claude/skills/`. On the client, `npm run lint` enforces most of these mechanically.

Server:

- `structuring-server-modules`:
  - One file per layer per module, `ctx`-first repositories, services that take input structs.
  - DTO mapping lives in `dtos`.
  - Paths, roles, messages and error codes are constants in `internal/constants`.
  - Swagger and migrations are regenerated, not edited.
- `writing-go-server-code`: error flow (repo raw → service `AppError` → handler `HandleError`), GORM through `r.db.From(ctx)` with `TxManager.Do` for multi-step writes, validation via custom tags, UTC and ISO weeks, FX DI, no package state.
- `testing-go-server-code`: table-driven tests with testify mocks, handler tests through `httptest`, RBAC tests through the real middleware, `-tags=integration` for a real DB.

Client:

- `translating-client-text`:
  - The i18n key is the English text, declared in `LanguageKey` (`src/shared/i18n/language-key.ts`, sorted alphabetically) and passed to a typed `t('Program {id}', { id })`.
  - There is no English locale file.
- `structuring-client-components`:
  - Each component lives in a kebab-case folder with `index.ts` (re-export), `component.tsx`, `types.ts` and optional `constants.ts` / `utils.ts` / `mock-data.ts`.
  - Exports are named `<Folder>Page` in `pages/` and PascalCase of the folder elsewhere.
  - `component.tsx` holds only imports and one exported component: no nested components, no types, no inline function props.
  - Children go in `components/<child>/`, at most 3 layers deep.
- `styling-client-components`: no `style={...}`. Use Tailwind theme tokens, `cn()` and shadcn variants.
- `organizing-client-shared-code`:
  - Shared constants and functions go in `src/utils`.
  - API paths, routes and RBAC go in `src/resources`.
  - Dates only through `@/utils/functions` (direct `dayjs` imports fail lint).
