const program = (id: string) => `/programs/${encodeURIComponent(id)}`
const project = (id: string) => `/projects/${encodeURIComponent(id)}`

export const PROGRAM_ENDPOINTS = {
  list: '/programs',
  detail: program,
  action: (id: string, action: string) => `${program(id)}/actions/${encodeURIComponent(action)}`,
  overview: (id: string) => `${program(id)}/overview`,
  projects: (id: string) => `${program(id)}/projects`,
  members: (id: string) => `${program(id)}/members`,
  team: (id: string) => `${program(id)}/team`,
  teamMember: (id: string, personId: string) => `${program(id)}/team/${encodeURIComponent(personId)}`,
  assistants: (id: string) => `${program(id)}/assistants`,
  assistant: (id: string, assistantId: string) => `${program(id)}/assistants/${encodeURIComponent(assistantId)}`,
  milestones: (id: string) => `${program(id)}/milestones`,
  forecast: (id: string) => `${program(id)}/forecast`,
} as const

export const PROJECT_ENDPOINTS = {
  list: '/projects',
  detail: project,
  action: (id: string, action: string) => `${project(id)}/actions/${encodeURIComponent(action)}`,
  phases: (id: string) => `${project(id)}/phases`,
  demands: (id: string) => `${project(id)}/demands`,
} as const

export const PHASE_ENDPOINTS = {
  detail: (id: string) => `/phases/${encodeURIComponent(id)}`,
} as const
