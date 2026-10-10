export const ABSENCE_ENDPOINTS = {
  detail: (id: string) => `/absences/${encodeURIComponent(id)}`,
} as const

export const HOLIDAY_ENDPOINTS = {
  list: '/holidays',
} as const
