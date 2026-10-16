/** Base path of the server API when VITE_API_BASE_URL is not set. */
export const DEFAULT_API_BASE_URL = '/api/v1'

/** Header that carries the Keycloak access token. */
export const AUTH_HEADER = 'Authorization'

/** Scheme prefix of the access token in AUTH_HEADER. */
export const AUTH_SCHEME = 'Bearer'

/** Axios leaves `method` undefined for GET. */
export const DEFAULT_METHOD = 'GET'

/** Header the server uses to expose, and expects back, the CSRF token. */
export const CSRF_HEADER = 'X-CSRF-Token'

/** Methods that do not change state, so they need no CSRF token. */
export const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS'])

/** Statuses a proxy returns when the API behind it is down. */
export const GATEWAY_STATUSES = new Set([502, 503, 504])

/** Client-side error codes for failures without a server response. */
export const CLIENT_ERROR_CODE = {
  network: 'NETWORK_ERROR',
  unknown: 'UNKNOWN_ERROR',
} as const

/** The server answers 401 when it no longer accepts the access token. */
export const UNAUTHORIZED_STATUS = 401

/** The server answers 404 when there is no such record. */
export const NOT_FOUND_STATUS = 404

/** The server answers 403 when the user may not see or do this; asking again will not change it. */
export const FORBIDDEN_STATUS = 403

/** The server answers 429 after too many requests; asking again at once will not help. */
export const TOO_MANY_REQUESTS_STATUS = 429

/** The server's `error_code` for 403. */
export const FORBIDDEN_CODE = 'FORBIDDEN'

/** Statuses where asking again at once cannot succeed: the session ended, access is denied, or the server asks to wait. */
export const NO_RETRY_STATUSES = new Set([UNAUTHORIZED_STATUS, FORBIDDEN_STATUS, TOO_MANY_REQUESTS_STATUS])
