---
name: organizing-client-shared-code
description: Use when the Slot Finder client needs a constant, helper function, API endpoint path, route path, role or permission check, or any date/week calculation or formatting, and whenever a literal URL, path, role name, number or format string is about to be written inline.
---

# Organizing client shared code

## Overview

**Never hardcode.** Every URL, route, role, threshold, timeout, format string and magic number has one named definition. Where it lives depends on who uses it.

## Where it goes

| What | Used by one component | Used in several places |
|---|---|---|
| Constant | `<component>/constants.ts` | `src/utils/constants/<topic>.ts` |
| Helper function | `<component>/utils.ts` | `src/utils/functions/<topic>.ts` |
| Server API path | — | `src/resources/endpoints/<domain>.ts` |
| Client route path | — | `src/resources/routes/index.ts` (`ROUTES`) |
| Role or permission | — | `src/resources/rbac/` (`ROLES`, `PERMISSIONS`, `can()`) |
| Module-level setting of a shared module (e.g. HTTP headers) | — | `constants.ts` next to that module, e.g. `src/shared/api/constants.ts` |
| API calls, query hooks and types of a domain | `features/<feature>/api.ts`, `use-<name>.ts`, `types.ts` | `src/shared/<domain>/` with an `index.ts`, e.g. `shared/programs`, `shared/expertises` |

Each `src/utils/*` and `src/resources/*` folder has an `index.ts` that re-exports it. Import from the folder: `@/utils/constants`, `@/utils/functions`, `@/resources/endpoints`, `@/resources/routes`, `@/resources/rbac`.

## Core pattern

```ts
// ❌ hardcoded
request({ url: `/programs/${id}/overview` })
<Link to="/" />
if (user.role === 'planner') ...
if (load > 100) ...

// ✅ named once, used everywhere
request({ url: PROGRAM_ENDPOINTS.overview(id) })
<Link to={ROUTES.dashboard} />
if (can(user.role, 'people:manage')) ...
if (load > FULL_LOAD_PERCENT) ...   // FULL_LOAD_PERCENT in constants.ts
```

## Endpoints, routes and RBAC

- **Endpoints** are relative to the API base (`/api/v1`). Use a builder function for ids, with `encodeURIComponent`: `overview: (id: string) => \`${program(id)}/overview\``.
- **Routes:** `ROUTES` holds every path for the router, nav and links. For params, link with `generatePath(ROUTES.programDetail, { id })`.
- **RBAC:** `ROLES` must match the server's `internal/constants/roles.go`. Add a permission to `PERMISSIONS` and check it with `can(role, permission)`. This only shows or hides UI; the server enforces the real rule.

## Dates and ISO weeks: always dayjs

Use the helpers in `@/utils/functions`. They load dayjs with the `isoWeek` and `utc` plugins.

| Need | Use |
|---|---|
| Week label of a date | `toIsoWeek(date)` → `'2026-W42'` |
| Monday–Sunday of a week | `isoWeekRange('2026-W42')` → `{ start, end }` (Dayjs) |
| Range text | `formatIsoWeekRange('2026-W42')` → `'Oct 12 – Oct 18'` |
| Anything else | `dayjs` from `@/utils/functions`, formats from `DATE_FORMAT` in `@/utils/constants` |

Date and week fields are `DatePicker` and `WeekPicker` from `@/shared/components` (values `YYYY-MM-DD` and `2026-W42`), never `<Input type="date|week">`: the browser's own pickers do not match the app, and lint rejects them. Convert to and from a calendar's `Date` with `parseDateValue` / `toDateValue`.

Never use `new Date()` arithmetic, `toLocaleDateString` or hand-written week math. A new date helper goes in `src/utils/functions/date.ts`, and its format strings go in `src/utils/constants/date.ts`.

## Common mistakes

| Mistake | Fix |
|---|---|
| A shared constant copied into two `constants.ts` files | Move it to `src/utils/constants/<topic>.ts` |
| `'/api/v1/programs'` written in a feature | `PROGRAM_ENDPOINTS.list` (the base path is added by the HTTP client) |
| `role === 'lead'` in a component | `can(role, '<permission>')`, adding the permission to `PERMISSIONS` if needed |
| `date.getDay()` math to find a week | `toIsoWeek` / `isoWeekRange` |
| Importing `dayjs` directly from `'dayjs'` | Import from `@/utils/functions` so the plugins are always loaded |
