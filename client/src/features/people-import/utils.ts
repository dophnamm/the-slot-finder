import type { ApiError } from '@/shared/api/types'
import type { RowProblem } from '@/shared/components/problem-table'
import { PEOPLE_IMPORT_INVALID } from './constants'
import type { PeopleImportInvalidDetails } from './types'

/** The row problems of a failed import; empty for any other error. */
export function rowProblems(error: ApiError | null): RowProblem[] {
  if (error?.code !== PEOPLE_IMPORT_INVALID) return []
  return (error.details as Partial<PeopleImportInvalidDetails> | undefined)?.problems ?? []
}
