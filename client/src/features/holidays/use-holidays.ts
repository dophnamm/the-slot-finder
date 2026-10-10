import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createHoliday, deleteHoliday, listHolidays } from './api'
import { HOLIDAYS_QUERY_KEY } from './constants'

export function useHolidays() {
  return useQuery({ queryKey: HOLIDAYS_QUERY_KEY, queryFn: listHolidays })
}

export function useCreateHoliday() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createHoliday,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: HOLIDAYS_QUERY_KEY }),
  })
}

export function useDeleteHoliday() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteHoliday,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: HOLIDAYS_QUERY_KEY }),
  })
}
