export const ABSENCE_ENDPOINTS = {
  detail: (id: string) => `/absences/${encodeURIComponent(id)}`,
  approve: (id: string) => `/absences/${encodeURIComponent(id)}/approve`,
} as const

export const HOLIDAY_ENDPOINTS = {
  list: '/holidays',
} as const
