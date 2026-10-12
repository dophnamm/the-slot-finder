import { useQuery } from '@tanstack/react-query'
import { listPeopleOptions } from './api'
import { PEOPLE_OPTIONS_QUERY_KEY } from './constants'
import type { PersonOption } from './types'

function activeOnly(people: PersonOption[]) {
  return people.filter((person) => person.active)
}

/** Active people for leave assumptions. */
export function usePeopleOptions() {
  return useQuery({ queryKey: PEOPLE_OPTIONS_QUERY_KEY, queryFn: listPeopleOptions, select: activeOnly })
}
