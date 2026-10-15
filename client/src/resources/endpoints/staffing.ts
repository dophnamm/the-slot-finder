const draft = (id: string) => `/staffing-drafts/${encodeURIComponent(id)}`

export const STAFFING_ENDPOINTS = {
  drafts: '/staffing-drafts',
  open: '/plans/drafts',
  draft,
  assignments: (id: string) => `${draft(id)}/assignments`,
  accept: (id: string) => `${draft(id)}/accept`,
  reject: (id: string) => `${draft(id)}/reject`,
} as const
