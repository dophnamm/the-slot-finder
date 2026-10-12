const scenario = (id: string) => `/scenarios/${encodeURIComponent(id)}`

export const SCENARIO_ENDPOINTS = {
  list: '/scenarios',
  detail: scenario,
  rerun: (id: string) => `${scenario(id)}/rerun`,
} as const
