import type { ApiError } from '@/shared/api/types'
import type { PlainKey } from '@/shared/i18n'
import { addIsoWeeks, currentIsoWeek } from '@/utils/functions'
import {
  CSV_EXTENSION,
  DEFAULT_TEMPLATE_WEEKS,
  ROW_PROBLEM_LABELS,
  TEMPLATE_FILE_PREFIX,
  TIMESHEET_INVALID,
} from './constants'
import type { RowProblem, RowProblemCode, TimesheetInvalidDetails, WeekRange } from './types'

/** The last DEFAULT_TEMPLATE_WEEKS weeks, up to the current one. */
export function defaultRange(): WeekRange {
  const to = currentIsoWeek()
  return { from: addIsoWeeks(to, 1 - DEFAULT_TEMPLATE_WEEKS), to }
}

/** The downloaded template's name, e.g. timesheet-2026-W39-2026-W42.csv. */
export function templateFileName({ from, to }: WeekRange): string {
  return [TEMPLATE_FILE_PREFIX, from, to].join('-') + CSV_EXTENSION
}

/** The row problems of a failed import; empty for any other error. */
export function rowProblems(error: ApiError | null): RowProblem[] {
  if (error?.code !== TIMESHEET_INVALID) return []
  return (error.details as Partial<TimesheetInvalidDetails> | undefined)?.problems ?? []
}

/** What a problem code means, or null for a code this client does not know yet. */
export function problemLabel(code: string): PlainKey | null {
  return code in ROW_PROBLEM_LABELS ? ROW_PROBLEM_LABELS[code as RowProblemCode] : null
}
