/** Metadata block sent with every server response. */
export interface ApiMeta {
  error_code: string
  message: string
  code?: number
  page?: number
  page_size?: number
  total?: number
}

/** Envelope that wraps every server response. */
export interface ApiEnvelope<T> {
  meta: ApiMeta
  data: T
}

/** Unwrapped payload together with its meta, e.g. for paginated lists. */
export interface ApiResult<T> {
  data: T
  meta: ApiMeta
}

interface ApiErrorInit {
  status: number
  code: string
  message: string
  details?: unknown
}

/** The single error type for every failed request, HTTP or network. */
export class ApiError extends Error {
  /** HTTP status, or 0 when no response was received. */
  readonly status: number
  /** Server `error_code`, or a client-side code such as `NETWORK_ERROR`. */
  readonly code: string
  readonly details?: unknown

  constructor({ status, code, message, details }: ApiErrorInit) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}
