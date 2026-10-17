import type { PlainKey } from '@/shared/i18n'
import type { PeopleImportProblemCode } from './types'

/** Error code of an import with row problems; the dialog lists them instead of a toast. */
export const PEOPLE_IMPORT_INVALID = 'PEOPLE_IMPORT_INVALID'

/** Multipart field the server reads the file from. */
export const PEOPLE_FORM_FIELD = 'file'

/** Files the picker offers. */
export const PEOPLE_FILE_ACCEPT = '.csv,text/csv'

/** Name of the downloaded template. */
export const PEOPLE_TEMPLATE_FILE_NAME = 'people.csv'

export const FILE_FIELD_ID = 'people-file'

/** What each problem code means. */
export const ROW_PROBLEM_LABELS: Record<PeopleImportProblemCode, PlainKey> = {
  COLUMN_MISSING: 'The header lacks this column.',
  CSV_MALFORMED: 'The file is not valid CSV from this line.',
  ROW_DUPLICATE: 'The same email appears more than once.',
  NAME_INVALID: 'The name is empty or too long.',
  EMAIL_INVALID: 'Not a valid email address.',
  ROLE_INVALID: 'The role must be planner, manager or specialist.',
  HOURS_INVALID: 'Hours must be a number within the allowed limit.',
  DATE_INVALID: 'Not a date such as 2026-10-19, or the end is before the start.',
  EXPERTISE_UNKNOWN: 'No expertise has this name.',
  LEVEL_INVALID: 'The level must be a whole number from 1 to 5.',
  SKILLS_INVALID: 'Skills must look like QA:2;Analyst:1.',
  SKILL_DUPLICATE: 'An expertise appears more than once for this person.',
}
