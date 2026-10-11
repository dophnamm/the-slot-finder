/** Why the automatic sign-in stopped: Keycloak is unreachable, or the server keeps rejecting the session. */
export type SignInFailure = 'unreachable' | 'repeated'
