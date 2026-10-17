import { useMutation, useQueryClient } from '@tanstack/react-query'
import { PEOPLE_QUERY_KEY } from '@/shared/people'
import { saveFile } from '@/utils/functions'
import { downloadPeopleTemplate, importPeople } from './api'
import { PEOPLE_IMPORT_INVALID, PEOPLE_TEMPLATE_FILE_NAME } from './constants'

/** Downloads the people template and saves it as a CSV file. */
export function useDownloadPeopleTemplate() {
  return useMutation({
    mutationFn: async () => saveFile(await downloadPeopleTemplate(), PEOPLE_TEMPLATE_FILE_NAME),
  })
}

/** Imports a filled-in file. Row problems stay in the dialog instead of a toast. */
export function useImportPeople() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => importPeople(file),
    meta: { quietErrorCodes: [PEOPLE_IMPORT_INVALID] },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: PEOPLE_QUERY_KEY }),
  })
}
