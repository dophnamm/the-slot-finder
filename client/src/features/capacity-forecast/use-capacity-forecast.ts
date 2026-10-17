import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { getCapacityForecast } from './api'
import { CAPACITY_FORECAST_QUERY_NAME } from './constants'
import type { CapacityForecastQuery } from './types'

/** The forecast for a week range; it refreshes with the other insights when the baseline changes. */
export function useCapacityForecast(query: CapacityForecastQuery) {
  return useQuery({
    queryKey: [...INSIGHTS_QUERY_KEY, CAPACITY_FORECAST_QUERY_NAME, query.from, query.to],
    queryFn: () => getCapacityForecast(query),
    placeholderData: keepPreviousData,
  })
}
