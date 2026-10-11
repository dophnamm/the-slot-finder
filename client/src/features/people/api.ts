import { PEOPLE_ENDPOINTS } from '@/resources/endpoints'
import { request, requestWithMeta } from '@/shared/api/http'
import type { CreatePersonInput, Person, SkillInput, UpdatePersonInput } from './types'

export function listPeople(page: number, pageSize: number) {
  return requestWithMeta<Person[]>({ url: PEOPLE_ENDPOINTS.list, params: { page, page_size: pageSize } })
}

export function createPerson(person: CreatePersonInput) {
  return request<Person>({ url: PEOPLE_ENDPOINTS.list, method: 'POST', data: person })
}

export function updatePerson(id: string, changes: UpdatePersonInput) {
  return request<Person>({ url: PEOPLE_ENDPOINTS.detail(id), method: 'PATCH', data: changes })
}

export function replaceSkills(id: string, skills: SkillInput[]) {
  return request<Person>({ url: PEOPLE_ENDPOINTS.skills(id), method: 'PUT', data: { skills } })
}
