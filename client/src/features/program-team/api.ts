import { PEOPLE_ENDPOINTS, PROGRAM_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { Person } from '@/shared/people'
import { MAX_PAGE_SIZE } from '@/utils/constants'
import type { TeamMember } from './types'

export function listProgramTeam(programId: string) {
  return request<TeamMember[]>({ url: PROGRAM_ENDPOINTS.team(programId) })
}

export function addTeamMember(programId: string, personId: string) {
  return request<null>({ url: PROGRAM_ENDPOINTS.team(programId), method: 'POST', data: { person_id: personId } })
}

export function removeTeamMember(programId: string, personId: string) {
  return request<null>({ url: PROGRAM_ENDPOINTS.teamMember(programId, personId), method: 'DELETE' })
}

export function listTeamCandidates() {
  return request<Person[]>({ url: PEOPLE_ENDPOINTS.list, params: { page: 1, page_size: MAX_PAGE_SIZE } })
}
