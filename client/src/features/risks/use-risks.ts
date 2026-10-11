import { useQuery } from '@tanstack/react-query'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { getRisks } from './api'
import { RISKS_QUERY_NAME } from './constants'

/** The team's risks; refreshed with the other insights when a staffing draft is accepted. */
export function useRisks() {
  return useQuery({ queryKey: [...INSIGHTS_QUERY_KEY, RISKS_QUERY_NAME], queryFn: getRisks })
}
