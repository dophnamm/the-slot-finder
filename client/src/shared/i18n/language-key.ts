/**
 * Every piece of user-facing text in the client.
 *
 * The key is the English text itself, so English needs no locale file: a missing
 * translation falls back to the key. Placeholders use {name} and are type-checked by t().
 * Keep the keys sorted alphabetically (case-insensitive). Only the keys matter; the
 * value type is a placeholder.
 */
export type LanguageKey = {
  'A customer engagement that groups projects under one timeline.': string
  'A new project starts as incoming until a staffing draft is accepted.': string
  'Absence added.': string
  Absences: string
  'Absences of {name}': string
  Active: string
  'Actual hours arrive in phase 2.': string
  'Add absence': string
  'Add at least one skill.': string
  'Add holiday': string
  'Add person': string
  'Add project': string
  'Add skill': string
  Approved: string
  'Approved absence always lowers capacity; planned absence does when the setting is on.': string
  Assistant: string
  'Back to dashboard': string
  Budget: string
  'Budget, planned and forecast hours per expertise.': string
  'Build scenarios (move a project, add leave) and view the diff.': string
  'Cancel program': string
  'Cancel project': string
  Cancelled: string
  'Change history': string
  'Change history and re-plan status.': string
  'Charter link': string
  Client: string
  Close: string
  Closed: string
  Closing: string
  'Company holidays': string
  'Contribution %': string
  Cost: string
  'Cost (hours)': string
  'Could not finish signing in.': string
  'Could not reach the sign-in service.': string
  'Customer engagements and their projects.': string
  Dashboard: string
  Dates: string
  Deadline: string
  Default: string
  'Demand, timeline, assigned people, risks and scope points.': string
  Description: string
  Done: string
  'Done %': string
  'Done points': string
  'Earliest start': string
  Edit: string
  'Edit person': string
  'Edit program': string
  'Edit project': string
  Email: string
  'End date': string
  Events: string
  'Exactly one skill is primary: the expertise this person is staffed for first.': string
  Expertise: string
  'Expertise, levels and weekly hours of everyone who can be staffed.': string
  External: string
  Forecast: string
  From: string
  'From {date}': string
  Gap: string
  'Go to your home page': string
  Goal: string
  'Goal: {goal}': string
  Holiday: string
  'Holiday added.': string
  'Holidays lower everyone’s capacity on those days.': string
  Hours: string
  'Hours per day': string
  Inactive: string
  Incoming: string
  'Incoming projects across programs, with “Propose squad”.': string
  Intake: string
  'Intake queue': string
  Internal: string
  Lead: string
  Leave: string
  Level: string
  'Level {level}': string
  'Loading…': string
  'Mark done': string
  Member: string
  Members: string
  'My absences': string
  'My schedule': string
  Name: string
  'New program': string
  Next: string
  'Next milestone: {name}, {week}': string
  'No absences yet.': string
  'No access': string
  'No holidays yet.': string
  'No lead': string
  'No people yet.': string
  'No person record': string
  'No programs yet.': string
  'Non-project hours': string
  'Objectives, one per line': string
  'On hold': string
  Open: string
  'Page not found': string
  'Page {page} of {pages}': string
  People: string
  'Per person: squad role, expertise, hours here and overall utilization over the next 4 weeks.': string
  'Per project: squad, status, start–end and next milestone, Time / Scope / Cost bars, staffing % and “Propose squad”.': string
  Person: string
  'Person saved.': string
  'Person × week heatmap, with overload and deadline-risk alerts.': string
  Planned: string
  'Planned to date': string
  Planner: string
  Previous: string
  Primary: string
  Priority: string
  'Priority {priority}': string
  Program: string
  'Program brief': string
  'Program saved.': string
  'Program {id}': string
  Programs: string
  'Project saved.': string
  'Project {id}': string
  Projects: string
  'Put on hold': string
  'Remove absence': string
  'Remove holiday': string
  'Remove skill': string
  'Report leave or training. A planner approves it.': string
  Resume: string
  'Review the assignment grid, reasons and diff; edit, lock and accept.': string
  Role: string
  Save: string
  Scenarios: string
  Scope: string
  'Scope (points)': string
  'Scope points': string
  'Server: checking…': string
  'Server: OK': string
  'Server: unreachable': string
  'Sign in': string
  'Sign in to reach the dashboard or your own schedule.': string
  'Sign in with Keycloak': string
  'Sign out': string
  'Signed in as {name}': string
  'Signing you in…': string
  Skills: string
  Specialist: string
  Staffing: string
  'Staffing %': string
  'Staffing by expertise': string
  'Staffing draft {id}': string
  'Start closing': string
  'Start date': string
  Status: string
  'The server is unreachable.': string
  'The Slot Finder': string
  'There is no screen at this address.': string
  Time: string
  Timeline: string
  'Timeline: {start} – {end} ({dates})': string
  To: string
  Total: string
  Training: string
  'Try again': string
  Type: string
  'Unexpected error.': string
  'Until {date}': string
  'Weekly hours': string
  'What-if scenarios': string
  'Your account is not linked to a person yet. Ask a planner to add you with the same email.': string
  'Your assignments by week appear here once a staffing draft that includes you is accepted.': string
  'Your role cannot open this page.': string
  'Your session has ended. Sign in again.': string
  '{expertise} · level {level}': string
  '{hours} h non-project': string
  '{hours} h per day': string
  '{hours} h per week': string
}
