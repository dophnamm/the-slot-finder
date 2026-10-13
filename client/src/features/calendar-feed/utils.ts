import { ApiError } from '@/shared/api/types'
import { CALENDAR_FEED_NOT_FOUND, CALENDAR_SYNC_DISABLED } from './constants'
import type { CalendarFeedState } from './types'

/** "No calendar yet" and "sync is off" are answers, not errors. Anything else returns null. */
export function stateFromError(error: unknown): CalendarFeedState | null {
  if (!(error instanceof ApiError)) return null
  if (error.code === CALENDAR_FEED_NOT_FOUND) return { kind: 'none' }
  if (error.code === CALENDAR_SYNC_DISABLED) return { kind: 'disabled' }
  return null
}
