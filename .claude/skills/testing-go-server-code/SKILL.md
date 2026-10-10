---
name: testing-go-server-code
description: Use when adding or changing behavior in the Slot Finder Go server and its tests need writing or updating, when a test needs a database, Keycloak or other external service, or when choosing between unit, handler and integration tests.
---

# Testing Go server code

## Overview

Tests sit next to the code (`person.go` → `person_test.go`, same package). They are table-driven, use testify, and never touch a real database or network unless tagged `integration`.

## By layer

| Layer | How |
|---|---|
| Service | Mock the repository interface with `testify/mock`. Cover success, each error branch, and every business rule (uniqueness, defaults, transitions) |
| Handler | `httptest` + `echo.New()` + a mocked service. Assert the status code and the `{meta, data}` body: `error_code` and the JSON field names in validation errors |
| Router / RBAC | Send requests through the real middleware chain (`routes.Router`) with a fake `auth.TokenVerifier` that maps test tokens to claims. Each role gets the right 200 / 403, and a missing token gets 401. CSRF runs before auth, so unsafe methods (POST, PUT, PATCH, DELETE) must send `Sec-Fetch-Site: same-origin` or a token, or they get 400 before reaching the role check |
| Repository | Query-shape tests with a GORM dry-run session (`DryRun: true, SkipDefaultTransaction: true`), or a real DB in an integration test. In dry-run the count is always 0, so the shared list query returns early. Test the filter/order builder on its own |
| FX wiring | Keep the provider list in a function (e.g. `appOptions()`), and let a test call `fx.ValidateApp(appOptions()...)` so a missing provider fails in CI, not at startup |
| Pure code (engine, metrics, utils) | Plain table tests. Engine invariants also get property tests: never over capacity, never below `min_level`, never outside the week range, never changes a locked allocation |
| Real DB | `//go:build integration` + testcontainers, run with `go test -tags=integration ./...` |

## Shape

```go
func TestPersonService_Create(t *testing.T) {
	tests := []struct {
		name     string
		input    CreatePersonInput
		setup    func(repo *MockPersonRepository)
		wantCode string // "" means success
	}{
		{name: "creates person", input: validInput(), setup: expectCreate(nil)},
		{name: "duplicate email", input: validInput(), setup: expectCreate(gorm.ErrDuplicatedKey),
			wantCode: constants.PersonEmailAlreadyExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockPersonRepository)
			tt.setup(repo)
			_, err := ProvidePersonService(repo).Create(context.Background(), tt.input)
			if tt.wantCode == "" {
				require.NoError(t, err)
			} else {
				var appErr *errors.AppError
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, tt.wantCode, appErr.Code)
			}
			repo.AssertExpectations(t)
		})
	}
}
```

- Name tests `Test<Type>_<Method>` and give cases readable `name`s.
- Use `require` for preconditions and `assert` for the remaining checks.
- Compare against constants (`constants.PersonEmailAlreadyExists`), never repeated literals.
- No `time.Sleep`. Sync on channels or use `require.Eventually`.
- `t.Parallel()` only when the test shares no mutable state.

## Commands

| Run | Command |
|---|---|
| All unit tests | `go test ./...` |
| One package / one test | `go test ./internal/services -run TestPersonService_Create -v` |
| CI-like (race, coverage) | `make tests` |
| Integration (Docker) | `go test -tags=integration ./...` |

A bug fix starts with a test that fails for the bug. A refactor keeps the tests green without editing their expectations, unless the expected behavior itself is changing.
