import type { ApiError } from '@/shared/api/types'

export interface LoadErrorProps {
  /** Why the request failed; its message says so under the title. After a 429 the button waits for `retry_after`. */
  error: ApiError
  onRetry: () => void
  /** Draws a bordered frame, for a message that stands in for a section. */
  framed?: boolean
  /** Renders as a table's only row spanning this many columns. */
  columns?: number
}
