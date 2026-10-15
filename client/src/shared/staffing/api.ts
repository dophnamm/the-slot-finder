import { STAFFING_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { Draft, DraftAcceptance, DraftEdit, OpenDraft } from './types'

export function generateDraft(projectIds: string[]) {
  return request<Draft>({ url: STAFFING_ENDPOINTS.drafts, method: 'POST', data: { project_ids: projectIds } })
}

export function listOpenDrafts() {
  return request<OpenDraft[]>({ url: STAFFING_ENDPOINTS.open })
}

export function getDraft(id: string) {
  return request<Draft>({ url: STAFFING_ENDPOINTS.draft(id) })
}

export function editDraft(id: string, edit: DraftEdit) {
  return request<Draft>({ url: STAFFING_ENDPOINTS.assignments(id), method: 'PATCH', data: edit })
}

export function acceptDraft(id: string, acceptance: DraftAcceptance) {
  return request<Draft>({ url: STAFFING_ENDPOINTS.accept(id), method: 'POST', data: acceptance })
}

export function rejectDraft(id: string) {
  return request<Draft>({ url: STAFFING_ENDPOINTS.reject(id), method: 'POST' })
}
