/** dayjs format strings used across the app. */
export const DATE_FORMAT = {
  /** Short day for ranges, e.g. "Sep 28". */
  shortDay: 'MMM D',
  /** Calendar date as the server sends it, e.g. "2026-09-28". */
  isoDate: 'YYYY-MM-DD',
  /** Day with the year, e.g. "Sep 28, 2026". */
  longDay: 'MMM D, YYYY',
} as const

/** ISO week label, e.g. "2026-W42". */
export const ISO_WEEK_PATTERN = /^(\d{4})-W(\d{2})$/

/** January 4th always falls in ISO week 1 of its year. */
export const ISO_WEEK_ONE_DAY = '01-04'
