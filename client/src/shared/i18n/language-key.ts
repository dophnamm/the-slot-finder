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
  'A manager sees the people in the teams of the programs they manage.': string
  'A new project starts as incoming until a staffing draft is accepted.': string
  'A planner adds the people this program’s projects are staffed from.': string
  'A program groups a customer’s projects under one timeline.': string
  'A project': string
  'A public holiday calendar from Google or Outlook, such as Holidays in Vietnam. The link is not saved.': string
  'Absence added.': string
  Absences: string
  'Absences of {name}': string
  'Accept draft': string
  'Accept this staffing draft? It becomes the baseline everyone plans against.': string
  Accepted: string
  'Accepted as baseline version {version}.': string
  Actions: string
  Active: string
  Actual: string
  'Add absence': string
  'Add assumption': string
  'Add at least one expertise.': string
  'Add demand': string
  'Add expertise': string
  'Add holiday': string
  'Add leave': string
  'Add member': string
  'Add person': string
  'Add phase': string
  'Add project': string
  'Add the hours each expertise needs, in a phase or across the whole project.': string
  'Add the people this program’s projects are staffed from.': string
  'Add the people who can be staffed, with their expertise and levels.': string
  'Add the program’s projects with their phases and demand.': string
  Added: string
  After: string
  'All expertises': string
  Approve: string
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
  Burn: string
  Calendar: string
  'Calendar connected.': string
  'Calendar link': string
  'Calendar sync': string
  'Calendar sync is turned off on this server.': string
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
  Column: string
  'Company holiday {dates}': string
  'Company holiday {dates} removed': string
  'Company holidays': string
  'Connect calendar': string
  'Connected to': string
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
  'Demand removed.': string
  'Demand saved.': string
  Disconnect: string
  'Disconnect this calendar? Leave it already added stays.': string
  Done: string
  'Done points': string
  'Download template': string
  'Download the template': string
  'Download the template for past weeks, fill in actual_hours, then import it. One row with a problem saves nothing.': string
  Draft: string
  'Draft accepted. It is now the baseline.': string
  'Draft created': string
  'Draft rejected.': string
  'Drafts for this program’s projects staff only people in its team.': string
  'Drag a row to change the order: the top project has the highest priority.': string
  'Drag to change the order, or use the arrow keys.': string
  'e.g. 0': string
  'e.g. 120': string
  'e.g. 2': string
  'e.g. 320': string
  'e.g. 40': string
  'e.g. 8': string
  'e.g. Acme Commerce Platform': string
  'e.g. Acme Retail': string
  'e.g. Alex Nguyen': string
  'e.g. alex@company.com': string
  'e.g. Checkout Revamp': string
  'e.g. Checkout slips two weeks': string
  'e.g. Discovery': string
  'e.g. https://calendar.google.com/calendar/ical/…/private-…/basic.ics': string
  'e.g. https://calendar.google.com/calendar/ical/…/public/basic.ics': string
  'e.g. https://docs.example.com/charter': string
  'e.g. Move the online store to the new platform before the holidays.': string
  'e.g. National Day': string
  'Each assumption is applied to a copy of the baseline. Nothing real changes.': string
  'Each event of a holiday calendar between the two dates becomes a company holiday. Holidays imported before are skipped.': string
  'Each person and demand whose hours or weeks differ in this scenario.': string
  'Earliest start': string
  Edit: string
  'Edit demand': string
  'Edit person': string
  'Edit phase': string
  'Edit program': string
  'Edit project': string
  Email: string
  'End date': string
  'End week': string
  Events: string
  'Every demand is covered.': string
  'Everyone is already in this team.': string
  'Exactly one expertise is primary: the one this person is staffed for first.': string
  Expertise: string
  'Expertise, levels and weekly hours of everyone who can be staffed.': string
  'Export a holiday calendar from Google or Outlook as an .ics file.': string
  Failed: string
  File: string
  Forecast: string
  From: string
  'From calendar': string
  'From week': string
  'From {date}': string
  Gap: string
  'Generate a new draft': string
  'Go to your home page': string
  Goal: string
  'Goal: {goal}': string
  'Google: Settings and sharing → Integrate calendar → Secret address in iCal format.': string
  'Has stretch': string
  Headcount: string
  Healthy: string
  'Hide details': string
  Holiday: string
  'Holiday added.': string
  Holidays: string
  'Holidays imported.': string
  'Holidays lower everyone’s capacity on those days.': string
  Home: string
  Hours: string
  'Hours by week against capacity, with leave and holidays by day.': string
  'Hours in this program': string
  'Hours in {week}': string
  'Hours moved': string
  'Hours must be a number from 0 to the weekly limit.': string
  'Hours per day': string
  'Hours per ISO week. Locked hours stay as they are when a draft is generated again; stretch is always locked.': string
  'Hours the planning engine could not place in this scenario.': string
  'ICS file': string
  Import: string
  'Import holidays': string
  'Import the filled-in file': string
  'Import timesheet': string
  Inactive: string
  Incoming: string
  'Incoming projects across programs, highest priority and earliest deadline first.': string
  Intake: string
  'Intake queue': string
  'Last sync': string
  Lead: string
  Leave: string
  'Leave, holidays and moved phases or deadlines appear here.': string
  'Leave, holidays and slipped phases or deadlines. Each change is checked against the baseline; a conflict produces a staffing draft.': string
  Level: string
  'Level {level}': string
  Line: string
  Link: string
  Load: string
  'Loading…': string
  'Lock hours': string
  Manager: string
  'Mark done': string
  'Member added.': string
  'Member removed.': string
  Members: string
  'Members ({count})': string
  'Minimum level': string
  'More actions': string
  'Move a project': string
  'Move {name}': string
  Moved: string
  'Moved, not saved yet': string
  'My absences': string
  'My schedule': string
  Name: string
  'Name or expertise': string
  Navigation: string
  'Near full': string
  Needed: string
  'New phase': string
  'New program': string
  'New scenario': string
  Next: string
  'Next 3 weeks': string
  'Next month': string
  'Next week': string
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
  'No manager': string
  'No one in the program team': string
  'No people yet.': string
  'No person has this email.': string
  'No person record': string
  'No phases yet': string
  'No primary expertise': string
  'No programs yet.': string
  'No project has this id.': string
  'No project needs attention.': string
  'No projects yet.': string
  'No scenarios yet. Create one to try a change.': string
  'Nobody has hours on this program yet.': string
  'Nobody has the minimum level': string
  'Nobody has this expertise': string
  'Nobody is assigned yet.': string
  'Nobody is in this team yet.': string
  'Nobody is over 100 % in the coming weeks.': string
  'Nobody matches these filters.': string
  'Nobody’s hours change in this scenario.': string
  'Non-project hours': string
  'None this week': string
  'Not an ISO week such as 2026-W41.': string
  'Not primary': string
  'Not staffed yet': string
  'Not yet': string
  'Nothing real has changed. To make it real, move the phases on the project’s page or add the leave on the People page; a re-plan draft follows.': string
  'Nothing was saved. Fix these rows and import again.': string
  'Objectives, one per line': string
  'On hold': string
  'One objective per line, e.g. Ship the new checkout.': string
  'Only the program team is listed. People with the expertise at the minimum level come first; the others would be stretch.': string
  Open: string
  'Open draft ({status})': string
  'Open navigation': string
  'Open the charter': string
  'Open the latest version': string
  'Ordered by priority, then deadline.': string
  'Out-of-office events': string
  'Out-of-office events in a Google or Outlook calendar become planned leave for a planner to approve.': string
  'Outlook: Settings → Calendar → Shared calendars → Publish a calendar → ICS link.': string
  'Over-planned': string
  Overloaded: string
  'Overloaded people': string
  'Overloaded people ({count})': string
  Overview: string
  'Page not found': string
  'Page {page} of {pages}': string
  'Past the end date': string
  'Past weeks show worked hours when a timesheet was imported.': string
  People: string
  'People appear here once a staffing draft for this program is accepted.': string
  'People changed': string
  Person: string
  'Person saved.': string
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
  'Planned {planned} h · Worked {actual} h': string
  Planner: string
  Previous: string
  'Previous month': string
  Primary: string
  'Primary expertise': string
  'Priority {priority}': string
  Problem: string
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
  'Projects added to a program wait here until a staffing draft covers them.': string
  'Projects that are understaffed, at risk or behind on delivery, and people over 100 % in the coming weeks.': string
  'Propose squad': string
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
  'Remove this demand?': string
  'Remove {name} from the team': string
  'Remove {name} from the team? Their baseline hours stay; the next draft plans without them.': string
  Removed: string
  'Report leave or training. A planner approves it.': string
  Resume: string
  'Review draft': string
  Risks: string
  Role: string
  'Run again': string
  Save: string
  'Save demand': string
  'Save phase': string
  'Saving moves every later phase by 1 week.': string
  'Saving moves every later phase by {weeks} weeks.': string
  Scenario: string
  Scenarios: string
  Schedule: string
  Scope: string
  'Scope points': string
  Search: string
  'Secret iCal address': string
  'Short {hours} h': string
  'Show details': string
  'Sign in': string
  'Sign in again': string
  'Sign out': string
  'Signed in as {name}': string
  'Signed out': string
  'Signing you in…': string
  Skills: string
  Someone: string
  Specialist: string
  'Split the project into phases such as Discovery, Build and Test.': string
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
  'Sync now': string
  Team: string
  'Team heatmap': string
  'The baseline changed to version {version} since this draft was made, so it can no longer be accepted.': string
  'The calendar could not be fetched. Your organization may block secret calendar links.': string
  'The calendar is too large to read.': string
  'The file is not valid CSV from this line.': string
  'The free hours are held by locked allocations': string
  'The header lacks this column.': string
  'The last phase ends after the deadline week. Move the deadline or shorten a phase.': string
  'The last phase would end after the deadline week.': string
  'The link does not return a calendar. Your organization may block secret calendar links.': string
  'The people this program’s projects are staffed from. A draft proposes only them.': string
  'The same person, project and week appear more than once.': string
  'The server is unreachable.': string
  'The server keeps rejecting your sign-in. Try again, or ask an admin to check the sign-in settings.': string
  'The Slot Finder': string
  'The week is after the current week.': string
  'There is no program at this address.': string
  'There is no project at this address.': string
  'There is no screen at this address.': string
  'There is no staffing draft at this address.': string
  'This cannot be undone. Continue?': string
  'This cannot be undone. Open projects of a cancelled program are cancelled too. Continue?': string
  'This draft has stretch assignments: people below the minimum level or without the expertise. Accept them anyway?': string
  Time: string
  Timeline: string
  'Timesheet file': string
  To: string
  'To week': string
  Today: string
  Training: string
  'Triggered by': string
  'Try a change, such as a project slipping or someone on leave, and see what it does to the plan. Nothing real changes.': string
  'Try again': string
  'Try another name or expertise, or clear the search.': string
  Type: string
  Unavailable: string
  Understaffed: string
  Underused: string
  'Unexpected error.': string
  'Unlock hours': string
  Unmet: string
  'Unmet demand': string
  'Until {date}': string
  'Use the secret iCal address from Google Calendar or Outlook. It starts with https://.': string
  'Version {version}': string
  Waiting: string
  Week: string
  'Weekly hours': string
  Weeks: string
  'Weeks (minus is earlier)': string
  'What-if scenarios': string
  When: string
  'Who gains or loses hours if this draft is accepted.': string
  'Whole project': string
  Worked: string
  'Worked {hours} h': string
  'You can’t see this person': string
  'You have signed out.': string
  'Your account is not linked to a person yet. Ask a planner to add you with the same email.': string
  'Your hours by week, leave and holidays.': string
  'Your role cannot open this page.': string
  'Your session has ended. Sign in again.': string
  '{added} added, {skipped} already there.': string
  '{client} · manager: {manager} · {timeline}': string
  '{count} attempts: {error}': string
  '{count} change is waiting for review.': string
  '{count} changes are waiting for review.': string
  '{count} weeks': string
  '{created} rows added, {updated} updated, {skipped} skipped.': string
  '{date} ({week})': string
  '{done} of {scope} points done': string
  '{expertise} (level {level}+)': string
  '{expertise} ({count})': string
  '{expertise} · level {level}': string
  '{expertise} · {hours} h': string
  '{expertise} · {program}': string
  '{expertise}, {hours} h per week, {nonProject} h of them non-project': string
  '{gap} h short of {budget} h': string
  '{hours} h': string
  '{hours} h non-project': string
  '{hours} h per day': string
  '{hours} h per week': string
  '{load} of {capacity} h · {band}': string
  '{load}/{capacity} h': string
  '{name} (stretch)': string
  '{name} · {expertise}': string
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
  '{type} (planned)': string
  '{week} ({dates})': string
  '{week} · {range}': string
  '{week}: {load}/{capacity} h': string
  '—': string
}
