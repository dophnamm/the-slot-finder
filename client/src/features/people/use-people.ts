import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createPerson, listExpertises, listPeople, replaceSkills, updatePerson } from './api'
import { EXPERTISES_QUERY_KEY, PEOPLE_PAGE_SIZE, PEOPLE_QUERY_KEY } from './constants'
import type { SavePersonInput } from './types'

export function usePeople(page: number) {
  return useQuery({
    queryKey: [...PEOPLE_QUERY_KEY, page],
    queryFn: () => listPeople(page, PEOPLE_PAGE_SIZE),
    placeholderData: keepPreviousData,
  })
}

export function useExpertises() {
  return useQuery({ queryKey: EXPERTISES_QUERY_KEY, queryFn: listExpertises })
}

async function savePerson(input: SavePersonInput) {
  if (input.mode === 'create') return createPerson(input.person)
  const person = await updatePerson(input.id, input.changes)
  return input.skills ? replaceSkills(input.id, input.skills) : person
}

/** Creates or edits a person (and their skills), then refetches the list. */
export function useSavePerson() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: savePerson,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: PEOPLE_QUERY_KEY }),
  })
}
