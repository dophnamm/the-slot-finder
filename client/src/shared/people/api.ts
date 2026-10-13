import { PEOPLE_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { Person, PersonSchedule } from './types'

export function getPerson(id: string) {
  return request<Person>({ url: PEOPLE_ENDPOINTS.detail(id) })
}

/** One person's weeks and days from one ISO week through another. */
export function getSchedule(personId: string, from: string, to: string) {
  return request<PersonSchedule>({ url: PEOPLE_ENDPOINTS.schedule(personId), params: { from, to } })
}
