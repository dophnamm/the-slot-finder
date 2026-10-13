/** Problems the server stores or answers with; they match its constants/error_codes.go. */
export type CalendarFeedErrorCode =
  'CALENDAR_FEED_HOST_NOT_ALLOWED' | 'CALENDAR_FEED_UNREACHABLE' | 'CALENDAR_FEED_NOT_ICS' | 'CALENDAR_FEED_TOO_LARGE'

/** A connected calendar as the server returns it. The link itself never comes back. */
export interface ConnectedCalendar {
  host: string
  /** ISO timestamp of the last good sync, or null before the first one. */
  last_synced_at: string | null
  last_error: CalendarFeedErrorCode | null
  events_found: number
}

/** What the panel shows: sync is off on the server, nothing connected yet, or a calendar. */
export type CalendarFeedState =
  { kind: 'disabled' } | { kind: 'none' } | { kind: 'connected'; calendar: ConnectedCalendar }

export interface CalendarFeedProps {
  personId: string
}
