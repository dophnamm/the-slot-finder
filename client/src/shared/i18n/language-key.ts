/**
 * Every piece of user-facing text in the client.
 *
 * The key is the English text itself, so English needs no locale file: a missing
 * translation falls back to the key. Placeholders use {name} and are type-checked by t().
 * Keep the keys sorted alphabetically (case-insensitive). Only the keys matter; the
 * value type is a placeholder.
 */
export type LanguageKey = {
  '+{expertise} {hours}': string
  'A change produced a newer version of this draft.': string
  'A change produced a staffing draft for this program.': string
  'A customer engagement that groups projects under one timeline.': string
  'A later end week moves every later phase by the same number of weeks.': string
  'A new project starts as incoming until a staffing draft is accepted.': string
  'A project': string
  'Absence added.': string
  Absences: string
  'Absences of {name}': string
  'Accept draft': string
  'Accept this staffing draft? It becomes the baseline everyone plans against.': string
  Accepted: string
  'Accepted as baseline version {version}.': string
  Active: string
  'Add absence': string
  'Add assumption': string
  'Add at least one expertise.': string
  'Add demand': string
  'Add expertise': string
  'Add holiday': string
  'Add leave': string
  'Add person': string
  'Add phase': string
  'Add project': string
  Added: string
  After: string
  'All expertises': string
  Approved: string
  'Approved absence always lowers capacity; planned absence does when the setting is on.': string
  Assigned: string
  Assignments: string
  Assumptions: string
  'At risk': string
  'Back to dashboard': string
  Baseline: string
  'Baseline load against capacity per person and ISO week. Click a cell to see its assignments.': string
  Before: string
  'Before the change': string
  'Behind on delivery': string
  Budget: string
  'Budget hours': string
  'Budget hours each demand needs and the hours this draft assigns.': string
  'Budget hours from the demand against hours planned in the baseline.': string
  'Budget hours per expertise and minimum level. Without a phase, a demand spans the whole project.': string
  'Build scenarios (move a project, add leave) and view the diff.': string
  'Built on baseline version {base}; version {current} is in force now. Run it again to see today’s effect.': string
  'Built on baseline version {version}.': string
  Cancel: string
  'Cancel program': string
  'Cancel project': string
  Cancelled: string
  Change: string
  'Change history': string
  Changed: string
  'Changes against the baseline': string
  'Charter link': string
  'Choose a person': string
  'Choose a project': string
  Client: string
  Close: string
  Closed: string
  Closing: string
  'Company holiday {dates}': string
  'Company holiday {dates} removed': string
  'Company holidays': string
  'Computing…': string
  'Could not finish signing in.': string
  'Could not reach the sign-in service.': string
  Coverage: string
  Covered: string
  Covers: string
  'Create scenario': string
  'Cross-expertise': string
  'Current phase': string
  'Customer engagements and their projects.': string
  Dashboard: string
  Dates: string
  Deadline: string
  'Deadline week': string
  'Deadline {week}': string
  Default: string
  Delete: string
  'Delete this scenario? The real plan is not affected.': string
  Demand: string
  'Demand saved.': string
  Done: string
  'Done points': string
  Draft: string
  'Draft accepted. It is now the baseline.': string
  'Draft created': string
  'Draft rejected.': string
  'e.g. Checkout slips two weeks': string
  'Each assumption is applied to a copy of the baseline. Nothing real changes.': string
  'Each person and demand whose hours or weeks differ in this scenario.': string
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
  'Every demand is covered.': string
  'Exactly one expertise is primary: the one this person is staffed for first.': string
  Expertise: string
  'Expertise, levels and weekly hours of everyone who can be staffed.': string
  Failed: string
  Forecast: string
  From: string
  'From week': string
  'From {date}': string
  Gap: string
  'Generate a new draft': string
  'Go to your home page': string
  Goal: string
  'Goal: {goal}': string
  'Has stretch': string
  Headcount: string
  Healthy: string
  'Hide details': string
  Holiday: string
  'Holiday added.': string
  'Holidays lower everyone’s capacity on those days.': string
  Home: string
  Hours: string
  'Hours in {week}': string
  'Hours moved': string
  'Hours per day': string
  'Hours per ISO week. Locked hours stay as they are when a draft is generated again; stretch is always locked.': string
  'Hours the planning engine could not place in this scenario.': string
  Inactive: string
  Incoming: string
  'Incoming projects across programs, highest priority and earliest deadline first.': string
  Intake: string
  'Intake queue': string
  Lead: string
  Leave: string
  'Leave, holidays and slipped phases or deadlines. Each change is checked against the baseline; a conflict produces a staffing draft.': string
  Level: string
  'Level {level}': string
  'Loading…': string
  'Lock hours': string
  'Mark done': string
  Members: string
  'Members ({count})': string
  'Minimum level': string
  'Move a project': string
  Moved: string
  'Moved, not saved yet': string
  'My absences': string
  'My schedule': string
  Name: string
  'Name or expertise': string
  'Near full': string
  Needed: string
  'New phase': string
  'New program': string
  'New scenario': string
  Next: string
  'No absences yet.': string
  'No access': string
  'No assignments this week.': string
  'No changes against the baseline.': string
  'No changes yet.': string
  'No demand yet.': string
  'No free capacity in the window': string
  'No goal yet.': string
  'No holidays yet.': string
  'No impact': string
  'No incoming projects.': string
  'No lead': string
  'No people yet.': string
  'No person record': string
  'No phases yet': string
  'No primary expertise': string
  'No programs yet.': string
  'No project needs attention.': string
  'No projects yet.': string
  'No scenarios yet. Create one to try a change.': string
  'Nobody has hours on this program yet.': string
  'Nobody has the minimum level': string
  'Nobody has this expertise': string
  'Nobody is assigned yet.': string
  'Nobody is over 100 % in the coming weeks.': string
  'Nobody matches these filters.': string
  'Nobody’s hours change in this scenario.': string
  'Non-project hours': string
  'None this week': string
  'Not primary': string
  'Nothing real has changed. To make it real, move the phases on the project’s page or add the leave on the People page; a re-plan draft follows.': string
  'Objectives, one per line': string
  'On hold': string
  Open: string
  'Open draft ({status})': string
  'Open the charter': string
  'Open the latest version': string
  'Ordered by priority, then deadline.': string
  'Over-planned': string
  Overloaded: string
  'Overloaded people ({count})': string
  'Page not found': string
  'Page {page} of {pages}': string
  'Past the end date': string
  People: string
  'People changed': string
  'People with hours in the baseline on this program’s projects, most hours first.': string
  Person: string
  'Person saved.': string
  'Person to add': string
  Phase: string
  'Phase name': string
  'Phase saved.': string
  'Phase saved. {count} later phase moved.': string
  'Phase saved. {count} later phases moved.': string
  Phases: string
  'Phases saved.': string
  'Pick a date': string
  'Pick a week': string
  Planned: string
  'Planned again:': string
  'Planned burn': string
  'Planned leave this week, not approved yet.': string
  'Planned leave, not approved yet': string
  'Planned to date': string
  Planner: string
  Previous: string
  Primary: string
  'Primary expertise': string
  Priority: string
  'Priority {priority}': string
  Program: string
  'Program brief': string
  'Program not found': string
  'Program saved.': string
  'Program updated.': string
  'Program {id}': string
  Programs: string
  Progress: string
  Project: string
  'Project not found': string
  'Project saved.': string
  'Project updated.': string
  Projects: string
  'Projects ({count})': string
  'Projects that are understaffed, at risk or behind on delivery, and people over 100 % in the coming weeks.': string
  'Propose squad': string
  'Proposing…': string
  'Put on hold': string
  'Re-plan': string
  Reason: string
  Reject: string
  'Reject this staffing draft?': string
  Rejected: string
  'Remove absence': string
  'Remove assumption': string
  'Remove demand': string
  'Remove expertise': string
  'Remove from the draft': string
  'Remove holiday': string
  'Remove phase': string
  'Remove the phase {name}?': string
  Removed: string
  'Report leave or training. A planner approves it.': string
  Reset: string
  Resume: string
  'Review draft': string
  Risks: string
  Role: string
  'Run again': string
  Save: string
  'Save demand': string
  'Saving moves every later phase by 1 week.': string
  'Saving moves every later phase by {weeks} weeks.': string
  Scenario: string
  Scenarios: string
  Scope: string
  'Scope points': string
  Search: string
  'Short {hours} h': string
  'Show details': string
  'Sign in': string
  'Sign in again': string
  'Sign out': string
  'Signed in as {name}': string
  'Signed out': string
  'Signing you in…': string
  Someone: string
  Specialist: string
  Staffing: string
  'Staffing by expertise': string
  'Staffing draft': string
  'Staffing draft not found': string
  Stale: string
  'Start closing': string
  'Start date': string
  'Start week': string
  'Start – deadline': string
  Status: string
  Stretch: string
  'Summed over {count} project that is not cancelled.': string
  'Summed over {count} projects that are not cancelled.': string
  Superseded: string
  'Team heatmap': string
  'The baseline changed to version {version} since this draft was made, so it can no longer be accepted.': string
  'The free hours are held by locked allocations': string
  'The last phase ends after the deadline week. Move the deadline or shorten a phase.': string
  'The last phase would end after the deadline week.': string
  'The server is unreachable.': string
  'The server keeps rejecting your sign-in. Try again, or ask an admin to check the sign-in settings.': string
  'The Slot Finder': string
  'There is no program at this address.': string
  'There is no project at this address.': string
  'There is no screen at this address.': string
  'There is no staffing draft at this address.': string
  'This cannot be undone. Continue?': string
  'This cannot be undone. Open projects of a cancelled program are cancelled too. Continue?': string
  'This draft has stretch assignments: people below the minimum level or without the expertise. Accept them anyway?': string
  Time: string
  Timeline: string
  To: string
  Training: string
  'Triggered by': string
  'Try a change, such as a project slipping or someone on leave, and see what it does to the plan. Nothing real changes.': string
  'Try again': string
  Type: string
  Unavailable: string
  Understaffed: string
  Underused: string
  'Unexpected error.': string
  'Unlock hours': string
  Unmet: string
  'Unmet demand': string
  'Until {date}': string
  'Version {version}': string
  Waiting: string
  'Weekly hours': string
  Weeks: string
  'Weeks (minus is earlier)': string
  'What-if scenarios': string
  When: string
  'Who gains or loses hours if this draft is accepted.': string
  'Whole project': string
  'You have signed out.': string
  'Your account is not linked to a person yet. Ask a planner to add you with the same email.': string
  'Your assignments by week appear here once a staffing draft that includes you is accepted.': string
  'Your role cannot open this page.': string
  'Your session has ended. Sign in again.': string
  '{client} · lead: {lead} · {timeline}': string
  '{count} attempts: {error}': string
  '{count} change is waiting for review.': string
  '{count} changes are waiting for review.': string
  '{count} weeks': string
  '{date} ({week})': string
  '{done} of {scope} points done': string
  '{expertise} (level {level}+)': string
  '{expertise} ({count})': string
  '{expertise} · level {level}': string
  '{expertise} · {hours} h': string
  '{expertise} · {program}': string
  '{gap} h short of {budget} h': string
  '{hours} h': string
  '{hours} h non-project': string
  '{hours} h per day': string
  '{hours} h per week': string
  '{load} of {capacity} h · {band}': string
  '{load}/{capacity} h': string
  '{name} (stretch)': string
  '{percent}%': string
  '{person} away {dates}, {hours} h a day': string
  '{person} in {week}': string
  '{person} is away {dates}': string
  '{person} is no longer away {dates}': string
  '{project} ({program})': string
  '{project} 1 week earlier': string
  '{project} 1 week later': string
  '{project} {weeks} weeks earlier': string
  '{project} {weeks} weeks later': string
  '{project} · {expertise} · {phase}': string
  '{project}: deadline moved from {from} to {to}': string
  '{project}: {phase} now ends {week}, {weeks} weeks later': string
  '{start} – {end}': string
  '{start} – {end} · level {level}+ · {assigned} of {hours} h': string
  '{week} ({dates})': string
  '{week} · {range}': string
  '{week}: {load}/{capacity} h': string
  '—': string
}
