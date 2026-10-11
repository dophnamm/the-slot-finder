import { INSIGHTS_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { RisksData } from './types'

export function getRisks() {
  return request<RisksData>({ url: INSIGHTS_ENDPOINTS.risks })
}
