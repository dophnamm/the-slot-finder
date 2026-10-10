interface ImportMetaEnv {
  /** Base URL of the server API. Defaults to `/api/v1` (proxied by Vite in dev). */
  readonly VITE_API_BASE_URL?: string
  /** Keycloak realm URL that issues tokens, e.g. `http://localhost:8080/auth/realms/slot-finder`. */
  readonly VITE_OIDC_AUTHORITY: string
  /** Public OIDC client of the web app in that realm. */
  readonly VITE_OIDC_CLIENT_ID: string
}
