# Sign-in

**Goal:** every user signs in through Keycloak and reaches exactly the screens and API their role allows. Every later feature, and so every goal (G1–G3 in the [root README](../../README.md#1-problem-and-scope)), builds on it.

**Roles:** planner, lead, specialist.
**Target:** 10–11 Oct · **Depends on:** none · **Branch:** `feat/sign-in`

## In scope

- Keycloak realm import: the realm, realm roles `planner` / `lead` / `specialist`, a public SPA client (Authorization Code + PKCE S256), the confidential server client, an audience mapper, and one demo user per role (development only).
- Client: OIDC sign-in with PKCE, silent token renewal, sign-out, route guards by role.
- Server: local JWT verification against cached JWKS, claims in the request context, `RequireRole`, `GET /me`.
- An `authz` service with named checks. It holds role checks now; per-program rules arrive with later features.
- Groundwork: body limit and log truncation, rate-limit envelope and config, service name from config, `/health/metrics` behind basic auth, a Clock that knows the organization time zone, CI on pull requests.

## Not now

- The PERSON link (`personId`) in `/me` → [people-and-skills](people-and-skills.md).
- Per-program rules (lead of this program, internal assistant) → [programs-and-demand](programs-and-demand.md) and [Later](README.md#later).
- Token introspection and UMA permissions (template leftovers, no longer used per request).
- Sentry. There is no DSN yet, and header and body redaction must be fixed before `SENTRY_DSN` is set.
- Managing users inside the app. Users are managed in Keycloak.

## Server

**Endpoints**

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/me` | The signed-in user from the token: `id` (subject), `username`, `name`, `email`, `role` | planner, lead, specialist |

**Tables:** none.

**Rules**
- `AuthMiddleware`:
  - verifies the Bearer token locally: RS256 signature against the realm's JWKS, `iss`, `exp`, and either `aud` containing the server client or `azp` equal to the SPA client
  - caches the JWKS and refreshes it on a schedule (`AUTH_JWKS_REFRESH_INTERVAL`), and right away (rate-limited) when a token carries an unknown `kid`
  - allows clock skew of `AUTH_CLOCK_SKEW`
  - makes no Keycloak call per request
- The issuer is the public Keycloak URL that the browser uses (`KEYCLOAK_PUBLIC_URL`). It may differ from the URL the server uses to fetch the JWKS (`KEYCLOAK_URL`), e.g. inside Docker.
- Claims go into the request `context.Context`. Handlers and services read them through a nil-safe accessor.
- A user's **role** is the highest of their realm roles, in the order planner, lead, specialist. A token without any of the three gets `403` on every protected route.
- Every 401 and 403 uses the `{ meta, data }` envelope.
- `authz` exposes named checks (e.g. `CanViewPeople`). Handlers and services call the checks, never raw role lists.

## Client

- **Library:** `oidc-client-ts` (standard OIDC, maintained, typed). Recorded in [`client/README.md`](../../client/README.md).
- **Routes:**
  - `/login`: a sign-in page with one button that redirects to Keycloak
  - `/auth/callback`: finishes the code exchange and returns to the page the user asked for
  - every other route needs a signed-in user
  - routes declare the roles that may open them; others see a "no access" page
- **Behaviour:**
  - the token renews silently before it expires
  - sign-out ends the Keycloak session
  - the navigation hides links the role cannot open
- **Code:** `shared/auth/` (OIDC manager, auth store, guards), and `shared/api/http.ts` reads the token from the auth store.

## Acceptance criteria

1. `make reset-data && make up` imports the realm. The three demo users exist, each with one role.
2. Opening any page while signed out shows `/login`. After signing in, the user lands on the page they first asked for.
3. API calls carry `Authorization: Bearer`. A tab left open longer than the access-token lifespan keeps working, because the token renews silently.
4. `GET /api/v1/me` returns the id, username, name, email and role of each demo user.
5. A missing, malformed, expired, wrong-issuer or wrong-audience token gets `401`. A valid token without an allowed role gets `403`. Both use the envelope.
6. The server fetches the JWKS once and on refresh only, never per request.
7. A specialist does not see planner-only links, and opening such a route directly shows the "no access" page.
8. Sign-out ends the Keycloak session. Going back in the browser does not reopen the app without signing in again.
9. Too many requests gets `429` in the envelope. A body over the limit gets `413`. `/health/metrics` without basic auth gets `401`.
10. CI runs `make check` on the pull request and passes.

## Tasks

**Server**
- [ ] `deploy/keycloak/realm.json` with placeholders filled from the environment; compose mounts it and runs `start-dev --import-realm`
- [ ] JWT verifier with cached JWKS (issuer, audience or azp, expiry, skew); `AuthMiddleware` uses it; no introspection per request
- [ ] Claims in the request context with nil-safe accessors; `RequireRole` reads them
- [ ] `authz` service with role checks; `GET /me` handler, DTO and route constants
- [ ] Groundwork: body limit before body logging, multipart skipped, payload truncated; rate limiter envelope and config, IP extractor; `cfg.AppName` instead of the hardcoded service name; basic auth on `/health/metrics`; Clock via FX instead of `time.Local`
- [ ] `.env.example` entries for every new setting; Swagger regenerated

**Client**
- [ ] `oidc-client-ts` setup from env (`VITE_OIDC_AUTHORITY`, `VITE_OIDC_CLIENT_ID`), auth store, `/login` and `/auth/callback` pages
- [ ] Silent renewal and sign-out
- [ ] Route guards by role, a "no access" page, navigation filtered by role
- [ ] `http.ts` takes the token from the auth store; `/me` hook

**Tests**
- [ ] Verifier: valid, expired, wrong issuer, wrong audience, unknown `kid` refresh, bad signature
- [ ] RBAC through the real middleware chain: each role gets 200 or 403, no token gets 401
- [ ] `/me` handler and nil-safe claims
- [ ] Rate-limit, body-limit and metrics basic-auth responses
- [ ] `.github/workflows/check.yml` runs `make check` on pull requests
