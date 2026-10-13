import type { ApiError } from '@/shared/api/types'
import { CALENDAR_FEED_ERROR_CODES } from '../../constants'
import type { CalendarFeedErrorCode } from '../../types'

/** The link problem to show under the field, or null for none (other errors use the toast). */
export function linkProblem(error: ApiError | null): CalendarFeedErrorCode | null {
  return CALENDAR_FEED_ERROR_CODES.find((code) => code === error?.code) ?? null
}
