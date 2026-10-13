import { type QueryClient, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ABSENCES_QUERY_KEY, INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { connectCalendarFeed, disconnectCalendarFeed, getCalendarFeed, syncCalendarFeed } from './api'
import { CALENDAR_FEED_ERROR_CODES, calendarFeedQueryKey } from './constants'

/** A sync or a disconnect changes absences, so the panel, the absence lists and the calendar refresh together. */
function refreshAfterSync(queryClient: QueryClient, personId: string) {
  return Promise.all([
    queryClient.invalidateQueries({ queryKey: calendarFeedQueryKey(personId) }),
    queryClient.invalidateQueries({ queryKey: ABSENCES_QUERY_KEY }),
    queryClient.invalidateQueries({ queryKey: INSIGHTS_QUERY_KEY }),
  ])
}

export function useCalendarFeed(personId: string) {
  return useQuery({ queryKey: calendarFeedQueryKey(personId), queryFn: () => getCalendarFeed(personId) })
}

export function useConnectCalendarFeed(personId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (url: string) => connectCalendarFeed(personId, url),
    onSuccess: () => refreshAfterSync(queryClient, personId),
    meta: { quietErrorCodes: CALENDAR_FEED_ERROR_CODES },
  })
}

export function useSyncCalendarFeed(personId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => syncCalendarFeed(personId),
    onSuccess: () => refreshAfterSync(queryClient, personId),
  })
}

export function useDisconnectCalendarFeed(personId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => disconnectCalendarFeed(personId),
    // The leave the calendar added becomes ordinary absences: their marks go away.
    onSuccess: () => refreshAfterSync(queryClient, personId),
  })
}
