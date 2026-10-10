import { ABSENCE_ENDPOINTS, HOLIDAY_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { CreateHolidayInput, Holiday } from './types'

export function listHolidays() {
  return request<Holiday[]>({ url: HOLIDAY_ENDPOINTS.list })
}

export function createHoliday(holiday: CreateHolidayInput) {
  return request<Holiday>({ url: HOLIDAY_ENDPOINTS.list, method: 'POST', data: holiday })
}

/** Holidays are absences without a person, so they are removed through the absence endpoint. */
export function deleteHoliday(id: string) {
  return request<null>({ url: ABSENCE_ENDPOINTS.detail(id), method: 'DELETE' })
}
