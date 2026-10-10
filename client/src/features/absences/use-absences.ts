import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createAbsence, deleteAbsence, listAbsences } from './api'
import { absencesQueryKey } from './constants'
import type { CreateAbsenceInput } from './types'

export function useAbsences(personId: string) {
  return useQuery({ queryKey: absencesQueryKey(personId), queryFn: () => listAbsences(personId) })
}

export function useCreateAbsence(personId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (absence: CreateAbsenceInput) => createAbsence(personId, absence),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: absencesQueryKey(personId) }),
  })
}

export function useDeleteAbsence(personId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteAbsence,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: absencesQueryKey(personId) }),
  })
}
