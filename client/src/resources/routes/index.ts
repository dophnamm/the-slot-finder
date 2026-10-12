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
  mySchedule: '/me',
  events: '/events',
  notFound: '*',
} as const
