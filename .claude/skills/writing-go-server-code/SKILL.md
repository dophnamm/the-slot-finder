---
name: writing-go-server-code
description: Use when writing or reviewing Go code in the Slot Finder server, including error handling, context propagation, GORM queries and transactions, validation, logging, time and ISO weeks, interfaces, goroutines and naming.
---

# Writing Go server code

## Overview

Idiomatic Go (Effective Go, Go Code Review Comments) applied to this template. When in doubt, match the nearest existing code. **REQUIRED:** use structuring-server-modules for where code goes.

## Errors

- Check every error. Never assign one to `_`.
- **Repositories** return raw errors (`gorm.ErrRecordNotFound`, `gorm.ErrDuplicatedKey`, driver errors).
- **Services** turn them into `*errors.AppError` with `WithOperation` and `WithResource`, returned as `error`. Compare with `errors.Is` / `errors.As` (alias the stdlib package as `stderrors`).
- **Handlers** pass errors to `h.HandleError`. It logs and reports them once, so do not log and also return the same error.
- Wrap low-level errors with context: `fmt.Errorf("load snapshot: %w", err)`.
- No `panic` in request paths. Return early instead of nesting.

```go
if err := s.repo.Create(ctx, person); err != nil {
	if stderrors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, errors.WrapError(err, constants.PersonEmailAlreadyExists, constants.MsgPersonEmailTaken,
			errors.ErrorTypeConflict, http.StatusConflict).WithOperation(opCreatePerson).WithResource(resourcePerson)
	}
	return nil, errors.DatabaseError(constants.MsgPersonCreateFailed, err).
		WithOperation(opCreatePerson).WithResource(resourcePerson)
}
```

## Context

- `ctx context.Context` is the **first** parameter of every service and repository method that does I/O. Handlers pass `c.Request().Context()`.
- Never store a context in a struct. Never pass `echo.Context` below the handler.
- Outbound calls get a timeout from config.

## Database (GORM)

- Always `r.db.WithContext(ctx)`.
- Use parameters, never string concatenation: `Where("email = ?", email)`. Escape `%`, `_` and `\` in `ILIKE` search terms.
- Unique rules belong in the model (`uniqueIndex`) and are enforced by the DB. A pre-check is only a friendlier message; still handle `gorm.ErrDuplicatedKey`.
- A multi-step write is one transaction, owned by the service: `db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { ... })`, with repository methods that accept `tx`.
- List order is always deterministic: `Order("name ASC").Order("id ASC")`.

## Validation and input

- Validate at the handler with struct tags. Call `Normalize()` (trim and lowercase where meaningful) **before** validating.
- Custom rules are tags registered in `utils.RegisterValidations`. Error messages use JSON field names.
- Business rules (uniqueness, state transitions, capacity) live in the service, not in tags.

## Time

- Everything is UTC (GORM `NowFunc` is UTC).
- ISO weeks: `year, week := t.ISOWeek()`, formatted as `fmt.Sprintf("%d-W%02d", year, week)`. Put the format in a constant.
- Pure code (engine, metrics) receives "now" as a parameter.

## Design

- Each service and repository has an interface and an unexported impl, and its `Provide*` returns the interface. This makes them testable with mocks.
- Accept interfaces, return concrete types, except where FX provides the interface.
- No package-level mutable state. Dependencies come through FX constructors. Read-only lookup tables (`constants.Roles`, role groups) are fine, but are never modified.
- Every goroutine has an exit path tied to a `ctx` or an FX lifecycle hook. Use `errgroup` for parallel work.
- Logging: `logger.Log` (zap) with typed fields. Never `fmt.Println`. Never log tokens, passwords or full emails.

## Naming

- MixedCaps. Acronyms in caps: `ID`, `URL`, `HTTPStatus`.
- Short receivers (`s`, `h`, `r`). File names are lowercase, without underscores (except `_test.go`).
- Package names are lowercase single words without stutter: `services.PersonService`, not `services.PersonServiceService`.

## Before you finish

`gofmt -l .` is empty, then `go build ./...`, `go vet ./...`, `go test ./...`, `make lint`. Run `make swagger-load` if any handler annotation changed.
