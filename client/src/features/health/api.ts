import { request } from '@/shared/api/http'
import { HEALTH_ENDPOINTS } from '@/resources/endpoints'
import type { Health } from './types'

export function getHealth() {
  return request<Health>({ url: HEALTH_ENDPOINTS.check })
}
