import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { FORBIDDEN_CODE } from '@/shared/api/constants'
import {
  applyProgramAction,
  createProgram,
  getProgram,
  getProgramOverview,
  listLeadCandidates,
  listProgramMembers,
  listPrograms,
  updateProgram,
} from './api'
import {
  LEAD_OPTIONS_QUERY_KEY,
  LEAD_ROLES,
  PROGRAM_OPTIONS_PAGE_SIZE,
  PROGRAMS_PAGE_SIZE,
  PROGRAMS_QUERY_KEY,
  PROJECTS_QUERY_KEY,
  programMembersQueryKey,
  programOverviewQueryKey,
  programQueryKey,
} from './constants'
import type { LeadOption, ProgramAction, ProgramInput } from './types'

export function usePrograms(page: number) {
  return useQuery({
    queryKey: [...PROGRAMS_QUERY_KEY, 'page', page],
    queryFn: () => listPrograms(page, PROGRAMS_PAGE_SIZE),
    placeholderData: keepPreviousData,
  })
}

/** Every program the user may read, for a select; a manager gets only theirs. */
export function useProgramOptions() {
  return useQuery({
    queryKey: [...PROGRAMS_QUERY_KEY, 'options'],
    queryFn: () => listPrograms(1, PROGRAM_OPTIONS_PAGE_SIZE),
    select: (result) => result.data,
  })
}

/** One program; a 403 stays quiet because the program page says so itself. */
export function useProgram(id: string) {
  return useQuery({
    queryKey: programQueryKey(id),
    queryFn: () => getProgram(id),
    meta: { quietErrorCodes: [FORBIDDEN_CODE] },
  })
}

/** Everyone with baseline hours on the program; refetched with programs after a draft is accepted. */
export function useProgramMembers(id: string) {
  return useQuery({ queryKey: programMembersQueryKey(id), queryFn: () => listProgramMembers(id) })
}

/** Creates a program, or edits one when `id` is given; refetches programs afterwards. */
export function useSaveProgram(id?: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (program: ProgramInput | Partial<ProgramInput>) =>
      id ? updateProgram(id, program) : createProgram(program as ProgramInput),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: PROGRAMS_QUERY_KEY }),
  })
}

/** Changes a program's status; a cancel also cancels its projects, so projects refetch too. */
export function useProgramAction(id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (action: ProgramAction) => applyProgramAction(id, action),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: PROGRAMS_QUERY_KEY })
      await queryClient.invalidateQueries({ queryKey: PROJECTS_QUERY_KEY })
    },
  })
}

/** People with a role that can lead (LEAD_ROLES); only fetched when `enabled` (planners). */
/** Progress, cost figures and staffing by expertise; refetched with the program's other data. */
export function useProgramOverview(id: string, enabled = true) {
  return useQuery({ queryKey: programOverviewQueryKey(id), queryFn: () => getProgramOverview(id), enabled })
}

export function useLeadOptions(enabled: boolean) {
  return useQuery({
    queryKey: LEAD_OPTIONS_QUERY_KEY,
    queryFn: listLeadCandidates,
    enabled,
    select: (people): LeadOption[] =>
      people
        .filter((person) => (LEAD_ROLES as readonly string[]).includes(person.role))
        .map(({ id, name }) => ({ id, name })),
  })
}
