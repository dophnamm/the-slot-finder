import { PEOPLE_ENDPOINTS, PHASE_ENDPOINTS, PROGRAM_ENDPOINTS, PROJECT_ENDPOINTS } from '@/resources/endpoints'
import { request, requestWithMeta } from '@/shared/api/http'
import { LEAD_OPTIONS_PAGE_SIZE } from './constants'
import type {
  DemandInput,
  Phase,
  PhaseChanges,
  PhaseInput,
  Program,
  ProgramAction,
  ProgramInput,
  Project,
  ProjectAction,
  ProjectDetail,
  ProjectInput,
  ProjectStatus,
} from './types'

export function listPrograms(page: number, pageSize: number) {
  return requestWithMeta<Program[]>({ url: PROGRAM_ENDPOINTS.list, params: { page, page_size: pageSize } })
}

export function getProgram(id: string) {
  return request<Program>({ url: PROGRAM_ENDPOINTS.detail(id) })
}

export function createProgram(program: ProgramInput) {
  return request<Program>({ url: PROGRAM_ENDPOINTS.list, method: 'POST', data: program })
}

export function updateProgram(id: string, changes: Partial<ProgramInput>) {
  return request<Program>({ url: PROGRAM_ENDPOINTS.detail(id), method: 'PATCH', data: changes })
}

export function applyProgramAction(id: string, action: ProgramAction) {
  return request<Program>({ url: PROGRAM_ENDPOINTS.action(id, action), method: 'POST' })
}

export function listProgramProjects(programId: string) {
  return request<Project[]>({ url: PROGRAM_ENDPOINTS.projects(programId) })
}

export function listProjects(status: ProjectStatus | undefined, page: number, pageSize: number) {
  return requestWithMeta<Project[]>({ url: PROJECT_ENDPOINTS.list, params: { status, page, page_size: pageSize } })
}

export function getProject(id: string) {
  return request<ProjectDetail>({ url: PROJECT_ENDPOINTS.detail(id) })
}

export function createProject(programId: string, project: ProjectInput) {
  return request<ProjectDetail>({ url: PROGRAM_ENDPOINTS.projects(programId), method: 'POST', data: project })
}

export function updateProject(id: string, changes: Partial<ProjectInput>) {
  return request<ProjectDetail>({ url: PROJECT_ENDPOINTS.detail(id), method: 'PATCH', data: changes })
}

export function applyProjectAction(id: string, action: ProjectAction) {
  return request<ProjectDetail>({ url: PROJECT_ENDPOINTS.action(id, action), method: 'POST' })
}

export function replacePhases(projectId: string, phases: PhaseInput[]) {
  return request<ProjectDetail>({ url: PROJECT_ENDPOINTS.phases(projectId), method: 'PUT', data: { phases } })
}

/** Edits one phase; a later end moves the later phases on the server. Returns the changed phases. */
export function updatePhase(phaseId: string, changes: PhaseChanges) {
  return request<Phase[]>({ url: PHASE_ENDPOINTS.detail(phaseId), method: 'PATCH', data: changes })
}

export function replaceDemands(projectId: string, demands: DemandInput[]) {
  return request<ProjectDetail>({ url: PROJECT_ENDPOINTS.demands(projectId), method: 'PUT', data: { demands } })
}

/** People who can be chosen as a lead (planners only can list people for this). */
export function listLeadCandidates() {
  return request<{ id: string; name: string; role: string }[]>({
    url: PEOPLE_ENDPOINTS.list,
    params: { page: 1, page_size: LEAD_OPTIONS_PAGE_SIZE },
  })
}
