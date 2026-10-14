/** Client routes. Build links to routes with params via generatePath(ROUTES.programDetail, { id }). */
export const ROUTES = {
  signedOut: '/signed-out',
  authCallback: '/auth/callback',
  dashboard: '/',
  programs: '/programs',
  programDetail: '/programs/:id',
  intake: '/intake',
  // Nested like the breadcrumbs, so the sidebar marks Programs on a project and Intake on a draft.
  projectDetail: '/programs/:programId/projects/:id',
  staffingDraft: '/intake/staffing-drafts/:id',
  scenarios: '/scenarios',
  scenarioDetail: '/scenarios/:id',
  people: '/people',
  // Nested under People, so the sidebar marks People on a person's page.
  personDetail: '/people/:id',
  mySchedule: '/me',
  events: '/events',
  notFound: '*',
} as const

/** The query parameter that opens one tab of a page, e.g. `/programs/:id?tab=projects`. */
export const TAB_PARAM = 'tab'

/** The program page's tabs; a link opens one with tabPath(). */
export const PROGRAM_TABS = {
  overview: 'overview',
  projects: 'projects',
  team: 'team',
} as const

/** A link to one tab of the page at `path`. */
export function tabPath(path: string, tab: string): string {
  return `${path}?${new URLSearchParams({ [TAB_PARAM]: tab })}`
}
