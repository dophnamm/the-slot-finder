/** dayjs format strings used across the app. */
export const DATE_FORMAT = {
  /** Short day for ranges, e.g. "Sep 28". */
  shortDay: 'MMM D',
  /** Calendar date as the server sends it, e.g. "2026-09-28". */
  isoDate: 'YYYY-MM-DD',
  /** Day with the year, e.g. "Sep 28, 2026". */
  longDay: 'MMM D, YYYY',
  /** Day and local time of a timestamp, e.g. "Sep 28, 14:05". */
  dateTime: 'MMM D, HH:mm',
  /** A month as the calendar keeps it, e.g. "2026-10". */
  isoMonth: 'YYYY-MM',
  /** A month heading, e.g. "October 2026". */
  monthTitle: 'MMMM YYYY',
  /** Day of the month in a calendar cell, e.g. "14". */
  dayOfMonth: 'D',
  /** Short weekday in a column header, e.g. "Mon". */
  weekdayShort: 'ddd',
} as const

/** Weeks either side of the current week that project and phase dates may fall in; mirrors the server's PROJECT_WINDOW_WEEKS. */
export const PROJECT_WINDOW_WEEKS = 156

/** ISO week label, e.g. "2026-W42". */
export const ISO_WEEK_PATTERN = /^(\d{4})-W(\d{2})$/

/** January 4th always falls in ISO week 1 of its year. */
export const ISO_WEEK_ONE_DAY = '01-04'
