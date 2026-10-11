import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  applyProjectAction,
  createProject,
  getProject,
  listProgramProjects,
  listProjects,
  replaceDemands,
  replacePhases,
  updatePhase,
  updateProject,
} from './api'
import {
  PROGRAMS_PAGE_SIZE,
  PROGRAMS_QUERY_KEY,
  PROJECTS_QUERY_KEY,
  programProjectsQueryKey,
  projectQueryKey,
  projectsQueryKey,
} from './constants'
import type { DemandInput, PhaseChanges, PhaseInput, ProjectAction, ProjectInput, ProjectStatus } from './types'

export function useProgramProjects(programId: string) {
  return useQuery({ queryKey: programProjectsQueryKey(programId), queryFn: () => listProgramProjects(programId) })
}

export function useProjects(status: ProjectStatus | undefined, page: number) {
  return useQuery({
    queryKey: [...projectsQueryKey(status ?? 'all'), page],
    queryFn: () => listProjects(status, page, PROGRAMS_PAGE_SIZE),
    placeholderData: keepPreviousData,
  })
}

export function useProject(id: string) {
  return useQuery({ queryKey: projectQueryKey(id), queryFn: () => getProject(id) })
}

/** After any project write: refetch every project list and the program's project rows. */
function useInvalidateProjects() {
  const queryClient = useQueryClient()
  return async () => {
    await queryClient.invalidateQueries({ queryKey: PROJECTS_QUERY_KEY })
    await queryClient.invalidateQueries({ queryKey: PROGRAMS_QUERY_KEY })
  }
}

/** Adds a project to a program, or edits one when `projectId` is given. */
export function useSaveProject(programId: string, projectId?: string) {
  const invalidate = useInvalidateProjects()
  return useMutation({
    mutationFn: (project: ProjectInput | Partial<ProjectInput>) =>
      projectId ? updateProject(projectId, project) : createProject(programId, project as ProjectInput),
    onSuccess: invalidate,
  })
}

export function useProjectAction(id: string) {
  const invalidate = useInvalidateProjects()
  return useMutation({ mutationFn: (action: ProjectAction) => applyProjectAction(id, action), onSuccess: invalidate })
}

export function useReplacePhases(projectId: string) {
  const invalidate = useInvalidateProjects()
  return useMutation({ mutationFn: (phases: PhaseInput[]) => replacePhases(projectId, phases), onSuccess: invalidate })
}

export function useUpdatePhase() {
  const invalidate = useInvalidateProjects()
  return useMutation({
    mutationFn: ({ phaseId, changes }: { phaseId: string; changes: PhaseChanges }) => updatePhase(phaseId, changes),
    onSuccess: invalidate,
  })
}

export function useReplaceDemands(projectId: string) {
  const invalidate = useInvalidateProjects()
  return useMutation({ mutationFn: (demands: DemandInput[]) => replaceDemands(projectId, demands), onSuccess: invalidate })
}
