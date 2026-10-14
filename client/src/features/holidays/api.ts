import { ABSENCE_ENDPOINTS, HOLIDAY_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { CreateHolidayInput, Holiday, HolidayImportResult, ImportHolidaysInput } from './types'
import { toImportForm } from './utils'

export function listHolidays() {
  return request<Holiday[]>({ url: HOLIDAY_ENDPOINTS.list })
}

export function createHoliday(holiday: CreateHolidayInput) {
  return request<Holiday>({ url: HOLIDAY_ENDPOINTS.list, method: 'POST', data: holiday })
}

export function importHolidays(input: ImportHolidaysInput) {
  return request<HolidayImportResult>({ url: HOLIDAY_ENDPOINTS.import, method: 'POST', data: toImportForm(input) })
}

/** Holidays are absences without a person, so they are removed through the absence endpoint. */
export function deleteHoliday(id: string) {
  return request<null>({ url: ABSENCE_ENDPOINTS.detail(id), method: 'DELETE' })
}
