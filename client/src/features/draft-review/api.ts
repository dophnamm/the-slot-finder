import { PEOPLE_ENDPOINTS, STAFFING_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import { MAX_PAGE_SIZE } from '@/utils/constants'
import type { DraftSummary, StaffablePerson } from './types'

export function listStaffablePeople() {
  return request<StaffablePerson[]>({ url: PEOPLE_ENDPOINTS.list, params: { page: 1, page_size: MAX_PAGE_SIZE } })
}

export function summarizeDraft(id: string) {
  return request<DraftSummary>({ url: STAFFING_ENDPOINTS.summary(id), method: 'POST' })
}
