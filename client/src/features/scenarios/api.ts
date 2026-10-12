import { PEOPLE_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import { MAX_PAGE_SIZE } from '@/utils/constants'
import type { PersonOption } from './types'

export function listPeopleOptions() {
  return request<PersonOption[]>({ url: PEOPLE_ENDPOINTS.list, params: { page: 1, page_size: MAX_PAGE_SIZE } })
}
