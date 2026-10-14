import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { addTeamMember, listProgramTeam, listTeamCandidates, removeTeamMember } from './api'
import { programTeamQueryKey, TEAM_CANDIDATES_QUERY_KEY } from './constants'

export function useProgramTeam(programId: string) {
  return useQuery({ queryKey: programTeamQueryKey(programId), queryFn: () => listProgramTeam(programId) })
}

/** People a planner may add; only loaded while the add dialog is open. */
export function useTeamCandidates(enabled: boolean) {
  return useQuery({ queryKey: TEAM_CANDIDATES_QUERY_KEY, queryFn: listTeamCandidates, enabled })
}

export function useAddTeamMember(programId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (personId: string) => addTeamMember(programId, personId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: programTeamQueryKey(programId) }),
  })
}

export function useRemoveTeamMember(programId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (personId: string) => removeTeamMember(programId, personId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: programTeamQueryKey(programId) }),
  })
}
