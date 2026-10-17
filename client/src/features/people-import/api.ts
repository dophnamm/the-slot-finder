import { PEOPLE_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import { PEOPLE_FORM_FIELD } from './constants'
import type { PeopleImportResult } from './types'

export function downloadPeopleTemplate() {
  return request<Blob>({ url: PEOPLE_ENDPOINTS.importTemplate, responseType: 'blob' })
}

export function importPeople(file: File) {
  const form = new FormData()
  form.append(PEOPLE_FORM_FIELD, file)
  return request<PeopleImportResult>({ url: PEOPLE_ENDPOINTS.import, method: 'POST', data: form })
}
