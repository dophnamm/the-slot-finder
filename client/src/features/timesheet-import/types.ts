import type { RowProblem } from '@/shared/components/problem-table'

/** Problem codes the server reports for a timesheet import (TIMESHEET_INVALID). */
export type RowProblemCode =
  | 'COLUMN_MISSING'
  | 'CSV_MALFORMED'
  | 'PERSON_UNKNOWN'
  | 'PROJECT_UNKNOWN'
  | 'WEEK_INVALID'
  | 'WEEK_IN_FUTURE'
  | 'HOURS_INVALID'
  | 'ROW_DUPLICATE'

/** The data of a TIMESHEET_INVALID error. */
export interface TimesheetInvalidDetails {
  problems: RowProblem[]
}

/** What an import saved: new rows, replaced values and rows skipped for an empty actual_hours. */
export interface ImportResult {
  created: number
  updated: number
  skipped: number
}

/** The weeks a template covers. */
export interface WeekRange {
  from: string
  to: string
}
