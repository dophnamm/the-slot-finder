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
  'Add absence': string
  'Add at least one skill.': string
  'Add holiday': string
  'Add person': string
  'Add project': string
  'Add skill': string
  Approved: string
  'Approved absence always lowers capacity; planned absence does when the setting is on.': string
  'At risk': string
  'Back to dashboard': string
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
  'Could not finish signing in.': string
  'Could not reach the sign-in service.': string
  'Current phase': string
  'Customer engagements and their projects.': string
  Dashboard: string
  Dates: string
  Deadline: string
  Default: string
  'Demand, timeline, assigned people, risks and scope points.': string
  Done: string
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
  From: string
  'From {date}': string
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
  Lead: string
  Leave: string
  Level: string
  'Level {level}': string
  'Loading…': string
  'Mark done': string
  Members: string
  'Members appear here once a staffing draft for this program is accepted.': string
  'My absences': string
  'My schedule': string
  Name: string
  'New program': string
  Next: string
  'No absences yet.': string
  'No access': string
  'No goal yet.': string
  'No holidays yet.': string
  'No lead': string
  'No people yet.': string
  'No person record': string
  'No phases yet': string
  'No programs yet.': string
  'No projects yet.': string
  'Non-project hours': string
  'Objectives, one per line': string
  'On hold': string
  Open: string
  'Open the charter': string
  'Ordered by priority, then deadline.': string
  'Page not found': string
  'Page {page} of {pages}': string
  People: string
  Person: string
  'Person saved.': string
  'Person × week heatmap, with overload and deadline-risk alerts.': string
  Planned: string
  Planner: string
  Previous: string
  Primary: string
  Priority: string
  'Priority {priority}': string
  Program: string
  'Program brief': string
  'Program not found': string
  'Program saved.': string
  'Program updated.': string
  'Program {id}': string
  Programs: string
  Project: string
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
  'Staffing draft {id}': string
  'Start closing': string
  'Start date': string
  'Start – deadline': string
  Status: string
  'The server is unreachable.': string
  'The Slot Finder': string
  'There is no program at this address.': string
  'There is no screen at this address.': string
  'This cannot be undone. Open projects of a cancelled program are cancelled too. Continue?': string
  Timeline: string
  To: string
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
  '{client} · lead: {lead} · {timeline}': string
  '{expertise} · level {level}': string
  '{hours} h non-project': string
  '{hours} h per day': string
  '{hours} h per week': string
  '{start} – {end}': string
}
