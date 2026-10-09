---
name: structuring-server-modules
description: Use when adding or changing an API endpoint, module, model, repository, service or handler in the Slot Finder Go server, when wiring a provider into FX or a route into the router, or when a route path, role, response message, error code or config value is about to be written as a literal.
---

# Structuring server modules

## Overview

The server is Echo v5 + Uber FX + GORM, layered `handlers → services → repositories → models`. A module (people, programs, ...) is **one file per layer with the same name**. Every literal that means something (path, role, message, error code, setting) has one named definition. **Never hardcode.**

## Files of a module

| File | Holds | Rules |
|---|---|---|
| `internal/models/person.go` | GORM model embedding `BaseModel` | DB tags only, no `json` tags |
| `internal/dtos/person.go` | Request DTOs (`json`, `query`, `validate` tags, `Normalize()`), response DTOs, and `NewPersonResponse(*models.Person)` mapping | The only place a model becomes JSON |
| `internal/repositories/person.go` | `PersonRepository` interface embedding `AbstractRepository[models.Person]`, impl embedding `abstractRepository`, `ProvidePersonRepository` | `ctx` first on every method, `r.db.WithContext(ctx)` |
| `internal/services/person.go` | `PersonService` interface, unexported impl, `ProvidePersonService`, input structs (`CreatePersonInput`) | Takes `context.Context` and input structs, never DTOs or `echo.Context`. Returns models, plus an `error` that holds an `*errors.AppError` (declare the return type as `error` to avoid the typed-nil trap) |
| `internal/handlers/person.go` | `PersonHandler` embedding `BaseHandler`, `ProvidePersonHandler`, one method per route with swag annotations | Bind → `Normalize()` → validate → map DTO to input → call service → respond |
| `*_test.go` next to each | See testing-go-server-code | |

Then wire it in:
1. `cmd/server/main.go`: add `Provide*` for the repository, service and handler to `fx.Provide`.
2. `cmd/server/routes/router.go`: add a group under `v1` with `AuthMiddleware` on the group and `RequireRole` on each route.
3. Regenerate Swagger: `make swagger-load`. Without the `swag` binary, run the same command as `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init` with the Makefile's flags.
4. Create the migration: `make migrate-generate name=<change>`. This needs the Atlas CLI (`brew install ariga/tap/atlas`) and Docker. If either is missing, say so in your report. Never hand-write migration SQL or edit `atlas.sum`.

## Named definitions (never hardcode)

| Kind | Lives in |
|---|---|
| Route path | `internal/constants/routes.go`, e.g. `PathPeople = "/people"` |
| Role list for a route | `internal/constants/roles.go`, e.g. `PeopleViewRoles = []string{RolePlanner, RoleLead}` |
| Response message | `internal/constants/messages.go`, e.g. `MsgPeopleListed` |
| Error code | `internal/constants/error_codes.go`, e.g. `PersonEmailAlreadyExists = "PERSON_EMAIL_ALREADY_EXISTS"` |
| Allowed values in validation | A custom tag in `internal/utils/validation.go` (`validate:"role"`), not `oneof=planner lead` |
| Paging limits | `constants.DefaultPage`, `DefaultPageSize`, `MaxPageSize`, applied by `PageableRequest.Normalize()` |
| Operation and resource names for errors | Unexported constants at the top of the service file: `opCreatePerson`, `resourcePerson` |
| Setting, URL, timeout, secret | A field in `internal/config/config.go`, read with `getEnv`, added to `cmd/server/.env.example`. Never `os.Getenv` elsewhere |

Struct tags cannot reference constants, so numeric bounds in `validate` tags (`gte=1,lte=60`) and swag `enums:` are the only allowed literals. Keep every allowed-value list in a custom tag.

## Lists and paging

- A list request DTO embeds `dtos.PageableRequest` (bound from `?page=&page_size=`) and calls its `Normalize()` from its own `Normalize()`.
- The service builds a `*dtos.PageableRequest` from its input struct. The repository returns `*dtos.DataResponse[T]`.
- The handler answers with `h.SuccessResponse(c, msg, dtos.NewXResponses(page.Data), page.Pageable)`. `NewXResponses` never returns nil, so empty lists are sent as `[]`.

## Responses and errors

- Success: `h.SuccessResponse(c, constants.MsgX, data, pageable)`. Use `h.CreatedResponse` for 201.
- Failure: return `h.HandleError(c, err)` with an `*errors.AppError` from the service. Constructors: `NotFoundError`, `ConflictError`, `DatabaseError`, `ValidationError`, `ForbiddenError`, ... Use `WrapError(cause, code, msg, type, status)` for a module-specific code.
- Binding and validation failures: `h.ValidationErrorResponse(c, msg, errors.ParseValidationErrors(err))`. Field names come out in JSON form.
- Every response is `{meta, data}`, including 401 and 403 from the auth middleware. Swag annotations describe it as `object{meta=dtos.Meta,data=dtos.PersonResponse}`.

## Planning engine

`internal/engine` imports **only the standard library**: no GORM, config, logger or `time.Now()`. The planning service builds a snapshot and passes it in.

## Common mistakes

| Mistake | Fix |
|---|---|
| A `toPersonResponse` helper in the handler | `dtos.NewPersonResponse` |
| A service method taking `*dtos.CreatePersonRequest` | Take `CreatePersonInput`; the handler maps the DTO to it |
| `Create(entity)` without `ctx` | The shared repository methods all take `ctx` first |
| Converting duplicate-key errors inside a repository | GORM runs with `TranslateError`; check `errors.Is(err, gorm.ErrDuplicatedKey)` in the service |
| `"People retrieved"` or `"/people"` written inline | Constants in `messages.go` and `routes.go` |
| Editing `docs/` by hand | Change the annotations and run `make swagger-load` |
