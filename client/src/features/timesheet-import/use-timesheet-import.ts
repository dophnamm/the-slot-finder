import { useMutation, useQueryClient } from '@tanstack/react-query'
import { saveFile } from '@/utils/functions'
import { downloadTemplate, importTimesheet } from './api'
import { TIMESHEET_INVALID } from './constants'
import type { WeekRange } from './types'
import { templateFileName } from './utils'

/** Downloads the template for a range of weeks and saves it as a CSV file. */
export function useDownloadTemplate() {
  return useMutation({
    mutationFn: async (range: WeekRange) => saveFile(await downloadTemplate(range), templateFileName(range)),
  })
}

/**
 * Imports a filled-in template. Row problems stay in the dialog instead of a toast. Actual
 * hours feed the program, project and person views, so every active query refetches.
 */
export function useImportTimesheet() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => importTimesheet(file),
    meta: { quietErrorCodes: [TIMESHEET_INVALID] },
    onSuccess: () => queryClient.invalidateQueries(),
  })
}
