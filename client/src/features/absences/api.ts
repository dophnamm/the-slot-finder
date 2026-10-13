import { ABSENCE_ENDPOINTS, PEOPLE_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { Absence, CreateAbsenceInput } from './types'

export function listAbsences(personId: string) {
  return request<Absence[]>({ url: PEOPLE_ENDPOINTS.absences(personId) })
}

export function createAbsence(personId: string, absence: CreateAbsenceInput) {
  return request<Absence>({ url: PEOPLE_ENDPOINTS.absences(personId), method: 'POST', data: absence })
}

export function deleteAbsence(id: string) {
  return request<null>({ url: ABSENCE_ENDPOINTS.detail(id), method: 'DELETE' })
}

/** Turns a planned absence into an approved one (planners only). */
export function approveAbsence(id: string) {
  return request<Absence>({ url: ABSENCE_ENDPOINTS.approve(id), method: 'POST' })
}
