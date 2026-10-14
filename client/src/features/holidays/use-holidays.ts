import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { createHoliday, deleteHoliday, importHolidays, listHolidays } from './api'
import { HOLIDAYS_QUERY_KEY } from './constants'

export function useHolidays() {
  return useQuery({ queryKey: HOLIDAYS_QUERY_KEY, queryFn: listHolidays })
}

export function useCreateHoliday() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createHoliday,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: HOLIDAYS_QUERY_KEY })
      await queryClient.invalidateQueries({ queryKey: INSIGHTS_QUERY_KEY })
    },
  })
}

/** Imports a holiday calendar; the holidays and every capacity figure refetch afterwards. */
export function useImportHolidays() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: importHolidays,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: HOLIDAYS_QUERY_KEY })
      await queryClient.invalidateQueries({ queryKey: INSIGHTS_QUERY_KEY })
    },
  })
}

export function useDeleteHoliday() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteHoliday,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: HOLIDAYS_QUERY_KEY })
      await queryClient.invalidateQueries({ queryKey: INSIGHTS_QUERY_KEY })
    },
  })
}
