import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { listEvents } from './api'
import { EVENTS_QUERY_KEY, EVENTS_REFETCH_MS } from './constants'

/** The change history, newest first; with a program, only changes that touch its projects. */
export function useEvents(programId: string | undefined, page: number, pageSize: number) {
  return useQuery({
    queryKey: [...EVENTS_QUERY_KEY, programId ?? 'all', page, pageSize],
    queryFn: () => listEvents(programId, page, pageSize),
    placeholderData: keepPreviousData,
    refetchInterval: EVENTS_REFETCH_MS,
  })
}
