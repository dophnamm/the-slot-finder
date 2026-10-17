import { INSIGHTS_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { CapacityForecastData, CapacityForecastQuery } from './types'

export function getCapacityForecast({ from, to }: CapacityForecastQuery) {
  return request<CapacityForecastData>({ url: INSIGHTS_ENDPOINTS.capacityForecast, params: { from, to } })
}
