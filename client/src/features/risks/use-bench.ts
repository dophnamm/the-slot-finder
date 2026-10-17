import { useQuery } from '@tanstack/react-query'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { getBench } from './api'
import { BENCH_QUERY_NAME } from './constants'

/** People on the bench; refreshed with the other insights when a staffing draft is accepted. */
export function useBench() {
  return useQuery({ queryKey: [...INSIGHTS_QUERY_KEY, BENCH_QUERY_NAME], queryFn: getBench })
}
