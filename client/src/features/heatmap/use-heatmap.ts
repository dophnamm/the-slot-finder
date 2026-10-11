import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { getHeatmap, listPersonWeek } from './api'
import { HEATMAP_QUERY_NAME, PERSON_WEEK_QUERY_NAME } from './constants'
import type { HeatmapQuery } from './types'

/** The heatmap for the filters; it refreshes when a staffing draft is accepted. */
export function useHeatmap(query: HeatmapQuery) {
  return useQuery({
    queryKey: [...INSIGHTS_QUERY_KEY, HEATMAP_QUERY_NAME, query.from, query.to, query.expertise],
    queryFn: () => getHeatmap(query),
    placeholderData: keepPreviousData,
  })
}

/** One person's assignments in one week, for a cell's detail. */
export function usePersonWeek(personId: string, week: string) {
  return useQuery({
    queryKey: [...INSIGHTS_QUERY_KEY, PERSON_WEEK_QUERY_NAME, personId, week],
    queryFn: () => listPersonWeek(personId, week),
  })
}
