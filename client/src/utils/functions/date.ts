import dayjs from 'dayjs'
import isoWeek from 'dayjs/plugin/isoWeek'
import utc from 'dayjs/plugin/utc'
import { DATE_FORMAT, ISO_WEEK_ONE_DAY, ISO_WEEK_PATTERN } from '@/utils/constants'
import type { IsoWeekRange } from './types'

dayjs.extend(isoWeek)
dayjs.extend(utc)

export { dayjs }

/** ISO week label of a date: `toIsoWeek('2026-10-14')` → `'2026-W42'`. */
export function toIsoWeek(date: dayjs.ConfigType): string {
  const day = dayjs(date)
  return `${day.isoWeekYear()}-W${String(day.isoWeek()).padStart(2, '0')}`
}

/** Monday and Sunday of an ISO week label such as `'2026-W42'`. */
export function isoWeekRange(week: string): IsoWeekRange {
  const match = ISO_WEEK_PATTERN.exec(week)
  if (!match) throw new Error(`Not an ISO week label: ${week}`)
  const [, year, weekNumber] = match
  const start = dayjs(`${year}-${ISO_WEEK_ONE_DAY}`).isoWeek(Number(weekNumber)).startOf('isoWeek')
  return { start, end: start.endOf('isoWeek') }
}

/** Date range of an ISO week for display: `'2026-W42'` → `'Oct 12 – Oct 18'`. */
export function formatIsoWeekRange(week: string, format: string = DATE_FORMAT.shortDay): string {
  const { start, end } = isoWeekRange(week)
  return `${start.format(format)} – ${end.format(format)}`
}

/** A server date (`2026-10-19`) for display: `'Oct 19, 2026'`. */
export function formatDate(date: string, format: string = DATE_FORMAT.longDay): string {
  return dayjs(date, DATE_FORMAT.isoDate).format(format)
}

/** An inclusive date range for display; a single day shows once. */
export function formatDateRange(from: string, to: string): string {
  return from === to ? formatDate(from) : `${formatDate(from)} – ${formatDate(to)}`
}
