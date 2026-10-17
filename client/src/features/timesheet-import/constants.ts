import type { PlainKey } from '@/shared/i18n'
import type { RowProblemCode } from './types'

/** Error code of an import with row problems; the dialog lists them instead of a toast. */
export const TIMESHEET_INVALID = 'TIMESHEET_INVALID'

/** Multipart field the server reads the file from. */
export const TIMESHEET_FORM_FIELD = 'file'

/** Files the picker offers. */
export const TIMESHEET_FILE_ACCEPT = '.csv,text/csv'

/** Past weeks the template covers by default, up to the current week (TIMESHEET_TEMPLATE_WEEKS). */
export const DEFAULT_TEMPLATE_WEEKS = 4

/** The downloaded template is named `timesheet-<from>-<to>.csv`. */
export const TEMPLATE_FILE_PREFIX = 'timesheet'
export const CSV_EXTENSION = '.csv'

/** Names and ids of the dialog's fields. */
export const RANGE_FIELDS = { from: 'from', to: 'to' } as const
export const FIELD_IDS = { from: 'timesheet-from', to: 'timesheet-to', file: 'timesheet-file' } as const

/** What each problem code means. */
export const ROW_PROBLEM_LABELS: Record<RowProblemCode, PlainKey> = {
  COLUMN_MISSING: 'The header lacks this column.',
  CSV_MALFORMED: 'The file is not valid CSV from this line.',
  PERSON_UNKNOWN: 'No person has this email.',
  PROJECT_UNKNOWN: 'No project has this id.',
  WEEK_INVALID: 'Not a week such as 2026-W41.',
  WEEK_IN_FUTURE: 'The week is after the current week.',
  HOURS_INVALID: 'Hours must be a number from 0 to the weekly limit.',
  ROW_DUPLICATE: 'The same person, project and week appear more than once.',
}
