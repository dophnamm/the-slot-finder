import { UserManager, WebStorageStateStore } from 'oidc-client-ts'
import { ROUTES } from '@/resources/routes'
import { OIDC_SCOPE } from './constants'

/**
 * OIDC client for Keycloak: Authorization Code with PKCE (S256), silent renewal with the
 * refresh token, and the session kept in sessionStorage (one tab, cleared when it closes).
 * After sign-out Keycloak returns to the dashboard, which sends the signed-out user straight on
 * to the sign-in page.
 */
export const userManager = new UserManager({
  authority: import.meta.env.VITE_OIDC_AUTHORITY,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: new URL(ROUTES.authCallback, window.location.origin).toString(),
  post_logout_redirect_uri: new URL(ROUTES.dashboard, window.location.origin).toString(),
  scope: OIDC_SCOPE,
  automaticSilentRenew: true,
  userStore: new WebStorageStateStore({ store: window.sessionStorage }),
})
