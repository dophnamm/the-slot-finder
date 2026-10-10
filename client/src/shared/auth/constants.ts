/** Scopes asked for at sign-in: the user's id, name and email. */
export const OIDC_SCOPE = 'openid profile email'

/** Query key of the signed-in user (`GET /me`). */
export const CURRENT_USER_QUERY_KEY = ['current-user'] as const
