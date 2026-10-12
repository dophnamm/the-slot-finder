import { EVENT_ENDPOINTS } from '@/resources/endpoints'
import { requestWithMeta } from '@/shared/api/http'
import type { ChangeEvent } from './types'

export function listEvents(programId: string | undefined, page: number, pageSize: number) {
  return requestWithMeta<ChangeEvent[]>({
    url: EVENT_ENDPOINTS.list,
    params: { program_id: programId, page, page_size: pageSize },
  })
}
