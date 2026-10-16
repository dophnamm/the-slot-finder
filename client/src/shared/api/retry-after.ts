import { TOO_MANY_REQUESTS_STATUS } from './constants'
import type { ApiError } from './types'

/** Seconds to wait after a 429 answer, when the server sent them in `data.retry_after`. */
export function retryAfterSeconds(error: ApiError): number | undefined {
  if (error.status !== TOO_MANY_REQUESTS_STATUS) return undefined
  const { details } = error
  if (typeof details !== 'object' || details === null || !('retry_after' in details)) return undefined
  const seconds = details.retry_after
  return typeof seconds === 'number' && seconds > 0 ? seconds : undefined
}
