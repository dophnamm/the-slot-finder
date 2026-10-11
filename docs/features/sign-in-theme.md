# Sign-in theme

**Goal:** signing in looks and feels like part of The Slot Finder, with no "Sign in with Keycloak" step and no Keycloak branding, while Keycloak still handles credentials, sessions and lockouts.

**Roles:** everyone who signs in (planner, lead, specialist).
**Target:** before the 17 Oct freeze · **Depends on:** [sign-in](sign-in.md) · **Branch:** `feat/sign-in-theme`

## Approach

A **Keycloak login theme** named `slot-finder`: FreeMarker templates that extend Keycloak's `base` theme, one CSS file with the app's tokens, the app's font and logo, and English message overrides. The browser still redirects to Keycloak with Authorization Code + PKCE, but the page it lands on looks like the app.

| Option | Why not |
|---|---|
| Our own login form posting the password to the server (password grant / BFF) | The app would handle passwords; the grant is dropped in OAuth 2.1; MFA, password reset and social login would have to be rebuilt |
| Keycloak's page in an iframe | Keycloak forbids framing by default; cookie and clickjacking problems |
| Keycloakify (a React-built theme packaged as a JAR) | More tooling (separate build, JAR in the image) than this needs; a later option if the login pages need richer UI |

## In scope

- Theme files in `server/deploy/keycloak/themes/slot-finder/login/`:
  - `theme.properties` (`parent=base`), `template.ftl` (page layout), `login.ftl` (sign-in form)
  - styled states: wrong username or password, account temporarily locked (brute-force protection is on), expired login page, generic error and info pages
  - `resources/css/login.css`: colours, radius and spacing copied from `client/src/index.css` as CSS custom properties in one place, light and dark (`prefers-color-scheme`)
  - `resources/fonts/`: the app's font (Geist, OFL), served by Keycloak
  - `resources/img/logo.svg` and `messages/messages_en.properties` ("Sign in to The Slot Finder", field labels, error texts)
- Keycloak setup:
  - docker compose mounts the themes folder into `/opt/keycloak/themes` (read-only)
  - `realm.json` sets `loginTheme` from `KEYCLOAK_LOGIN_THEME` (`slot-finder` in `.env.example`), so a fresh realm import uses it
  - `make keycloak-theme` applies the theme to an existing realm through `kcadm.sh`, without resetting data
  - `start-dev` already disables theme caching, so template edits show on reload
- Client:
  - A signed-out user who opens any page goes straight to the themed sign-in page (`RequireAuth` calls `signIn(returnTo)`); the intermediate page with the "Sign in with Keycloak" button goes away
  - After sign-out, Keycloak returns to a new **Signed out** page in the app with a "Sign in again" button, so the app does not sign straight back in
  - If Keycloak cannot be reached, the app shows a short error with "Try again" instead of a blank page

## Not now

- Registration, social login and MFA setup pages: they keep the `base` look until they are used → [Later](README.md#later).
- Email and account-console themes → [Later](README.md#later).
- Serving Keycloak under the app's domain (e.g. `/auth` behind the same host), so the address bar does not change → [Later](README.md#later), a deployment task.

## Server

- No API changes. The server keeps verifying tokens locally against the realm's JWKS.
- Settings: `KEYCLOAK_LOGIN_THEME` in `cmd/server/.env.example`; the theme name is never written elsewhere.

## Client

- `RequireAuth`: signed out → `signIn(returnTo)`; keeps `returnTo` so the user lands where they started.
- Routes: `/signed-out` (new, public) replaces the current `/login` button page; `post_logout_redirect_uri` points to it.
- Texts through `t()`; the "Sign in with Keycloak" key is removed.

## Acceptance criteria

1. Opening the app while signed out lands directly on the sign-in page; there is no extra click.
2. The sign-in page shows The Slot Finder's name and logo, the app's font and colours, and no Keycloak name or logo, in light and dark mode.
3. A wrong password shows an inline error in the same style; after the brute-force limit, the lockout message is styled too.
4. After signing in, the user is on the page they first asked for.
5. Signing out lands on the app's **Signed out** page; it stays there until "Sign in again" is clicked.
6. `make reset-data && make up` uses the theme with no manual step; on existing data, `make keycloak-theme` switches it on.
7. Keycloak's own sign-in flow still runs: PKCE, brute-force protection and session handling are unchanged.

## Tasks

**Keycloak**
- [ ] Theme templates, CSS with the app's tokens, font, logo, messages; error, expired and info pages
- [ ] Compose mount, `KEYCLOAK_LOGIN_THEME`, `loginTheme` in `realm.json`, `make keycloak-theme`

**Client**
- [ ] `RequireAuth` redirects straight to sign-in; `/signed-out` page; post-logout redirect; unreachable-Keycloak error
- [ ] Remove the button page and its texts

**Tests**
- [ ] Component tests for the Signed out page and the redirect in `RequireAuth`
- [ ] Browser check: sign in, wrong password, lockout, sign out, light and dark

**Estimate:** about one day.
