# Calendar sync

**Goal:** when someone blocks time as out of office in Google Calendar or Outlook, the Slot Finder picks it up as planned leave. The planner approves it, and the [re-plan flow](replan-on-change.md) runs as for any other leave, without anyone typing the leave twice.

**Roles:** the person, whatever their role (connects their own calendar), planner (connects anyone's, sees the status, syncs now).
**Target:** 15–16 Oct · **Depends on:** [member-calendar](member-calendar.md), [replan-on-change](replan-on-change.md) · **Branch:** `feat/calendar-sync` · **Status:** In review (PR #34)

## In scope

- **Connect.** A person pastes the **secret iCal (ICS) address** of their calendar on My schedule. A planner can do it on the person page. One calendar per person.
  - Saving checks the link (allowed host, reachable, really ICS) and runs the first sync.
  - Google: Settings and sharing → Integrate calendar → Secret address in iCal format. Outlook: Settings → Calendar → Shared calendars → Publish a calendar → ICS link.
- **Status.** The host, the last sync time, the last error and how many out-of-office events were found. The link itself is never sent back to the client.
- **Disconnect** removes the link. Absences it already created stay, as ordinary absences.
- **Sync.** A runner inside the server, like the re-plan runner, syncs each calendar every `CALENDAR_SYNC_INTERVAL`. **Sync now** runs one calendar at once.
- **Which events count.** Events from today to `CALENDAR_SYNC_WEEKS` weeks ahead that are not cancelled, and that:
  - are marked out of office (`X-MICROSOFT-CDO-BUSYSTATUS:OOF`, which Outlook writes), or
  - have a title containing one of `CALENDAR_LEAVE_KEYWORDS`, ignoring case
- **From event to absence.** Each event becomes one absence of type `leave`, status `planned`, source `calendar`, keyed by the event's `UID`.
  - An all-day event covers its days at the person's daily hours (weekly hours ÷ working weekdays).
  - A timed event on one day uses its length, capped at the daily hours. A timed event over several days counts full days.
  - Dates use the organization's time zone.
- **The absences follow the calendar.** On each sync, in one transaction per calendar:
  - a new event creates its absence
  - a moved or resized event updates its absence and sets it back to `planned`; it records `AbsenceRemoved` for the old dates and `AbsenceAdded` for the new ones
  - an event that is gone, or no longer counts, deletes its absence
  - each of these records a change event through the existing absence service, so re-planning works unchanged
  - absences entered by hand are never touched
- **UI marks.** The absence list and the member calendar mark these absences "From calendar".

## Not in this feature

- Signing in to Google or Microsoft (OAuth), the Google Calendar API or Microsoft Graph.
- Slack: it has no calendar API.
- Meeting hours replacing non-project hours, and busy or free blocks on the member calendar.
- Recurring out-of-office events (`RRULE`). They are skipped.
- Writing anything back to a calendar.

**As built**
- `calendar_feeds.next_sync_at`: the runner claims a feed by moving it one interval ahead, so no lock is held while it fetches.
- Leave hours are rounded to two decimals, as absences store them, so an unchanged event never looks moved.
- A moved event records `AbsenceRemoved` and `AbsenceAdded`. Approving goes through member-calendar's `POST /absences/{id}/approve`.
- ICS is read with `github.com/arran4/golang-ical` v0.3.6.
- A redirect to a host outside the list, such as the sign-in page an organization shows for a blocked link, is reported as `CALENDAR_FEED_UNREACHABLE`, whose message names the block.
- Leave running since before today whose event is shortened to end before today is shortened, not deleted. Past events are read only for that; they never add leave.
- New weekly hours change the hours of calendar leave and keep its status. Only new dates need approval again.
- Disconnecting turns the calendar's leave into ordinary absences (`source = manual`, no event UID), so a calendar connected later never removes it.
- The request log and Sentry keep `[redacted]` instead of the body of `PUT /people/{id}/calendar-feed` (`constants.SecretBodyPaths`), even when CSRF or auth refuses the request.
- The test fixtures are written from the Google and Outlook formats. A real link from the company tenant is still to be checked before the demo.
- Checked on the seeded stack, with the repo's own fixture served from `raw.githubusercontent.com` (a host allowed only in the local `.env`):
  - an `http` link and another host show the form's message, with no toast
  - connecting shows "Calendar connected." and planned leave marked "From calendar" on 21 Oct (4 h) and 26–28 Oct
  - approving it gives a re-plan draft after `REPLAN_DEBOUNCE`, and Sync now keeps it approved and leaves a hand-entered absence alone
  - a lead sees no panel and gets `403`, and a specialist gets `403` on someone else's calendar
  - no response and no log line holds the link

## Server

| Method | Path | Purpose | Roles |
|---|---|---|---|
| GET | `/people/{id}/calendar-feed` | The calendar's status, or `404` when none is connected | planner, the person themselves |
| PUT | `/people/{id}/calendar-feed` | Connect or replace: `{ url }`. Checks the link and syncs once | planner, the person themselves |
| DELETE | `/people/{id}/calendar-feed` | Disconnect | planner, the person themselves |
| POST | `/people/{id}/calendar-feed/sync` | Sync now | planner, the person themselves |

**Tables** (one Atlas migration: `make migrate-generate name=calendar_feeds`)
- `calendar_feeds`: `person_id` (unique), `url_ciphertext`, `host`, `last_synced_at`, `next_sync_at`, `last_error`, `events_found`, timestamps. The runner claims a calendar by moving its `next_sync_at` in a short transaction, so no lock is held while it downloads.
- `absences` gets `source` (`manual` by default, or `calendar`) and `external_uid` (unique per person when set).

**Security.** The server fetches a URL that a user typed, so:
- The link is encrypted at rest (AES-GCM, key `CALENDAR_FEED_KEY`). Without a key the feature is off and these routes answer `503` with `CALENDAR_SYNC_DISABLED`.
- The link must be `https`, and its host must be in `CALENDAR_ALLOWED_HOSTS` (default `calendar.google.com`, `outlook.office365.com`, `outlook.live.com`). A redirect may only go to an allowed host.
- The server refuses addresses that resolve to private, loopback or link-local IPs.
- Each fetch has a timeout (`CALENDAR_FETCH_TIMEOUT`) and a size cap (`CALENDAR_MAX_BYTES`).
- The link never appears in logs, errors or responses. Problems are stored as codes: `CALENDAR_FEED_UNREACHABLE`, `CALENDAR_FEED_NOT_ICS`, `CALENDAR_FEED_TOO_LARGE`, `CALENDAR_FEED_HOST_NOT_ALLOWED`.

**Runner**
- Started and stopped by an FX lifecycle hook. Every `CALENDAR_SYNC_INTERVAL` it claims up to `CALENDAR_SYNC_BATCH_SIZE` calendars due for a sync, with `FOR UPDATE SKIP LOCKED`, so two servers never sync the same one.
- A failed fetch keeps the absences as they are and stores the error code. The next run tries again.
- The fetcher is an interface. Tests use a fake one, and no test reaches the internet.

**Settings** (in `cmd/server/.env.example` and [server/README §8](../../server/README.md#8-settings)): `CALENDAR_FEED_KEY`, `CALENDAR_ALLOWED_HOSTS`, `CALENDAR_SYNC_INTERVAL` (15m), `CALENDAR_SYNC_BATCH_SIZE` (20), `CALENDAR_SYNC_WEEKS` (26), `CALENDAR_LEAVE_KEYWORDS` (`OOO,Out of office,Vacation,Annual leave,Nghỉ phép`), `CALENDAR_FETCH_TIMEOUT` (10s), `CALENDAR_MAX_BYTES` (5 MB).

## Client

- `features/calendar-feed`: the connect form (paste the link, with where to find it in Google and Outlook), the status (host, last sync, error, events found), **Sync now** and **Disconnect**.
- My schedule and the person page show it next to the absences.
- A "From calendar" badge on these absences, in the absence list and on the member calendar.

## Acceptance criteria

1. Pasting a Google secret iCal link that holds an all-day "OOO" event next week creates planned leave on those days, marked "From calendar", in the first sync.
2. When the planner approves it and Alex's hours no longer fit, the re-plan runner creates a staffing draft.
3. Moving the event in the calendar moves the absence at the next sync and sets it back to planned. Deleting the event removes the absence.
4. An `http` link, a host outside the allowed list, or a host that resolves to a private IP is refused with `400`, and nothing is fetched.
5. No response ever holds the link, and the logs never do.
6. Anyone but a planner gets `403` on someone else's calendar.
7. A sync never changes an absence entered by hand.

## Tasks

**Server**
- [x] Migration (`calendar_feeds`, `absences.source`, `absences.external_uid`), constants, settings
- [x] Link encryption, and a fetcher with the host, IP, redirect, timeout and size checks
- [x] ICS parsing: which events count, all-day and timed events, time zones
- [x] Sync service: compare events with the calendar's absences and apply the changes through the absence service, in one transaction
- [x] Runner with an FX hook; handlers, routes, Swagger
- [x] Tests: event rules with ICS fixtures (Google and Outlook samples), the sync comparison, SSRF checks, RBAC, an integration test that a sync records change events

**Client**
- [x] Endpoints, types and hooks; `features/calendar-feed` on My schedule and the person page
- [x] "From calendar" badge on the absence list and the calendar
- [x] Tests: connect, status with an error, disconnect

**Before building:** check a real Google and a real Outlook link from the company tenant. Look at how an out-of-office event appears in each, and keep both files as test fixtures, with personal details removed.

**Estimate:** about 8 hours (server 6, client 2).
