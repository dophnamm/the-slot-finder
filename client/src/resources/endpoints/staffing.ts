const draft = (id: string) => `/staffing-drafts/${encodeURIComponent(id)}`

export const STAFFING_ENDPOINTS = {
  drafts: '/staffing-drafts',
  open: '/plans/drafts',
  draft,
  assignments: (id: string) => `${draft(id)}/assignments`,
  alternatives: (id: string, assignmentId: string) =>
    `${draft(id)}/assignments/${encodeURIComponent(assignmentId)}/alternatives`,
  accept: (id: string) => `${draft(id)}/accept`,
  reject: (id: string) => `${draft(id)}/reject`,
  summary: (id: string) => `${draft(id)}/summary`,
} as const
