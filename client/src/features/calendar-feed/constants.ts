import type { PlainKey } from '@/shared/i18n'
import type { CalendarFeedErrorCode } from './types'

export const calendarFeedQueryKey = (personId: string) => ['calendar-feed', personId] as const

/** Answers that are states, not errors: no calendar yet, or sync turned off on the server. */
export const CALENDAR_FEED_NOT_FOUND = 'CALENDAR_FEED_NOT_FOUND'
export const CALENDAR_SYNC_DISABLED = 'CALENDAR_SYNC_DISABLED'

/** What to tell the person for each problem with their link. */
export const CALENDAR_FEED_ERROR_LABELS: Record<CalendarFeedErrorCode, PlainKey> = {
  CALENDAR_FEED_HOST_NOT_ALLOWED:
    'Use the secret iCal address from Google Calendar or Outlook. It starts with https://.',
  CALENDAR_FEED_UNREACHABLE: 'The calendar could not be fetched. Your organization may block secret calendar links.',
  CALENDAR_FEED_NOT_ICS: 'The link does not return a calendar. Your organization may block secret calendar links.',
  CALENDAR_FEED_TOO_LARGE: 'The calendar is too large to read.',
}

/** The codes the connect form shows itself, so the shared error toast skips them. */
export const CALENDAR_FEED_ERROR_CODES = [
  'CALENDAR_FEED_HOST_NOT_ALLOWED',
  'CALENDAR_FEED_UNREACHABLE',
  'CALENDAR_FEED_NOT_ICS',
  'CALENDAR_FEED_TOO_LARGE',
] as const satisfies readonly CalendarFeedErrorCode[]
