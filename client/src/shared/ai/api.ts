import { AI_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { AskAnswer, AskRequest } from './types'

export function askAi(body: AskRequest) {
  return request<AskAnswer>({ url: AI_ENDPOINTS.ask, method: 'POST', data: body })
}
