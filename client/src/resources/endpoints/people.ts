const person = (id: string) => `/people/${encodeURIComponent(id)}`

export const PEOPLE_ENDPOINTS = {
  list: '/people',
  import: '/people/import',
  importTemplate: '/people/import-template',
  detail: person,
  skills: (id: string) => `${person(id)}/skills`,
  absences: (id: string) => `${person(id)}/absences`,
  assignments: (id: string) => `${person(id)}/assignments`,
  calendarFeed: (id: string) => `${person(id)}/calendar-feed`,
  calendarFeedSync: (id: string) => `${person(id)}/calendar-feed/sync`,
  schedule: (id: string) => `${person(id)}/schedule`,
} as const

export const EXPERTISE_ENDPOINTS = {
  list: '/expertises',
} as const
