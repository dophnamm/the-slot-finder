import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getPerson, getSchedule } from './api'
import { personQueryKey, scheduleQueryKey } from './constants'

export function usePerson(id: string) {
  return useQuery({ queryKey: personQueryKey(id), queryFn: () => getPerson(id) })
}

/** A person's schedule; the previous month stays on screen while the next one loads. */
export function useSchedule(personId: string, from: string, to: string) {
  return useQuery({
    queryKey: scheduleQueryKey(personId, from, to),
    queryFn: () => getSchedule(personId, from, to),
    placeholderData: keepPreviousData,
  })
}
