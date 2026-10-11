import { EXPERTISE_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { Expertise } from './types'

export function listExpertises() {
  return request<Expertise[]>({ url: EXPERTISE_ENDPOINTS.list })
}
