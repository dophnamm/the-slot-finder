import dayjs from 'dayjs'
import isoWeek from 'dayjs/plugin/isoWeek'
import utc from 'dayjs/plugin/utc'
import { DATE_FORMAT, ISO_WEEK_ONE_DAY, ISO_WEEK_PATTERN, PROJECT_WINDOW_WEEKS } from '@/utils/constants'
import type { IsoWeekRange } from './types'

dayjs.extend(isoWeek)
dayjs.extend(utc)

export { dayjs }

/** ISO week label of a date: `toIsoWeek('2026-10-14')` → `'2026-W42'`. */
export function toIsoWeek(date: dayjs.ConfigType): string {
  const day = dayjs(date)
  return `${day.isoWeekYear()}-W${String(day.isoWeek()).padStart(2, '0')}`
}

/** A date value (`YYYY-MM-DD`) as a Date at local midnight, e.g. for a calendar; undefined when empty. */
export function parseDateValue(value: string): Date | undefined {
  return value ? dayjs(value).toDate() : undefined
}

/** A Date as a date value (`YYYY-MM-DD`), the format the server and the forms use. */
export function toDateValue(date: Date): string {
  return dayjs(date).format(DATE_FORMAT.isoDate)
}

/** The week of today, in the browser's time zone. */
export function currentIsoWeek(): string {
  return toIsoWeek(dayjs())
}

/** Whether a value is an ISO week label such as `'2026-W42'`. */
export function isIsoWeek(value: string): boolean {
  return ISO_WEEK_PATTERN.test(value)
}

/** The week part of a label for narrow columns: `'2026-W42'` → `'W42'`. */
export function shortWeekLabel(week: string): string {
  const match = ISO_WEEK_PATTERN.exec(week)
  return match ? `W${match[2]}` : week
}

/** Monday and Sunday of an ISO week label such as `'2026-W42'`. */
export function isoWeekRange(week: string): IsoWeekRange {
  const match = ISO_WEEK_PATTERN.exec(week)
  if (!match) throw new Error(`Not a week label: ${week}`)
  const [, year, weekNumber] = match
  const start = dayjs(`${year}-${ISO_WEEK_ONE_DAY}`).isoWeek(Number(weekNumber)).startOf('isoWeek')
  return { start, end: start.endOf('isoWeek') }
}

/** Date range of a week for display: `'2026-W42'` → `'Oct 12 – Oct 18'`. */
export function formatIsoWeekRange(week: string, format: string = DATE_FORMAT.shortDay): string {
  const { start, end } = isoWeekRange(week)
  return `${start.format(format)} – ${end.format(format)}`
}

/** A server date (`2026-10-19`) for display: `'Oct 19, 2026'`. */
export function formatDate(date: string, format: string = DATE_FORMAT.longDay): string {
  return dayjs(date, DATE_FORMAT.isoDate).format(format)
}

/** A server timestamp in local time for display: `'Oct 11, 14:05'`. */
export function formatDateTime(timestamp: string, format: string = DATE_FORMAT.dateTime): string {
  return dayjs(timestamp).format(format)
}

/** An inclusive date range for display; a single day shows once. */
export function formatDateRange(from: string, to: string): string {
  return from === to ? formatDate(from) : `${formatDate(from)} – ${formatDate(to)}`
}

/** Dates from the Monday of `from` to the Sunday of `to`: `'Oct 12 – Nov 22'`. */
export function formatIsoWeekSpan(from: string, to: string, format: string = DATE_FORMAT.shortDay): string {
  return `${isoWeekRange(from).start.format(format)} – ${isoWeekRange(to).end.format(format)}`
}

/** The week `n` weeks after `week` (before it when `n` is negative). */
export function addIsoWeeks(week: string, n: number): string {
  return toIsoWeek(isoWeekRange(week).start.add(n, 'week'))
}

/** Weeks from `from` to `to`; negative when `to` is earlier. */
export function isoWeeksBetween(from: string, to: string): number {
  return isoWeekRange(to).start.diff(isoWeekRange(from).start, 'week')
}

/** Every week from `from` through `to`, in order. */
export function isoWeekList(from: string, to: string): string[] {
  const count = isoWeeksBetween(from, to)
  return Array.from({ length: Math.max(0, count) + 1 }, (_, i) => addIsoWeeks(from, i))
}

/** The month of today, in the browser's time zone: `'2026-10'`. */
export function currentMonth(): string {
  return dayjs().format(DATE_FORMAT.isoMonth)
}

/** The month `n` months after `month` (before it when `n` is negative): `('2026-12', 1)` → `'2027-01'`. */
export function addMonths(month: string, n: number): string {
  return dayjs(month).add(n, 'month').format(DATE_FORMAT.isoMonth)
}

/** Every week with a day in the month, in order: `'2026-10'` → `'2026-W40'` … `'2026-W44'`. */
export function monthIsoWeeks(month: string): string[] {
  const first = dayjs(month).startOf('month')
  return isoWeekList(toIsoWeek(first), toIsoWeek(first.endOf('month')))
}

/** A month for display: `'2026-10'` → `'October 2026'`. */
export function formatMonth(month: string, format: string = DATE_FORMAT.monthTitle): string {
  return dayjs(month).format(format)
}

/** Whether a date value (`'2026-10-31'`) falls in a month (`'2026-10'`). */
export function isInMonth(date: string, month: string): boolean {
  return dayjs(date).format(DATE_FORMAT.isoMonth) === month
}

/** The date of a weekday (1 = Monday … 7 = Sunday) in a week: `('2026-W42', 3)` → `'2026-10-14'`. */
export function isoWeekDate(week: string, isoWeekday: number): string {
  return isoWeekRange(week).start.isoWeekday(isoWeekday).format(DATE_FORMAT.isoDate)
}

/** The short name of a weekday for a column header: `1` → `'Mon'`. */
export function formatIsoWeekday(isoWeekday: number, format: string = DATE_FORMAT.weekdayShort): string {
  return dayjs().isoWeekday(isoWeekday).format(format)
}

/** Today as a date value (`'2026-10-14'`), in the browser's time zone. */
export function todayDateValue(): string {
  return dayjs().format(DATE_FORMAT.isoDate)
}

/** The weeks and days that project and phase dates may use: `PROJECT_WINDOW_WEEKS` either side of today's week. */
export function planningWindow(today: dayjs.ConfigType = dayjs()) {
  const week = toIsoWeek(today)
  const minWeek = addIsoWeeks(week, -PROJECT_WINDOW_WEEKS)
  const maxWeek = addIsoWeeks(week, PROJECT_WINDOW_WEEKS)
  return { minWeek, maxWeek, minDate: isoWeekDate(minWeek, 1), maxDate: isoWeekDate(maxWeek, 7) }
}

/** Whether a week is inside the planning window. */
export function isWeekInWindow(week: string, today?: dayjs.ConfigType): boolean {
  const { minWeek, maxWeek } = planningWindow(today)
  return isIsoWeek(week) && week >= minWeek && week <= maxWeek
}

/** Whether a date value (`YYYY-MM-DD`) is inside the planning window. */
export function isDateInWindow(date: string, today?: dayjs.ConfigType): boolean {
  return isWeekInWindow(toIsoWeek(date), today)
}
