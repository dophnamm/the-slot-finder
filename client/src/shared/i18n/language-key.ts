/**
 * Every piece of user-facing text in the client.
 *
 * The key is the English text itself, so English needs no locale file: a missing
 * translation falls back to the key. Placeholders use {name} and are type-checked by t().
 * Keep the keys sorted alphabetically (case-insensitive). Only the keys matter; the
 * value type is a placeholder.
 */
export type LanguageKey = {
  'Actual hours arrive in phase 2.': string
  Assistant: string
  'Back to dashboard': string
  Budget: string
  'Budget, planned and forecast hours per expertise.': string
  'Build scenarios (move a project, add leave) and view the diff.': string
  'Change history': string
  'Change history and re-plan status.': string
  'Contribution %': string
  Cost: string
  'Cost (hours)': string
  'Could not finish signing in.': string
  'Could not reach the sign-in service.': string
  Dashboard: string
  'Demand, timeline, assigned people, risks and scope points.': string
  Description: string
  Done: string
  'Done %': string
  Email: string
  Events: string
  Expertise: string
  'Expertise and level matrix, leave calendar and capacity.': string
  External: string
  Forecast: string
  Gap: string
  'Go to your home page': string
  'Goal: {goal}': string
  'Incoming projects across programs, with “Propose squad”.': string
  Intake: string
  'Intake queue': string
  Internal: string
  Lead: string
  'List of programs with status, timeline and the four progress bars.': string
  Member: string
  Members: string
  'My schedule': string
  Name: string
  'Next milestone: {name}, {week}': string
  'No access': string
  Open: string
  'Page not found': string
  People: string
  'Per person: squad role, expertise, hours here and overall utilization over the next 4 weeks.': string
  'Per project: squad, status, start–end and next milestone, Time / Scope / Cost bars, staffing % and “Propose squad”.': string
  Person: string
  'Person × week heatmap, with overload and deadline-risk alerts.': string
  Planned: string
  'Planned to date': string
  Planner: string
  'Program brief': string
  'Program {id}': string
  Programs: string
  'Project {id}': string
  Projects: string
  'Review the assignment grid, reasons and diff; edit, lock and accept.': string
  Scenarios: string
  Scope: string
  'Scope (points)': string
  'Server: checking…': string
  'Server: OK': string
  'Server: unreachable': string
  'Sign in': string
  'Sign in to reach the dashboard or your own schedule.': string
  'Sign in with Keycloak': string
  'Sign out': string
  'Signed in as {name}': string
  'Signing you in…': string
  Specialist: string
  Staffing: string
  'Staffing %': string
  'Staffing by expertise': string
  'Staffing draft {id}': string
  'The server is unreachable.': string
  'The Slot Finder': string
  'There is no screen at this address.': string
  Time: string
  'Timeline: {start} – {end} ({dates})': string
  Total: string
  'Try again': string
  'Unexpected error.': string
  'What-if scenarios': string
  'Your projects by week, and a place to report leave.': string
  'Your role cannot open this page.': string
  'Your session has ended. Sign in again.': string
}
