# The Slot Finder: user manual

This manual is for the people who use the app: Planners, Managers and Specialists. It explains each screen, what you can do on it, and what the messages mean. For setup and technical detail, see the [root README](../README.md), [`server/README.md`](../server/README.md) and [`client/README.md`](../client/README.md).

The Slot Finder assigns customer projects to squads (PM, Analyst, Engineer, QA), shows who is overloaded or idle each ISO week (for example `2026-W42`), and proposes a new plan when a project slips or someone takes leave. Nothing changes the real plan until a staffing draft is accepted: by a Planner, or by a Manager for their own programs.

## Contents

1. [Sign in](#1-sign-in)
2. [Roles](#2-roles)
3. [Dashboard](#3-dashboard)
4. [People and skills](#4-people-and-skills)
5. [Programs and projects](#5-programs-and-projects)
6. [Staffing drafts](#6-staffing-drafts)
7. [What-if scenarios](#7-what-if-scenarios)
8. [Leave and absences](#8-leave-and-absences)
9. [Events and re-plan](#9-events-and-re-plan)
10. [Ask AI](#10-ask-ai)
11. [Guided tour](#11-guided-tour)
12. [Glossary](#12-glossary)
13. [Troubleshooting](#13-troubleshooting)

## 1. Sign in

Open the app's address. You are sent to the sign-in page; enter your account there and you return to the page you asked for. If you have no account yet, choose "Don't have an account?" under the sign-in button; it points you to the accounts below.

1. Open the app.
2. On the sign-in page, enter your account and sign in.
3. You land on the **Dashboard** (Planner, Manager) or on **My schedule** (Specialist).
4. To sign out, open the account menu and choose **Sign out**.

If you sign in but see "Your account is not linked to a person yet. Ask a planner to add you with the same email.", a Planner must add you to **People** with the same email address.

**Accounts for review.** Three accounts are already created, one per role. Sign in with each to see what that role sees (see [Roles](#2-roles)):

| Username | Password | Name | Role | Start here |
|---|---|---|---|---|
| `planner` | `P@ssw0rd123` | Pat Planner | **Planner** | **Dashboard**: the whole team, every program, drafts, scenarios and events |
| `manager` | `P@ssw0rd123` | Mia Manager | **Manager** | **Dashboard**: only the programs Mia manages (Acme Commerce Platform) |
| `specialist` | `P@ssw0rd123` | Sam Specialist | **Specialist** | **My schedule**: Sam's own hours, leave and holidays |

## 2. Roles

Your role decides which screens appear in the sidebar and which actions you can take. The server checks every action again.

| Role (as shown) | Sidebar screens | What you can see | What you can do |
|---|---|---|---|
| **Planner** | Dashboard, Programs, Intake, Drafts, Scenarios, People, Events | Everything: the whole team, all programs, all drafts, change history | Everything a Manager can, plus: program teams, approve leave, holidays, timesheet import, scenarios, and staffing drafts for any program. Uses Ask AI |
| **Manager** (role `manager`) | Dashboard, Programs, Intake, Drafts, People (read only) | Only the programs they lead or lead a project in, and the people in those programs' teams (and themselves) | Create programs; edit their programs and projects (phases, demand, done points, confidence); reorder projects; propose, edit, accept and reject drafts for their projects. Uses Ask AI |
| **Specialist** | My schedule | Their own schedule: hours per project, leave, holidays | Report their own leave; connect their own calendar |

Notes:
- A Manager cannot open a program, project or person outside their teams. The screen says "You can’t see this program", "You can’t see this project" or "You can’t see this person".
- Only a Planner can change a program's team, approve leave, or add holidays.
- Only a Planner sees **Scenarios** and **Events**.

## 3. Dashboard

The Dashboard (Planner and Manager) shows the team heatmap, the capacity forecast, and the risks. A Manager sees only the people in their program teams.

**Heatmap: switch the view.** The card header has two buttons:
- **By person**: one row per person, one column per ISO week. This is the default.
- **By expertise**: the capacity forecast, one row per expertise, one column per week (see below).

**Change the range and filter.**
1. Set the first week with **From week**. The **Previous weeks** and **Next weeks** buttons move the range by the number of weeks shown.
2. Search with **Name or expertise**.
3. Filter with **All expertises** or **All programs** (choosing a program shows its team).
4. Switching between **By person** and **By expertise** keeps the same range.

**How to read a cell (By person).** The colour is the person's utilization in that week: planned hours divided by capacity.

| Colour band | Utilization | Meaning |
|---|---|---|
| Underused | under 50 % | Spare capacity |
| Healthy | 50 % to 85 % | Normal load |
| Near full | 85 % to 100 % | Little room left |
| Overloaded | over 100 % | More planned than capacity |

A hatched cell is **tentative**: it contains leave that is planned but not approved yet. A past week with an imported timesheet shows the hours actually worked, and its tooltip shows the planned hours.

**Click a cell** to see the week's assignments for that person. The popover links to the project.

**How to read a cell (By expertise).** Each cell shows free hours for that expertise in that week (capacity minus load, per person, never below 0).
- Green: enough free hours.
- Amber, `Short N hrs`: the unmet demand of planned and active projects, plus the expected demand of incoming projects, is larger than the free hours.
- Red, `Overloaded`: load is above capacity.

The tooltip gives capacity, load, unmet and pipeline hours. **Click a cell** to see the people of that expertise in that week, highest load first, with their utilization and free hours. An expertise overloaded for several weeks in a row gets an **Overloaded N weeks in a row** badge.

**Risks card.** Lists, in this order:
- **Sustained overload** (people over 100 % for several weeks in a row, such as "Bao Tran: over 100 % for 3 weeks from 2026-W45"). This group is listed first and hidden when nobody has one.
- Projects that are understaffed, at risk or behind on delivery, and people over 100 % in the coming weeks.

Click a name to open the person or project. When nothing needs attention, the card says "No project needs attention."

## 4. People and skills

Open **People** from the sidebar. The screen has two tabs: **People** and **Holidays**. A Manager reads both tabs, without the buttons, and sees only the people in their program teams.

**People tab (Planner).**
1. Click **Add person**.
2. Enter the name, email, location, start and end dates, weekly hours and non-project hours.
3. Set the expertises: each person has one **Primary expertise** and any number of other expertises, each at a level from 1 to 5. Exactly one expertise must be primary.
4. Save.

Click a row to open the person's page (see [Leave and absences](#8-leave-and-absences)). Rows show each person's expertise and level matrix, start and end dates and location.

**Hours rules.**
- **Weekly hours** are the contracted hours per week.
- **Non-project hours** are kept for meetings, support and management. They must not exceed weekly hours.
- Capacity in a week is weekly hours, minus non-project hours, minus absence hours. A person has zero capacity before their start date and after their end date.

**Import timesheets (Planner).** Click **Import timesheet** on the People tab.
1. Click **Download template** for the past weeks you need. The file lists every person, project and week planned in the baseline.
2. Fill in the actual hours only. Keep the other columns.
3. Click **Import the filled-in file**.

One row with a problem saves nothing. The message gives the line and the problem (see [Troubleshooting](#13-troubleshooting)). Imported hours feed **Actual hours** and **Burn %** on programs and projects.

## 5. Programs and projects

**Programs** lists the programs you can see, with their status and timeline. The arrow on a row opens a preview with its figures, projects and members. A Manager with a single program opens it directly.

To create a program, click **New program**.

**Program page.** Click a program to open it. It has three tabs:
- **Overview**: progress bars (**Time**, **Scope**, **Planned burn**, **Staffing**), the cost figures, staffing by expertise, and the program brief.
- **Projects**: the projects in priority order, with the hours planned next week and in the next 3 weeks. Drag a row, or focus its handle and press the arrow keys, to change the order. The top project has the highest priority.
- **Team**: the people the program's projects are staffed from. Planners use **Add member** and **Remove**. Removing a member keeps their baseline hours.

Use **Edit** to change the program. The **⋯** menu holds **Put on hold**, **Start closing** and **Cancel program**. Cancelling a program also cancels its open projects.

A warning sign (⚠) on a progress bar means the app flags it: over 100 % time, behind on delivery, over-planned, or understaffed.

**Add a project.**
1. Open the program's **Projects** tab and click **Add project**.
2. Enter the name, earliest start, deadline and scope points. For an incoming project, set the **confidence** (0 to 100 %).
3. Save.

Both dates must lie within 156 weeks before or after the current week.

**Project page.** Open a project from the Projects tab. Here you edit the project form, set done points, and manage its phases and demand. **Propose squad** is in the header.

**Phases.** A project is split into ordered phases, such as Discovery, Build and Test, each with a start and end week.
1. Under **Phases**, click **Add phase** (or edit an existing phase).
2. Enter the name, start week and end week, then **Save phase**.

Moving a phase's end later moves every later phase of the same project by the same number of weeks. The dialog previews which phases will shift before you save. A phase you add or move must lie inside the project's start and deadline weeks. If a shift pushes the last phase past the deadline, the project is flagged at risk; the deadline itself does not move. A phase that still has demand cannot be removed.

**Demand.** Demand is the budget hours a project needs from one expertise.
1. Under **Demand**, click **Add demand**.
2. Choose the **Expertise**, the **Minimum level**, the **Headcount** (people working the demand in parallel), the **Budget hours** and the **Phase** (or **Whole project**).
3. **Save demand**.

Budget hours must be whole hours. Otherwise the form shows "Use whole hours" and keeps **Save** off.

**Project status.** A project is **Incoming** until a staffing draft covering it is accepted. Other statuses are **Planned**, **Active**, **On hold**, **Done** and **Cancelled**. The engine plans only projects in plannable statuses.

## 6. Staffing drafts

A staffing draft is a plan the engine proposes for one or more projects. It is not the real plan until a Planner accepts it.

**Generate a draft.**
1. Open **Intake** (the incoming projects, highest priority first) or a project page.
2. Click **Propose squad** on the project.
3. The draft opens for review.

**Review a draft.** Open **Drafts** from the sidebar, or follow the link from a project or from a re-plan alert. The list shows every open draft, newest first, with its origin (manual or re-plan), the projects and programs it covers, the baseline version it was made on (with **Stale** when the baseline has moved on), and its unmet demands. Click a row, or **Review draft**, to open it.

On the draft review screen:
- Each row is one person's assignment to one demand. Each week shows the person's weekly load, coloured with the same bands as the heatmap.
- **Reason**: why the engine chose the person, and the kind of assignment (primary or cross-expertise).
- **Changes against the baseline**: who gains or loses hours if the draft is accepted.
- **Unmet demand**: hours the engine could not place, with a reason (see the table in [Troubleshooting](#13-troubleshooting)).

**Edit a draft.**
- Type hours into a week cell.
- Drag a cell's hours to another week of the same row, or to the same week of another person on the same demand. Each drop saves once.
- **Add person** lists the program team only.
- **Swap person** moves a person's hours to someone from the program team who holds the expertise at the level. Locked hours must be unlocked first.
- **Allocate hours** sets the same hours every week, or a total spread evenly, over a range of weeks in the demand's window.
- **Lock hours** keeps the hours as they are when the draft is generated again. **Unlock hours** releases them.

A person below the demand's minimum level (or without the expertise) is a **stretch** assignment. Only a Planner can add one, by hand. It is always locked and flagged with ⚠, and accepting the draft needs an explicit confirmation.

**Accept or reject.**
1. Click **Accept draft** (or **Reject**).
2. Confirm in the dialog: "Accept this staffing draft?" or "Reject this staffing draft?".
3. On accept, the draft becomes the new baseline, "Draft accepted. It is now the baseline." The previous baseline is superseded. The program page of the draft's first project opens.

On reject, the draft closes and nothing changes: "Draft rejected."

**Stale draft.** If the baseline changed while you were reviewing, the draft can no longer be accepted. The message says "The baseline changed to version N since this draft was made, so it can no longer be accepted." Click **Generate a new draft** and review the new one.

## 7. What-if scenarios

Scenarios (Planner only) test a change without touching the baseline. Nothing real changes.

**Create a scenario.**
1. Open **Scenarios** and click **New scenario**.
2. Name it and add one or more assumptions with **Add assumption**: **Move a project** (its start, deadline and phases shift by a number of weeks) or **Add leave** (a person is away from one date to another, for some hours per day).
3. **Create scenario**.

**Read a scenario.** The scenario page shows its assumptions and figures, then:
- **Changes**, week by week: Moved, Added, Removed or Changed, with the person and project in each line.
- **Unmet demand**: hours the engine could not place in the scenario.
- Links to the projects planned again.

**Run again** recalculates the scenario against the current baseline, with the same assumptions. Use **Delete** to remove it; the app asks first.

A scenario that was built on an older baseline is marked **Stale**.

**Make it real.** A scenario does not change anything itself. To make a change real, move the phases on the project page or add the leave on the People page. A re-plan draft then follows (see [Events and re-plan](#9-events-and-re-plan)).

## 8. Leave and absences

Absences are leave, holidays and training. They lower capacity on the days they cover.

**Report leave (Specialist).** Open **My absences** on **My schedule** and report the dates and hours per day. The page says: "Report leave or training. A planner approves it." Your leave stays planned until a Planner approves it.

**Add or approve leave (Planner).**
1. Open **People**, click the person, and open the **Absences** tab.
2. Click **Add absence** to enter leave for them, or **Approve** on a planned absence.
3. To delete an absence, use **Remove absence**.

**Holidays (Planner).** Open **People**, then the **Holidays** tab.
- **Add holiday** adds one holiday with its name and date.
- **Import holidays** loads a public holiday calendar as an iCalendar file, or from a link (for example a Google public holiday calendar). Holidays imported before are skipped. A link is not saved.

Holidays lower everyone's capacity on those days, or only the people at the holiday's location.

**Hours rules.**
- Approved absence always lowers capacity.
- Planned absence lowers capacity too, when the organization's setting is on (it is on by default). The schedule marks it as tentative.
- A person's absences cannot overlap. The app answers "The person already has an absence on some of these dates."

**Calendar sync.** A Specialist or Planner can connect a person's calendar on the **Calendar sync** tab (or on **My schedule**):
1. Export the secret iCal address from Google Calendar or Outlook (the link starts with `https://`).
2. Paste it in **Connect calendar** and save.
3. Out-of-office events in the calendar become **planned** leave. A Planner approves them like any other leave.

Use **Sync now** to check the calendar at once, or **Disconnect** to stop. The link is a secret and is never shown again.

## 9. Events and re-plan

**Events** (Planner only) lists the change history: leave, holidays, and phases or deadlines that slipped. Each event shows its status:
- **Waiting**: the change has not been checked yet.
- **No impact**: the baseline still fits.
- **Draft created**: the change produced a re-plan draft.

A change that creates a conflict produces one re-plan draft. Several changes close together are merged into one open draft. The draft keeps the Planner's edits and locks, and an older version is marked superseded.

When a program has a re-plan draft, its page shows a **Re-plan** banner. Click it to open the draft, review it, and accept or reject it as in [Staffing drafts](#6-staffing-drafts).

## 10. Ask AI

**Ask AI** is a question panel for Planners and Managers. The round button sits at the bottom right of every screen.

1. Click **Open Ask AI**.
2. Type your question in **Your question** (up to 500 characters), for example "Who is overloaded next week?" and click **Send**.
3. The answer appears in the panel. A typing indicator shows while it is pending.

Ask AI answers from the data you are allowed to see, and it only reads. It cannot change a plan, accept a draft or create anything. It also knows the page you are on, so "this project" works on a project page. Messages stay on screen until you reload the page; there is no history.

If the answer fails, the panel shows an error reply. Click **Try again** to send the question again. At the usage limit it tells you how many seconds to wait.

## 11. Guided tour

The guided tour explains the sidebar. It starts once on your first visit, on a wide screen, with one step per screen you can open and one on the Ask AI button (Planner and Manager).

- Move with **Back** and **Next**. The last step ends with **Done**.
- Press **Esc** or the close button to end the tour.
- To start it again, open the account menu and choose **Take the tour**.

## 12. Glossary

| Term | Meaning |
|---|---|
| **Baseline** | The plan in force: the staffing everyone works to until a newer draft is accepted. |
| **Draft** (staffing draft) | A plan the engine proposes for some projects. Nothing changes until it is accepted. A **re-plan draft** comes from a change such as leave or a slipped phase. |
| **Allocation** | The hours one person spends on one demand in one ISO week, within a plan. |
| **Utilization** | Planned hours divided by a person's capacity in a week. Over 100 % is overloaded. |
| **Free hours** | Capacity minus load, per person and never below 0. The capacity forecast adds them up per expertise. |
| **Shortfall** | Demand that free hours cannot cover. Shown as `Short N hrs` in the capacity forecast. |
| **Sustained overload** | A person over 100 % for several weeks in a row. It is a burnout risk and is listed before single overloaded weeks. |
| **Expertise** | A kind of work: PM, Analyst, Engineer or QA. Each person has one primary expertise and may hold others at a level from 1 to 5. |
| **Cross-expertise** | An assignment in an expertise other than the person's primary one, at or above the level the demand needs. |
| **Stretch** | An assignment below the demand's minimum level, or without the expertise. Added by a Planner by hand, locked, and confirmed on accept. |
| **Phase** | An ordered stage of a project with its own week range, such as Discovery, Build or Test. |
| **Demand** | The budget hours one expertise needs at a minimum level, inside one phase or across the whole project. |
| **Budget hours** | The hours a demand is planned to use. |
| **Scenario** | A what-if plan built from changes you choose. It does not change the baseline. |
| **Stale** | A draft or scenario built on a baseline that has since changed. |
| **Unmet demand** | Hours the engine could not place, with the reason. |
| **Tentative** | Planned leave that is not approved yet. It is shown hatched. |
| **ISO week** | A week numbered like `2026-W42`, Monday to Sunday. |

## 13. Troubleshooting

| Message | What it means | What to do |
|---|---|---|
| "Your session has ended. Sign in again." | Your sign-in expired. | Sign in again. |
| "Your role cannot open this page." / "No access" | Your role has no access to the screen. | Ask a Planner if you need access. |
| "You can’t see this program / project / person" | A Manager opened something outside their programs or teams. | Ask a Planner to check the team. |
| "There is no … at this address." / "Page not found" | The link points to something that does not exist, or was removed. | Go back to the Dashboard and find it again. |
| "Too many requests. Try again in N seconds." | You sent too many requests in a short time. | Wait, then try again. |
| "The assistant is at its usage limit. Try again in N seconds." | Ask AI reached its limit. | Wait, then ask again. |
| "The baseline changed to version N since this draft was made, so it can no longer be accepted." | Another change replaced the baseline while you were reviewing. | Click **Generate a new draft** and review it. |
| "A change produced a newer version of this draft; open the latest one." | The draft was superseded by a re-plan. Your save did not go through. | Open the latest version and check your edits. |
| "The person already has an absence on some of these dates." | The new leave overlaps another absence of the same person. | Change the dates, or remove the older absence. |
| "non_project_hours must not exceed weekly_hours" | Non-project hours are larger than the weekly hours. | Lower non-project hours, or raise weekly hours. |
| "Use whole hours" | Budget hours must be whole numbers. | Enter whole hours. |
| "Project dates and phase weeks must lie within 156 weeks of the current week." | A date is too far from the current week. | Pick a date within 156 weeks before or after today. |
| "A phase must lie between the project's start_earliest and deadline weeks" | A new or moved phase is outside the project's start and deadline. | Move the phase inside the project, or move the project dates first. |
| "The last phase ends after the deadline week." | A warning: the project is at risk. | Move the deadline or shorten a phase. |
| "A phase that still has demand cannot be removed; move or delete its demand first." | The phase still holds demand. | Move or delete that demand, then remove the phase. |
| "Weeks before the current week cannot be changed." | Past weeks are fixed. | Change only the current and future weeks. |
| "The week is outside the demand's window." | The hours are outside the demand's phase or project weeks. | Use a week inside the window. |
| "The draft has stretch assignments; confirm them to accept." | Accepting needs the stretch assignments confirmed. | Confirm in the dialog, or remove the stretch assignments. |
| "Nobody has this expertise" / "Nobody has the minimum level" / "No one in the program team" | Unmet demand: nobody fits this demand. | Add a person to the program team, or change the demand. |
| "Free hours exist, but only in weeks under the minimum" / "…the weekly expertise limit blocks them" | Unmet demand: free hours exist, but the rules block them. | Adjust the hours or the people in the draft by hand. |
| "The free hours are held by locked allocations" | Unmet demand: the free hours are locked to other work. | Unlock hours, if that is right. |
| "The file is not valid CSV from this line." / "The header lacks this column." / "The same person, project and week appear more than once." | The timesheet file has a problem. Nothing was saved. | Fix the rows named in the message and import again. |
| "The calendar could not be fetched. Your organization may block secret calendar links." | The calendar link failed. | Check the link is the secret iCal address, or ask an administrator. |
| "Calendar sync is turned off on this server." | The administrator has not enabled calendar sync. | Ask an administrator. |
| "This screen hit an unexpected error." | Something went wrong in the app. | Reload the page. If it happens again, tell an administrator. |
