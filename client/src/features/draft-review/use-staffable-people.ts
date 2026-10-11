import { useQuery } from '@tanstack/react-query'
import { listStaffablePeople } from './api'
import { STAFFABLE_PEOPLE_QUERY_KEY } from './constants'
import type { StaffablePerson } from './types'

function activeOnly(people: StaffablePerson[]) {
  return people.filter((person) => person.active)
}

/** Active people to add by hand; only loaded for someone who can edit the draft. */
export function useStaffablePeople(enabled: boolean) {
  return useQuery({ queryKey: STAFFABLE_PEOPLE_QUERY_KEY, queryFn: listStaffablePeople, select: activeOnly, enabled })
}
