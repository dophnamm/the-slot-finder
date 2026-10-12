import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createScenario, deleteScenario, getScenario, listScenarios, rerunScenario } from './api'
import { SCENARIOS_QUERY_KEY, scenarioQueryKey } from './constants'
import type { ScenarioInput } from './types'

export function useScenarios() {
  return useQuery({ queryKey: SCENARIOS_QUERY_KEY, queryFn: listScenarios })
}

export function useScenario(id: string) {
  return useQuery({ queryKey: scenarioQueryKey(id), queryFn: () => getScenario(id) })
}

/** Creates and computes a scenario; the list refetches. */
export function useCreateScenario() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: ScenarioInput) => createScenario(input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: SCENARIOS_QUERY_KEY }),
  })
}

/** Computes a scenario again against the current baseline. */
export function useRerunScenario(id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => rerunScenario(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: SCENARIOS_QUERY_KEY }),
  })
}

export function useDeleteScenario(id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => deleteScenario(id),
    onSuccess: () => queryClient.removeQueries({ queryKey: scenarioQueryKey(id) }),
    onSettled: () => queryClient.invalidateQueries({ queryKey: SCENARIOS_QUERY_KEY }),
  })
}
