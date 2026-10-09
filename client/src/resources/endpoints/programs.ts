const program = (id: string) => `/programs/${encodeURIComponent(id)}`

export const PROGRAM_ENDPOINTS = {
  list: '/programs',
  detail: program,
  overview: (id: string) => `${program(id)}/overview`,
  projects: (id: string) => `${program(id)}/projects`,
  members: (id: string) => `${program(id)}/members`,
  assistants: (id: string) => `${program(id)}/assistants`,
  assistant: (id: string, assistantId: string) =>
    `${program(id)}/assistants/${encodeURIComponent(assistantId)}`,
  milestones: (id: string) => `${program(id)}/milestones`,
  forecast: (id: string) => `${program(id)}/forecast`,
} as const
