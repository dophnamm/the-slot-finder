import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { FORBIDDEN_CODE } from '@/shared/api/constants'
import { getPerson, getSchedule } from './api'
import { personQueryKey, scheduleQueryKey } from './constants'

/** One person; a 403 stays quiet because the person page says so itself. */
export function usePerson(id: string) {
  return useQuery({
    queryKey: personQueryKey(id),
    queryFn: () => getPerson(id),
    meta: { quietErrorCodes: [FORBIDDEN_CODE] },
  })
}

/** A person's schedule; the previous month stays on screen while the next one loads. */
export function useSchedule(personId: string, from: string, to: string) {
  return useQuery({
    queryKey: scheduleQueryKey(personId, from, to),
    queryFn: () => getSchedule(personId, from, to),
    placeholderData: keepPreviousData,
  })
}
