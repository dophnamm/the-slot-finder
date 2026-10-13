const person = (id: string) => `/people/${encodeURIComponent(id)}`

export const PEOPLE_ENDPOINTS = {
  list: '/people',
  detail: person,
  skills: (id: string) => `${person(id)}/skills`,
  absences: (id: string) => `${person(id)}/absences`,
  assignments: (id: string) => `${person(id)}/assignments`,
  schedule: (id: string) => `${person(id)}/schedule`,
} as const

export const EXPERTISE_ENDPOINTS = {
  list: '/expertises',
} as const
