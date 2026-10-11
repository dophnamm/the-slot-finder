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
  'A later end week moves every later phase by the same number of weeks.': string
  'A new project starts as incoming until a staffing draft is accepted.': string
  'Absence added.': string
  Absences: string
  'Absences of {name}': string
  Active: string
  'Add absence': string
  'Add at least one skill.': string
  'Add demand': string
  'Add holiday': string
  'Add person': string
  'Add phase': string
  'Add project': string
  'Add skill': string
  Approved: string
  'Approved absence always lowers capacity; planned absence does when the setting is on.': string
  'At risk': string
  'Back to dashboard': string
  'Before the change': string
  'Budget hours': string
  'Budget hours per expertise and minimum level. Without a phase, a demand spans the whole project.': string
  'Build scenarios (move a project, add leave) and view the diff.': string
  Cancel: string
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
  'Deadline week': string
  Default: string
  Demand: string
  'Demand saved.': string
  Done: string
  'Done points': string
  'Earliest start': string
  Edit: string
  'Edit person': string
  'Edit phase': string
  'Edit program': string
  'Edit project': string
  Email: string
  'End date': string
  'End week': string
  Events: string
  'Exactly one skill is primary: the expertise this person is staffed for first.': string
  Expertise: string
  'Expertise, levels and weekly hours of everyone who can be staffed.': string
  From: string
  'From {date}': string
  'Go to your home page': string
  Goal: string
  'Goal: {goal}': string
  Headcount: string
  Holiday: string
  'Holiday added.': string
  'Holidays lower everyone’s capacity on those days.': string
  Hours: string
  'Hours per day': string
  Inactive: string
  Incoming: string
  'Incoming projects across programs, highest priority and earliest deadline first.': string
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
  'Minimum level': string
  'Moved, not saved yet': string
  'My absences': string
  'My schedule': string
  Name: string
  'New phase': string
  'New program': string
  Next: string
  'No absences yet.': string
  'No access': string
  'No demand yet.': string
  'No goal yet.': string
  'No holidays yet.': string
  'No incoming projects.': string
  'No lead': string
  'No people yet.': string
  'No person record': string
  'No phases yet': string
  'No programs yet.': string
  'No projects yet.': string
  'Non-project hours': string
  'None this week': string
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
  Phase: string
  'Phase name': string
  'Phase saved.': string
  'Phase saved. {count} later phase moved.': string
  'Phase saved. {count} later phases moved.': string
  Phases: string
  'Phases saved.': string
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
  'Project not found': string
  'Project saved.': string
  'Project updated.': string
  Projects: string
  'Put on hold': string
  'Remove absence': string
  'Remove demand': string
  'Remove holiday': string
  'Remove phase': string
  'Remove skill': string
  'Remove the phase {name}?': string
  'Report leave or training. A planner approves it.': string
  Reset: string
  Resume: string
  'Review the assignment grid, reasons and diff; edit, lock and accept.': string
  Role: string
  Save: string
  'Save demand': string
  'Saving moves every later phase by 1 week.': string
  'Saving moves every later phase by {weeks} weeks.': string
  Scenarios: string
  Scope: string
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
  'Start week': string
  'Start – deadline': string
  Status: string
  'The last phase ends after the deadline week. Move the deadline or shorten a phase.': string
  'The last phase would end after the deadline week.': string
  'The server is unreachable.': string
  'The Slot Finder': string
  'There is no program at this address.': string
  'There is no project at this address.': string
  'There is no screen at this address.': string
  'This cannot be undone. Continue?': string
  'This cannot be undone. Open projects of a cancelled program are cancelled too. Continue?': string
  Timeline: string
  To: string
  Training: string
  'Try again': string
  Type: string
  'Unexpected error.': string
  'Until {date}': string
  'Weekly hours': string
  Weeks: string
  'What-if scenarios': string
  'Whole project': string
  'Your account is not linked to a person yet. Ask a planner to add you with the same email.': string
  'Your assignments by week appear here once a staffing draft that includes you is accepted.': string
  'Your role cannot open this page.': string
  'Your session has ended. Sign in again.': string
  '{client} · lead: {lead} · {timeline}': string
  '{date} ({week})': string
  '{done} of {scope} points done': string
  '{expertise} · level {level}': string
  '{hours} h non-project': string
  '{hours} h per day': string
  '{hours} h per week': string
  '{start} – {end}': string
}
