import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
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
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: absencesQueryKey(personId) })
      await queryClient.invalidateQueries({ queryKey: INSIGHTS_QUERY_KEY })
    },
  })
}

export function useDeleteAbsence(personId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteAbsence,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: absencesQueryKey(personId) })
      await queryClient.invalidateQueries({ queryKey: INSIGHTS_QUERY_KEY })
    },
  })
}
