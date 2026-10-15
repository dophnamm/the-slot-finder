import { INSIGHTS_ENDPOINTS, PEOPLE_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { Assignment } from '@/shared/staffing'
import type { HeatmapData, HeatmapQuery } from './types'

export function getHeatmap({ from, to, expertise, programId }: HeatmapQuery) {
  return request<HeatmapData>({
    url: INSIGHTS_ENDPOINTS.heatmap,
    params: { from, to, expertise: expertise || undefined, program_id: programId || undefined },
  })
}

/** A person's baseline assignments with their hours in one week. */
export function listPersonWeek(personId: string, week: string) {
  return request<Assignment[]>({ url: PEOPLE_ENDPOINTS.assignments(personId), params: { from: week, to: week } })
}
