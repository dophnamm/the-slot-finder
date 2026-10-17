import type { RowProblem } from '@/shared/components/problem-table'

/** Problem codes the server reports for a people import (PEOPLE_IMPORT_INVALID). */
export type PeopleImportProblemCode =
  | 'COLUMN_MISSING'
  | 'CSV_MALFORMED'
  | 'ROW_DUPLICATE'
  | 'NAME_INVALID'
  | 'EMAIL_INVALID'
  | 'ROLE_INVALID'
  | 'HOURS_INVALID'
  | 'DATE_INVALID'
  | 'EXPERTISE_UNKNOWN'
  | 'LEVEL_INVALID'
  | 'SKILLS_INVALID'
  | 'SKILL_DUPLICATE'

/** The data of a PEOPLE_IMPORT_INVALID error. */
export interface PeopleImportInvalidDetails {
  problems: RowProblem[]
}

/** What an import saved. */
export interface PeopleImportResult {
  created: number
  updated: number
  unchanged: number
}
