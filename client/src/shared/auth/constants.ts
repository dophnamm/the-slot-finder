/** Scopes asked for at sign-in: the user's id, name and email. */
export const OIDC_SCOPE = 'openid profile email'

/** Query key of the signed-in user (`GET /me`). */
export const CURRENT_USER_QUERY_KEY = ['current-user'] as const

/** sessionStorage key of this tab's recent automatic sign-ins; it survives the Keycloak round trip. */
export const AUTO_SIGN_IN_STORAGE_KEY = 'slot-finder:auto-sign-ins'

/**
 * Automatic sign-ins allowed within AUTO_SIGN_IN_WINDOW_MS before the app stops and asks. A server
 * that keeps rejecting fresh tokens (e.g. a wrong issuer setting) would otherwise bounce the
 * browser between the app and Keycloak forever.
 */
export const AUTO_SIGN_IN_LIMIT = 3
export const AUTO_SIGN_IN_WINDOW_MS = 60_000
