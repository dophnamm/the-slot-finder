/** Client routes. Build links to routes with params via generatePath(ROUTES.programDetail, { id }). */
export const ROUTES = {
  signedOut: '/signed-out',
  authCallback: '/auth/callback',
  dashboard: '/',
  programs: '/programs',
  programDetail: '/programs/:id',
  intake: '/intake',
  projectDetail: '/projects/:id',
  staffingDraft: '/staffing-drafts/:id',
  scenarios: '/scenarios',
  people: '/people',
  mySchedule: '/me',
  events: '/events',
  notFound: '*',
} as const
