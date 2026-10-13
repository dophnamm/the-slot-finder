import { TIMESHEET_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import { TIMESHEET_FORM_FIELD } from './constants'
import type { ImportResult, WeekRange } from './types'

export function downloadTemplate(range: WeekRange) {
  return request<Blob>({ url: TIMESHEET_ENDPOINTS.template, params: range, responseType: 'blob' })
}

export function importTimesheet(file: File) {
  const form = new FormData()
  form.append(TIMESHEET_FORM_FIELD, file)
  return request<ImportResult>({ url: TIMESHEET_ENDPOINTS.import, method: 'POST', data: form })
}
