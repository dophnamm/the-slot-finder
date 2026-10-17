import { useMutation } from '@tanstack/react-query'
import { AI_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { IntakeRequest, IntakeSuggestion } from './types'

export function extractIntake(body: IntakeRequest) {
  return request<IntakeSuggestion>({ url: AI_ENDPOINTS.intake, method: 'POST', data: body })
}

/** Errors (429, 502, 503) go to the shared toast. */
export function useExtractIntake() {
  return useMutation({ mutationFn: extractIntake })
}
