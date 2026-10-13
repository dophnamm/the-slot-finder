# Calendar Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** a person pastes the secret iCal (ICS) link of their Google or Outlook calendar, and the server turns its out-of-office events into planned leave that follows the calendar and feeds the existing re-plan flow.

**Architecture:**
- **Fetching and reading the calendar.** A new package `internal/integration/calendar` does four things:
  - fetches the link safely (`Fetcher`)
  - encrypts it at rest (`Sealer`)
  - checks the link (`CheckURL`)
  - reads ICS into leave events (`Parse`)
- **The service.** `services.CalendarFeedService` handles connect, status, disconnect and sync now. It also runs a background runner with an FX lifecycle hook, like the re-plan runner.
- **Writing the absences.** The availability module (`AvailabilityService.ApplyCalendarLeave`) writes all absences, so each change records its change event through the existing outbox. Re-planning works unchanged.
- **The client.** It gets a `features/calendar-feed` panel on My schedule and on the person page, and a "From calendar" badge.

**Tech Stack:**
- Server: Go, Echo v5, Uber FX, GORM, Atlas migrations, testify, testcontainers, `github.com/arran4/golang-ical` v0.3.6
- Client: React, TanStack Query, shadcn/ui on Base UI, Vitest

**Spec:** [`docs/features/calendar-sync.md`](../features/calendar-sync.md). Design context: [root README §6, §8.4, ADR-22, ADR-23](../../README.md), [server/README §3, §6, §8](../../server/README.md).

**Depends on:** [member-calendar](../features/member-calendar.md) merged to `main` first. It adds:
- `models.ScheduleAbsence` in `internal/models/schedule.go` (aliased as `services.ScheduleAbsence`) and `ScheduleAbsenceResponse` in `internal/dtos/schedule.go`
- `POST /absences/{id}/approve` (`AvailabilityService.Approve`, planner only, records a change event, `409` when already approved). Approving calendar leave is that endpoint; this plan does not add another.
- `@/shared/people` on the client: `Person`, `AbsenceType`, `AbsenceStatus`, the absence labels, the schedule types (`ScheduleAbsence`) and `scheduleQueryKey` under `INSIGHTS_QUERY_KEY`
- `features/person-calendar` (`PersonCalendar`)
- `pages/person-detail` (`PersonDetailPage`, route `/people/:id`)

**Branch:** `feat/calendar-sync`, from `main` after member-calendar is merged. One PR. Merge only when the user says "merge" (rebase merge, keep the branch).

## Global Constraints

- **Language.** All code, comments, UI text and docs are in English.
- **Never hardcode.**
  - Server: paths go in `internal/constants/routes.go`, role lists in `roles.go`, messages in `messages.go`, error codes in `error_codes.go`, settings in `internal/config/config.go`, `cmd/server/.env.example` and `server/README.md` §8.
  - Client: endpoints go in `src/resources/endpoints`, shared constants in `src/utils/constants`, i18n keys in `src/shared/i18n/language-key.ts` (sorted case-insensitively).
- **Settings,** names and defaults exactly as in the spec:
  - `CALENDAR_FEED_KEY` (empty = feature off)
  - `CALENDAR_ALLOWED_HOSTS` = `calendar.google.com,outlook.office365.com,outlook.live.com`
  - `CALENDAR_SYNC_INTERVAL` = `15m`
  - `CALENDAR_SYNC_BATCH_SIZE` = `20`
  - `CALENDAR_SYNC_WEEKS` = `26`
  - `CALENDAR_LEAVE_KEYWORDS` = `OOO,Out of office,Vacation,Annual leave,Nghỉ phép`
  - `CALENDAR_FETCH_TIMEOUT` = `10s`
  - `CALENDAR_MAX_BYTES` = 5 MB, written as `5242880`
- **Error codes:**
  - `CALENDAR_SYNC_DISABLED` (503)
  - `CALENDAR_FEED_NOT_FOUND` (404)
  - `CALENDAR_FEED_HOST_NOT_ALLOWED`, `CALENDAR_FEED_UNREACHABLE`, `CALENDAR_FEED_NOT_ICS` and `CALENDAR_FEED_TOO_LARGE` (400 on connect; stored in `last_error` on a sync)
- **The link is a secret.**
  - It is stored only as AES-GCM ciphertext.
  - It is never written to a log, an error message or a response.
  - Every error the fetcher returns is one of four sentinels with no URL in it.
- **No test reaches the network.** `calendar.Fetcher` is an interface. The fetcher's own tests use `httptest.NewTLSServer` on loopback.
- **SSRF guard.** All of these apply:
  - `https` only, with no credentials in the link
  - the host must be in `CALENDAR_ALLOWED_HOSTS`
  - every redirect must also go to an allowed host, at most 3
  - the dialer refuses non-public IPs after DNS (`net.Dialer.Control`), which covers DNS rebinding
  - no proxy, a timeout, and an `io.LimitReader` size cap
- **Server checks.** Every server commit passes `cd server && go build ./... && go vet ./... && go test ./... && make lint`. Commits touching the database also pass `go test -tags=integration ./...`. If Docker is missing, say so in the report.
- **Client checks.** Every client commit passes `cd client && npx tsc -b && npm run lint && npm test && npm run build`. Run `npm run format` first.
- **Commits.** Server and client changes never share a commit. Keep commits small and logical. Every commit message ends with:
  ```
  Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
  ```
- **Generated files.** Swagger is regenerated (`make swagger-load`, or the `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init` command from the Makefile), never edited by hand. The migration is generated (`make migrate-generate name=calendar_feeds`), never hand-written; never edit `atlas.sum` by hand.
- **Client rules:**
  - one component per kebab-case folder (`index.ts`, `component.tsx`, `types.ts`, optional `constants.ts`, `utils.ts`, `api.ts`, `use-<name>.ts`)
  - no inline functions in JSX props, no raw JSX text (tests included), no `style={...}`
  - pastel `Badge` variants chosen through a `Record<…, BadgeVariant>`
  - Base UI `render` prop, not `asChild`
  - features never import each other; dates only through `@/utils/functions`

## Review Focus

These are the inputs most likely to hurt a real person that the spec does not spell out. Each one has a test in the task that owns the code.

1. **The company tenant blocks the secret link.** Google Workspace or Outlook then returns an HTML sign-in page or an error instead of ICS. Connect answers `400 CALENDAR_FEED_NOT_ICS` or `CALENDAR_FEED_UNREACHABLE`, saves nothing, and the UI says what to try. Tests:
   - Task 3: `TestParse_NotICS`
   - Task 7: the connect cases
   - Task 11: the error message is shown
2. **Leave that is already running** (started last week, ends next week). It is kept whole, not cut at today, and later syncs leave it alone. Tests:
   - Task 3: "leave that is already running is kept whole"
   - Task 6: `kept` starts before the window
3. **A person whose weekly hours are 0.** No zero-hour absence is created, because the absence rules need hours > 0. Test: Task 7, `TestCalendarFeedService_LeaveHours`.
4. **A timed event of an odd length** (2 h 35 min = 2.5833 h). After the database rounds it to `numeric(4,2)`, it would look "moved" on every sync and spam re-plan events. Hours are rounded to two decimals before they are saved. Test: Task 7, `TestCalendarFeedService_LeaveHours`.
5. **Two people invited to the same out-of-office event** (a team offsite with one UID). Both get their own absence, and a deleted absence never blocks the same UID coming back. Test: Task 8 (integration).

---

## File Structure

**Server (create)**

| File | Responsibility |
|---|---|
| `server/internal/integration/calendar/errors.go` | The four sentinel errors. None of them holds the link |
| `server/internal/integration/calendar/fetch.go` | `Fetcher`, `FetchPolicy`, `NewHTTPFetcher`, `CheckURL`, `publicAddr`: the SSRF-safe download |
| `server/internal/integration/calendar/sealer.go` | `Sealer` (AES-256-GCM) and `ErrBadKey` |
| `server/internal/integration/calendar/parse.go` | `LeaveRules`, `LeaveEvent`, `Parse`: ICS to out-of-office dates |
| `server/internal/integration/calendar/provide.go` | `ProvideFetcher(cfg)` for FX |
| `server/internal/integration/calendar/*_test.go`, `testdata/google.ics`, `testdata/outlook.ics` | Tests and fixtures |
| `server/internal/models/calendarfeed.go` | `models.CalendarFeed` |
| `server/internal/repositories/calendarfeed.go` | `CalendarFeedRepository`: find, claim due feeds (`SKIP LOCKED`), record a sync, hard delete |
| `server/internal/services/calendarfeed.go` | `CalendarFeedService`: connect, get, disconnect, sync now, `RunOnce` and `Run` |
| `server/internal/services/calendarfeed_test.go`, `calendarleave_test.go`, `calendarfeed_integration_test.go` | Service tests |
| `server/internal/dtos/calendarfeed.go` | `ConnectCalendarFeedRequest`, `CalendarFeedResponse` (never the link) |
| `server/internal/handlers/calendarfeed.go`, `calendarfeed_test.go` | `CalendarFeedHandler` |
| `server/cmd/migrations/sql/<timestamp>_calendar_feeds.sql` | Generated by Atlas |

**Server (modify)**

| File | Change |
|---|---|
| `server/internal/models/absence.go` | `Source`, `ExternalUID`, and a partial unique index on (person, UID) |
| `server/internal/constants/absence.go`, `routes.go`, `roles.go`, `messages.go`, `error_codes.go` | New names |
| `server/internal/config/config.go`, `config_test.go`, `server/cmd/server/.env.example` | `CALENDAR_*` settings, plus `getEnvAsList` and `getEnvAsInt64` |
| `server/internal/repositories/absence.go` | `ListCalendarLeave` |
| `server/internal/services/absence.go`, `absence_test.go` | `ApplyCalendarLeave`, `Source` on new absences, the mock's new method |
| `server/internal/dtos/absence.go` | `source` in `AbsenceResponse` |
| `server/internal/models/schedule.go`, `services/schedule.go`, `dtos/schedule.go` (from member-calendar) | `Source` / `source` on schedule absences |
| `server/cmd/server/routes/router.go`, `router_test.go`, `server/cmd/server/main.go` | Routes, wiring, runner hook |
| `server/go.mod`, `server/go.sum` | `github.com/arran4/golang-ical v0.3.6` |
| `server/docs/*` | Regenerated Swagger |

**Client (create)**

| File | Responsibility |
|---|---|
| `client/src/shared/components/absence-source-badge/` | "From calendar" badge, used by the absence list and the calendar |
| `client/src/features/calendar-feed/` (`component.tsx`, `types.ts`, `constants.ts`, `api.ts`, `use-calendar-feed.ts`, `index.ts`) | The panel: off, not connected, or connected |
| `client/src/features/calendar-feed/components/calendar-feed-form/` | Paste the link, with where to find it |
| `client/src/features/calendar-feed/components/calendar-feed-status/` | Host, last sync, error, events found, Sync now, Disconnect |

**Client (modify)**

| File | Change |
|---|---|
| `client/src/resources/endpoints/people.ts` | `calendarFeed`, `calendarFeedSync` |
| `client/src/utils/constants/query.ts` | `ABSENCES_QUERY_KEY` prefix |
| `client/src/shared/people` (from member-calendar) | `AbsenceSource`, `ABSENCE_SOURCE_BADGE`, `ScheduleAbsence.source` |
| `client/src/features/absences/*` | `source` on `Absence`, the badge in the row, query key from the shared prefix |
| `client/src/resources/rbac/permissions.ts` | `'calendar-feeds:manage'` (planner) |
| `client/src/features/person-calendar/*` | Compact badge on calendar absences, test fixtures |
| `client/src/pages/my-schedule/component.tsx`, `client/src/pages/person-detail/component.tsx` | Render `<CalendarFeed personId=… />` |
| `client/src/shared/i18n/language-key.ts` | New keys |

**Docs (modify, last task):** `docs/features/calendar-sync.md` (status, As built, ticks), `docs/features/README.md` (status).

---

## Tasks

### Task 1: Capture real Google and Outlook samples

The parser must handle what the two providers really send. This task produces the two test fixtures that Task 3 reads. It has no code and no commit of its own: Task 3 commits the fixtures with the parser.

**Files:**
- Create: `server/internal/integration/calendar/testdata/google.ics`
- Create: `server/internal/integration/calendar/testdata/outlook.ics`

**Interfaces:**
- Consumes: nothing.
- Produces: two ICS files with exactly the events below (same UIDs, titles and dates), because Task 3's expected values are written against them.

- [ ] **Step 1: Make a Google sample**

  In a Google calendar you may test with, create these events (time zone Asia/Ho_Chi_Minh):
  - all-day "OOO" from 26 to 28 Oct 2026
  - "Out of office (afternoon)" on 21 Oct, 13:00–17:00
  - a normal meeting, "Sprint review"
  - an all-day "Vacation" on 2 Nov, then cancel it
  - a weekly all-day "OOO Fridays"
  - an all-day "Annual leave" on 1–2 Sep 2026

  Copy the secret address (Settings and sharing → Integrate calendar → **Secret address in iCal format**) and download it:

  ```bash
  read -rs GOOGLE_ICS && curl -sS -o /tmp/google.ics "$GOOGLE_ICS"; unset GOOGLE_ICS
  ```

  `read -s` keeps the link out of shell history. Never paste the link into a commit, PR, issue or chat.

- [ ] **Step 2: Make an Outlook sample**

  In Outlook on the web, create these events:
  - all-day "Away" on 3–4 Nov 2026, **Show as: Away**
  - "Doctor" on 22 Oct, 08:00–12:00, **Show as: Away**
  - "Busy" on 22 Oct, 14:00–15:00, **Show as: Busy**

  Publish the calendar (Settings → Calendar → Shared calendars → Publish a calendar → **Can view all details**), copy the ICS link, and download it:

  ```bash
  read -rs OUTLOOK_ICS && curl -sS -o /tmp/outlook.ics "$OUTLOOK_ICS"; unset OUTLOOK_ICS
  ```

- [ ] **Step 3: Look at how each provider marks out of office**

  ```bash
  grep -nE "^(BEGIN:VEVENT|UID|SUMMARY|DTSTART|DTEND|STATUS|RRULE|RECURRENCE-ID|X-MICROSOFT-CDO-BUSYSTATUS)" /tmp/google.ics /tmp/outlook.ics
  ```

  Expected:
  - Google writes no out-of-office flag in the feed, so its events count through the title keywords.
  - Outlook writes `X-MICROSOFT-CDO-BUSYSTATUS:OOF` on the "Away" events, and uses a Windows `TZID` such as `SE Asia Standard Time`.

  Note anything different in the PR description.

- [ ] **Step 4: Remove personal details and write the fixtures**

  - Keep only these lines:
    - the `VCALENDAR` header lines and the `VTIMEZONE` blocks
    - per event: `UID`, `SUMMARY`, `DTSTART`, `DTEND`, `DTSTAMP`, `STATUS`, `TRANSP`, `RRULE`, `CLASS`, `X-MICROSOFT-CDO-BUSYSTATUS`, `X-MICROSOFT-CDO-ALLDAYEVENT`
  - Delete `ORGANIZER`, `ATTENDEE`, `DESCRIPTION`, `LOCATION`, `X-ALT-DESC` and every email address.
  - Replace the UIDs with the ones below.
  - Then make each file match this content line for line, keeping any extra header lines the provider wrote:

  `server/internal/integration/calendar/testdata/google.ics`:

```text
BEGIN:VCALENDAR
PRODID:-//Google Inc//Google Calendar 70.9054//EN
VERSION:2.0
CALSCALE:GREGORIAN
METHOD:PUBLISH
X-WR-TIMEZONE:Asia/Ho_Chi_Minh
BEGIN:VEVENT
DTSTART;VALUE=DATE:20261026
DTEND;VALUE=DATE:20261029
DTSTAMP:20261013T020000Z
UID:ooo-week@google.com
SUMMARY:OOO
STATUS:CONFIRMED
TRANSP:OPAQUE
END:VEVENT
BEGIN:VEVENT
DTSTART;TZID=Asia/Ho_Chi_Minh:20261021T130000
DTEND;TZID=Asia/Ho_Chi_Minh:20261021T170000
DTSTAMP:20261013T020000Z
UID:half-day@google.com
SUMMARY:Out of office (afternoon)
STATUS:CONFIRMED
END:VEVENT
BEGIN:VEVENT
DTSTART:20261020T020000Z
DTEND:20261020T030000Z
DTSTAMP:20261013T020000Z
UID:standup@google.com
SUMMARY:Sprint review
STATUS:CONFIRMED
END:VEVENT
BEGIN:VEVENT
DTSTART;VALUE=DATE:20261102
DTEND;VALUE=DATE:20261103
DTSTAMP:20261013T020000Z
UID:cancelled@google.com
SUMMARY:Vacation
STATUS:CANCELLED
END:VEVENT
BEGIN:VEVENT
DTSTART;VALUE=DATE:20261023
DTEND;VALUE=DATE:20261024
RRULE:FREQ=WEEKLY;BYDAY=FR
DTSTAMP:20261013T020000Z
UID:fridays@google.com
SUMMARY:OOO Fridays
END:VEVENT
BEGIN:VEVENT
DTSTART;VALUE=DATE:20260901
DTEND;VALUE=DATE:20260903
DTSTAMP:20261013T020000Z
UID:past@google.com
SUMMARY:Annual leave
END:VEVENT
END:VCALENDAR
```

  `server/internal/integration/calendar/testdata/outlook.ics`:

```text
BEGIN:VCALENDAR
METHOD:PUBLISH
PRODID:Microsoft Exchange Server 2010
VERSION:2.0
X-WR-CALNAME:Calendar
BEGIN:VTIMEZONE
TZID:SE Asia Standard Time
BEGIN:STANDARD
DTSTART:16010101T000000
TZOFFSETFROM:+0700
TZOFFSETTO:+0700
END:STANDARD
END:VTIMEZONE
BEGIN:VEVENT
UID:040000008200E00074C5B7101A82E0080000000001
SUMMARY:Away
DTSTART;VALUE=DATE:20261103
DTEND;VALUE=DATE:20261105
CLASS:PUBLIC
TRANSP:OPAQUE
STATUS:CONFIRMED
X-MICROSOFT-CDO-BUSYSTATUS:OOF
X-MICROSOFT-CDO-ALLDAYEVENT:TRUE
END:VEVENT
BEGIN:VEVENT
UID:040000008200E00074C5B7101A82E0080000000002
SUMMARY:Doctor
DTSTART;TZID=SE Asia Standard Time:20261022T080000
DTEND;TZID=SE Asia Standard Time:20261022T120000
STATUS:CONFIRMED
X-MICROSOFT-CDO-BUSYSTATUS:OOF
END:VEVENT
BEGIN:VEVENT
UID:040000008200E00074C5B7101A82E0080000000003
SUMMARY:Busy
DTSTART;TZID=SE Asia Standard Time:20261022T140000
DTEND;TZID=SE Asia Standard Time:20261022T150000
STATUS:CONFIRMED
X-MICROSOFT-CDO-BUSYSTATUS:BUSY
END:VEVENT
END:VCALENDAR
```

  **If the company tenant blocks it:**
  - Google Workspace hides the secret address when the admin turns off external sharing.
  - Outlook hides "Publish a calendar" when anonymous sharing is off.
  
  Then do three things:
  1. Use the files above as they are. They follow the formats the providers document.
  2. Tell the user before 15 Oct that the company tenant blocks secret links, so the demo needs a personal Google or outlook.com calendar (`outlook.live.com` is already allowed).
  3. Do not add hosts or change the approach without asking.

- [ ] **Step 5: No commit yet.** Task 3 commits the fixtures together with the parser that reads them.

---

### Task 2: Calendar sync settings

**Files:**
- Modify: `server/internal/config/config.go` (Config fields, `Load()`, new helpers)
- Modify: `server/internal/config/config_test.go`
- Modify: `server/cmd/server/.env.example`

**Interfaces:**
- Consumes: nothing.
- Produces: the `config.Config` fields `CalendarFeedKey string`, `CalendarAllowedHosts []string`, `CalendarSyncInterval time.Duration`, `CalendarSyncBatchSize int`, `CalendarSyncWeeks int`, `CalendarLeaveKeywords []string`, `CalendarFetchTimeout time.Duration` and `CalendarMaxBytes int64`.

- [ ] **Step 1: Write the failing test.** Append to `server/internal/config/config_test.go`:

```go
func TestLoad_CalendarSettings(t *testing.T) {
	t.Setenv("CALENDAR_FEED_KEY", "")
	t.Setenv("CALENDAR_ALLOWED_HOSTS", "")
	t.Setenv("CALENDAR_LEAVE_KEYWORDS", " OOO , Out of office ,, ")
	t.Setenv("CALENDAR_MAX_BYTES", "1048576")
	cfg, err := Load()
	require.NoError(t, err)
	assert.Empty(t, cfg.CalendarFeedKey, "no key: calendar sync is off")
	assert.Equal(t, []string{"calendar.google.com", "outlook.office365.com", "outlook.live.com"}, cfg.CalendarAllowedHosts)
	assert.Equal(t, []string{"OOO", "Out of office"}, cfg.CalendarLeaveKeywords)
	assert.Equal(t, int64(1048576), cfg.CalendarMaxBytes)
	assert.Equal(t, 15*time.Minute, cfg.CalendarSyncInterval)
	assert.Equal(t, 26, cfg.CalendarSyncWeeks)
}
```

- [ ] **Step 2: Run it to make sure it fails**

  Run: `cd server && go test ./internal/config -run TestLoad_CalendarSettings`
  Expected: FAIL to compile, with `cfg.CalendarFeedKey undefined`.

- [ ] **Step 3: Add the fields.** In `server/internal/config/config.go`, inside `type Config struct`, put this block just before the `// Program metrics (root README §7)` comment:

```go
	// Calendar sync (docs/features/calendar-sync.md): the key that encrypts calendar links (empty
	// turns the feature off), where a link may point, how often and how far ahead the runner
	// reads, which titles mean out of office, and the limits on one fetch.
	CalendarFeedKey       string
	CalendarAllowedHosts  []string
	CalendarSyncInterval  time.Duration
	CalendarSyncBatchSize int
	CalendarSyncWeeks     int
	CalendarLeaveKeywords []string
	CalendarFetchTimeout  time.Duration
	CalendarMaxBytes      int64
```

- [ ] **Step 4: Read them in `Load()`.** Put these lines just before `ScopeLagThreshold: getEnvAsFloat("SCOPE_LAG_THRESHOLD", 20),`:

```go
		CalendarFeedKey:       getEnv("CALENDAR_FEED_KEY", ""),
		CalendarAllowedHosts:  getEnvAsList("CALENDAR_ALLOWED_HOSTS", defaultCalendarAllowedHosts),
		CalendarSyncInterval:  getEnvAsDuration("CALENDAR_SYNC_INTERVAL", 15*time.Minute),
		CalendarSyncBatchSize: getEnvAsInt("CALENDAR_SYNC_BATCH_SIZE", 20),
		CalendarSyncWeeks:     getEnvAsInt("CALENDAR_SYNC_WEEKS", 26),
		CalendarLeaveKeywords: getEnvAsList("CALENDAR_LEAVE_KEYWORDS", defaultCalendarLeaveKeywords),
		CalendarFetchTimeout:  getEnvAsDuration("CALENDAR_FETCH_TIMEOUT", 10*time.Second),
		CalendarMaxBytes:      getEnvAsInt64("CALENDAR_MAX_BYTES", 5<<20),
```

  Put these constants just above `// defaultWorkingWeekdays is Monday to Friday`:

```go
// defaultCalendarAllowedHosts are the hosts that serve Google and Outlook calendar links.
const defaultCalendarAllowedHosts = "calendar.google.com,outlook.office365.com,outlook.live.com"

// defaultCalendarLeaveKeywords are title words that mark an event as out of office.
const defaultCalendarLeaveKeywords = "OOO,Out of office,Vacation,Annual leave,Nghỉ phép"
```

  Put these helpers just above `// getEnvAsBool gets an environment variable as boolean`:

```go
// getEnvAsInt64 gets an environment variable as a 64-bit integer with a fallback value.
func getEnvAsInt64(key string, fallback int64) int64 {
	if value := envValue(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvAsList reads a comma-separated list, trimming each item and dropping empty ones.
func getEnvAsList(key, fallback string) []string {
	var items []string
	for _, item := range strings.Split(getEnv(key, fallback), ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
```

- [ ] **Step 5: List them in `.env.example`.** In `server/cmd/server/.env.example`, add after `SCENARIO_MAX_ASSUMPTIONS=10`:

```bash
# Calendar sync (docs/features/calendar-sync.md). An empty key turns the feature off.
# Make a key with: openssl rand -base64 32  (a secret: never commit a real one)
CALENDAR_FEED_KEY=
CALENDAR_ALLOWED_HOSTS=calendar.google.com,outlook.office365.com,outlook.live.com
CALENDAR_SYNC_INTERVAL=15m
CALENDAR_SYNC_BATCH_SIZE=20
CALENDAR_SYNC_WEEKS=26
CALENDAR_LEAVE_KEYWORDS="OOO,Out of office,Vacation,Annual leave,Nghỉ phép"
CALENDAR_FETCH_TIMEOUT=10s
CALENDAR_MAX_BYTES=5242880
```

  `server/README.md` §8 already lists these settings, from the docs PR. Check that its defaults match.

- [ ] **Step 6: Run the tests to make sure they pass**

  Run: `cd server && gofmt -l . && go test ./internal/config`
  Expected: no files listed, then `ok`.

- [ ] **Step 7: Run the server checks, then commit**

```bash
cd server && go build ./... && go vet ./... && go test ./... && make lint
git add internal/config/config.go internal/config/config_test.go cmd/server/.env.example
git commit -m "feat(server): calendar sync settings" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Read ICS into out-of-office dates

**Why `github.com/arran4/golang-ical` v0.3.6:**
- It is maintained: v0.3.7 came out on 2026-09-30. v0.3.6 (2026-08-18) is pinned because it is more than two weeks old.
- It reads `VALUE=DATE` and `TZID` start and end times.
- `ics.WithWindowsTimezoneMapping()` maps Outlook's Windows zone names (`SE Asia Standard Time`) to IANA zones. Without that, Outlook timed events fail to parse.
- `github.com/emersion/go-ical` has only untagged pseudo-versions (the latest from June 2025).

**Files:**
- Create: `server/internal/integration/calendar/errors.go`
- Create: `server/internal/integration/calendar/parse.go`
- Create: `server/internal/integration/calendar/parse_test.go`
- Commit: the two fixtures from Task 1
- Modify: `server/go.mod`, `server/go.sum`

**Interfaces:**
- Consumes: the fixtures from Task 1.
- Produces:
  - `calendar.LeaveRules{Keywords []string; From, To time.Time}`
  - `calendar.LeaveEvent{UID string; From, To time.Time; HoursPerDay *float64}`. From and To are inclusive dates at UTC midnight; a nil HoursPerDay means full days.
  - `calendar.Parse(data []byte, rules LeaveRules, loc *time.Location) ([]LeaveEvent, error)`
  - the errors `calendar.ErrHostNotAllowed`, `ErrUnreachable`, `ErrNotICS` and `ErrTooLarge`

- [ ] **Step 1: Add the dependency**

  Run: `cd server && go get github.com/arran4/golang-ical@v0.3.6`
  Expected: `go: added github.com/arran4/golang-ical v0.3.6`.

- [ ] **Step 2: Write the errors.** Create `server/internal/integration/calendar/errors.go`:

```go
package calendar

import "errors"

// Errors a fetch or a parse can return. They never hold the link, which is a secret.
var (
	ErrHostNotAllowed = errors.New("calendar link host is not allowed")
	ErrUnreachable    = errors.New("calendar link could not be fetched")
	ErrNotICS         = errors.New("calendar link is not an ICS calendar")
	ErrTooLarge       = errors.New("calendar is larger than the limit")
)
```

- [ ] **Step 3: Write the failing test.** Create `server/internal/integration/calendar/parse_test.go`:

```go
package calendar

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKeywords = []string{"OOO", "Out of office", "Vacation", "Annual leave", "Nghỉ phép"}

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return data
}

func hoursOf(v float64) *float64 { return &v }

func TestParse(t *testing.T) {
	saigon, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	require.NoError(t, err)
	rules := LeaveRules{Keywords: testKeywords, From: day(2026, time.October, 13), To: day(2027, time.April, 12)}

	tests := []struct {
		name    string
		fixture string
		want    []LeaveEvent
	}{
		{name: "google: all-day OOO and a half day; skips meetings, cancelled, recurring and past events",
			fixture: "google.ics",
			want: []LeaveEvent{
				{UID: "half-day@google.com", From: day(2026, time.October, 21), To: day(2026, time.October, 21), HoursPerDay: hoursOf(4)},
				{UID: "ooo-week@google.com", From: day(2026, time.October, 26), To: day(2026, time.October, 28)},
			}},
		{name: "outlook: OOF flag with a Windows time zone; skips busy time",
			fixture: "outlook.ics",
			want: []LeaveEvent{
				{UID: "040000008200E00074C5B7101A82E0080000000002", From: day(2026, time.October, 22), To: day(2026, time.October, 22), HoursPerDay: hoursOf(4)},
				{UID: "040000008200E00074C5B7101A82E0080000000001", From: day(2026, time.November, 3), To: day(2026, time.November, 4)},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(readFixture(t, tt.fixture), rules, saigon)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParse_NotICS(t *testing.T) {
	_, err := Parse([]byte("<html>Sign in</html>"), LeaveRules{}, time.UTC)
	assert.ErrorIs(t, err, ErrNotICS)
}

func TestParse_Edges(t *testing.T) {
	saigon, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	require.NoError(t, err)
	rules := LeaveRules{Keywords: testKeywords, From: day(2026, time.October, 13), To: day(2027, time.April, 12)}
	wrap := func(event string) []byte {
		return []byte("BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\n" + event + "END:VEVENT\r\nEND:VCALENDAR\r\n")
	}
	tests := []struct {
		name string
		ics  []byte
		want []LeaveEvent
	}{
		{name: "leave that is already running is kept whole",
			ics:  wrap("UID:running\r\nSUMMARY:Annual leave\r\nDTSTART;VALUE=DATE:20261009\r\nDTEND;VALUE=DATE:20261017\r\n"),
			want: []LeaveEvent{{UID: "running", From: day(2026, time.October, 9), To: day(2026, time.October, 16)}}},
		{name: "a time without a zone is read in the organization's zone",
			ics: wrap("UID:floating\r\nSUMMARY:Out of office\r\nDTSTART:20261022T140000\r\nDTEND:20261022T160000\r\n"),
			want: []LeaveEvent{{UID: "floating", From: day(2026, time.October, 22), To: day(2026, time.October, 22),
				HoursPerDay: hoursOf(2)}}},
		{name: "a timed event over two days counts full days",
			ics:  wrap("UID:trip\r\nSUMMARY:OOO trip\r\nDTSTART:20261022T020000Z\r\nDTEND:20261023T100000Z\r\n"),
			want: []LeaveEvent{{UID: "trip", From: day(2026, time.October, 22), To: day(2026, time.October, 23)}}},
		{name: "keywords ignore case, including Vietnamese",
			ics:  wrap("UID:vn\r\nSUMMARY:nghỉ phép\r\nDTSTART;VALUE=DATE:20261102\r\n"),
			want: []LeaveEvent{{UID: "vn", From: day(2026, time.November, 2), To: day(2026, time.November, 2)}}},
		{name: "an event beyond the window is skipped",
			ics: wrap("UID:far\r\nSUMMARY:OOO\r\nDTSTART;VALUE=DATE:20270601\r\nDTEND;VALUE=DATE:20270602\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.ics, rules, saigon)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

- [ ] **Step 4: Run it to make sure it fails**

  Run: `cd server && go test ./internal/integration/calendar`
  Expected: FAIL to compile, with `undefined: LeaveRules`.

- [ ] **Step 5: Write the parser.** Create `server/internal/integration/calendar/parse.go`:

```go
package calendar

import (
	"bytes"
	"sort"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

// Outlook writes this property on every event; OOF means "out of office".
const (
	propBusyStatus  = ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS")
	busyStatusOOF   = "OOF"
	paramValue      = "VALUE"
	paramTZID       = "TZID"
	valueDate       = "DATE"
	statusCancelled = "CANCELLED"
	utcSuffix       = "Z"
	icsMarker       = "BEGIN:VCALENDAR"
)

// LeaveRules say which events count as out of office, and the dates to keep.
type LeaveRules struct {
	// Keywords are matched inside the event title, ignoring case.
	Keywords []string
	// From and To are the first and last date kept (inclusive, UTC midnight). An event that
	// touches the range is kept whole.
	From time.Time
	To   time.Time
}

// LeaveEvent is one out-of-office event as dates. From and To are inclusive dates at UTC
// midnight, like the dates the database stores. HoursPerDay is set for an event inside one
// day (its length); nil means full days.
type LeaveEvent struct {
	UID         string
	From        time.Time
	To          time.Time
	HoursPerDay *float64
}

// Parse reads an ICS document and returns its out-of-office events that touch the rules'
// range, sorted by date. Cancelled, recurring and unreadable events are skipped. Times are read
// in loc, the organization's time zone, when the event gives none.
func Parse(data []byte, rules LeaveRules, loc *time.Location) ([]LeaveEvent, error) {
	if !bytes.Contains(data, []byte(icsMarker)) {
		return nil, ErrNotICS
	}
	cal, err := ics.ParseCalendarWithOptions(bytes.NewReader(data), ics.WithWindowsTimezoneMapping())
	if err != nil {
		return nil, ErrNotICS
	}
	seen := map[string]bool{}
	var out []LeaveEvent
	for _, event := range cal.Events() {
		uid := event.Id()
		if uid == "" || seen[uid] || skipped(event) || !isLeave(event, rules.Keywords) {
			continue
		}
		leave, ok := toLeave(event, uid, loc)
		if !ok || leave.To.Before(rules.From) || leave.From.After(rules.To) {
			continue
		}
		seen[uid] = true
		out = append(out, leave)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].From.Equal(out[j].From) {
			return out[i].From.Before(out[j].From)
		}
		return out[i].UID < out[j].UID
	})
	return out, nil
}

// skipped reports a cancelled event, a recurring series, or one instance of a series.
func skipped(event *ics.VEvent) bool {
	if status := event.GetProperty(ics.ComponentPropertyStatus); status != nil && strings.EqualFold(status.Value, statusCancelled) {
		return true
	}
	return event.GetProperty(ics.ComponentPropertyRrule) != nil || event.GetProperty(ics.ComponentPropertyRecurrenceId) != nil
}

// isLeave reports an event Outlook marks out of office, or whose title holds a keyword.
func isLeave(event *ics.VEvent, keywords []string) bool {
	if busy := event.GetProperty(propBusyStatus); busy != nil && strings.EqualFold(busy.Value, busyStatusOOF) {
		return true
	}
	summary := event.GetProperty(ics.ComponentPropertySummary)
	if summary == nil {
		return false
	}
	title := strings.ToLower(summary.Value)
	for _, k := range keywords {
		if k = strings.ToLower(strings.TrimSpace(k)); k != "" && strings.Contains(title, k) {
			return true
		}
	}
	return false
}

// toLeave turns an event into dates. An all-day event covers DTSTART up to the day before
// DTEND. A timed event inside one day keeps its length; over several days it counts full days.
func toLeave(event *ics.VEvent, uid string, loc *time.Location) (LeaveEvent, bool) {
	start := event.GetProperty(ics.ComponentPropertyDtStart)
	if start == nil {
		return LeaveEvent{}, false
	}
	if allDay(start) {
		from, err := event.GetAllDayStartAt()
		if err != nil {
			return LeaveEvent{}, false
		}
		to := dateOf(from)
		if end, err := event.GetAllDayEndAt(); err == nil && dateOf(end).After(dateOf(from)) {
			to = dateOf(end).AddDate(0, 0, -1)
		}
		return LeaveEvent{UID: uid, From: dateOf(from), To: to}, true
	}

	from, err := event.GetStartAt()
	if err != nil {
		return LeaveEvent{}, false
	}
	to, err := event.GetEndAt()
	if err != nil || !to.After(from) {
		return LeaveEvent{}, false
	}
	from = inZone(from, start, loc)
	if end := event.GetProperty(ics.ComponentPropertyDtEnd); end != nil {
		to = inZone(to, end, loc)
	}
	first, last := dateOf(from), dateOf(to)
	if last.After(first) && to.Equal(time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, loc)) {
		last = last.AddDate(0, 0, -1) // ends at midnight: the last day is the one before
	}
	if last.After(first) {
		return LeaveEvent{UID: uid, From: first, To: last}, true
	}
	hours := to.Sub(from).Hours()
	return LeaveEvent{UID: uid, From: first, To: first, HoursPerDay: &hours}, true
}

// allDay reports a DTSTART written as a date (VALUE=DATE, or eight digits).
func allDay(start *ics.IANAProperty) bool {
	if values, ok := start.ICalParameters[paramValue]; ok && len(values) > 0 && strings.EqualFold(values[0], valueDate) {
		return true
	}
	return len(start.Value) == len("20060102")
}

// inZone moves a time into loc. A "floating" time (no TZID, no trailing Z) was read in
// time.Local by the library, so its clock reading is kept and placed in loc instead.
func inZone(t time.Time, prop *ics.IANAProperty, loc *time.Location) time.Time {
	_, hasTZID := prop.ICalParameters[paramTZID]
	if !hasTZID && !strings.HasSuffix(prop.Value, utcSuffix) {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
	}
	return t.In(loc)
}

// dateOf is the calendar date of t, at UTC midnight.
func dateOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
```

- [ ] **Step 6: Run the tests to make sure they pass**

  Run: `cd server && go mod tidy && go test ./internal/integration/calendar -v -run TestParse`
  Expected: `--- PASS: TestParse`, `--- PASS: TestParse_NotICS`, `--- PASS: TestParse_Edges`, and `go.mod` lists `github.com/arran4/golang-ical v0.3.6` without `// indirect`.

- [ ] **Step 7: Run the server checks, then commit**

```bash
cd server && go build ./... && go vet ./... && go test ./... && make lint
git add go.mod go.sum internal/integration/calendar/errors.go internal/integration/calendar/parse.go \
  internal/integration/calendar/parse_test.go internal/integration/calendar/testdata
git commit -m "feat(server): read out-of-office events from an ICS calendar" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Fetch a calendar safely, and encrypt its link

**Files:**
- Create: `server/internal/integration/calendar/fetch.go`
- Create: `server/internal/integration/calendar/fetch_test.go`
- Create: `server/internal/integration/calendar/sealer.go`
- Create: `server/internal/integration/calendar/sealer_test.go`
- Create: `server/internal/integration/calendar/provide.go`

**Interfaces:**
- Consumes:
  - the errors from Task 3
  - `config.Config.CalendarAllowedHosts`, `CalendarFetchTimeout` and `CalendarMaxBytes` from Task 2
- Produces:
  - `calendar.Fetcher` (interface `Fetch(ctx context.Context, rawURL string) ([]byte, error)`)
  - `calendar.FetchPolicy{AllowedHosts []string; Timeout time.Duration; MaxBytes int64}`
  - `calendar.NewHTTPFetcher(FetchPolicy) Fetcher`
  - `calendar.CheckURL(rawURL string, allowedHosts []string) (host string, err error)`
  - `calendar.ProvideFetcher(cfg *config.Config) Fetcher`
  - `calendar.NewSealer(keyBase64 string) (*Sealer, error)`, `(*Sealer).Seal(link string) ([]byte, error)`, `(*Sealer).Open([]byte) (string, error)`, `calendar.ErrBadKey`

- [ ] **Step 1: Write the failing fetch test.** Create `server/internal/integration/calendar/fetch_test.go`.

  It runs a local TLS server, so nothing leaves the machine. The "resolves to loopback" case shows that the dial-time check refuses an allowed host name whose address is not public: this is the DNS-rebinding guard.

```go
package calendar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	loopbackHost = "127.0.0.1"
	secretPath   = "/calendar/ical/secret-token-123/basic.ics"
	icsBody      = "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"
	testMaxBytes = 64
)

func allowAll(netip.Addr) bool { return true }

// testFetcher talks to a local TLS test server: it trusts its certificate and allows
// loopback, which the real fetcher refuses.
func testFetcher(srv *httptest.Server, allowAddr func(netip.Addr) bool, hosts ...string) *httpFetcher {
	policy := FetchPolicy{AllowedHosts: hosts, Timeout: 2 * time.Second, MaxBytes: testMaxBytes}
	return newHTTPFetcher(policy, allowAddr, srv.Client().Transport.(*http.Transport).TLSClientConfig.RootCAs)
}

func TestHTTPFetcher_Fetch(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case secretPath:
			_, _ = w.Write([]byte(icsBody))
		case "/big":
			_, _ = w.Write([]byte(strings.Repeat("x", testMaxBytes+1)))
		case "/away":
			http.Redirect(w, r, "https://evil.example/steal", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	tests := []struct {
		name    string
		fetcher *httpFetcher
		url     string
		want    string
		wantErr error
	}{
		{name: "reads an allowed calendar", fetcher: testFetcher(srv, allowAll, loopbackHost), url: srv.URL + secretPath, want: icsBody},
		{name: "http is refused before any request", fetcher: testFetcher(srv, allowAll, loopbackHost),
			url: strings.Replace(srv.URL, "https://", "http://", 1) + secretPath, wantErr: ErrHostNotAllowed},
		{name: "a host outside the list is refused", fetcher: testFetcher(srv, allowAll, "calendar.google.com"),
			url: srv.URL + secretPath, wantErr: ErrHostNotAllowed},
		{name: "an allowed name that resolves to loopback is refused when dialling", fetcher: testFetcher(srv, publicAddr, loopbackHost),
			url: srv.URL + secretPath, wantErr: ErrHostNotAllowed},
		{name: "a redirect to another host is refused", fetcher: testFetcher(srv, allowAll, loopbackHost),
			url: srv.URL + "/away", wantErr: ErrHostNotAllowed},
		{name: "a body over the limit", fetcher: testFetcher(srv, allowAll, loopbackHost), url: srv.URL + "/big", wantErr: ErrTooLarge},
		{name: "a missing calendar", fetcher: testFetcher(srv, allowAll, loopbackHost), url: srv.URL + "/gone", wantErr: ErrUnreachable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fetcher.Fetch(context.Background(), tt.url)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.NotContains(t, err.Error(), "secret-token", "an error must never hold the link")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestCheckURL(t *testing.T) {
	hosts := []string{"calendar.google.com", "outlook.office365.com"}
	tests := []struct {
		name     string
		url      string
		wantHost string
	}{
		{name: "google secret address", url: "https://calendar.google.com/calendar/ical/x%40y/private-abc/basic.ics", wantHost: "calendar.google.com"},
		{name: "host matched ignoring case", url: "https://Outlook.Office365.com/owa/calendar/a/b/calendar.ics", wantHost: "outlook.office365.com"},
		{name: "http", url: "http://calendar.google.com/x.ics"},
		{name: "credentials in the link", url: "https://user:pw@calendar.google.com/x.ics"},
		{name: "look-alike host", url: "https://calendar.google.com.evil.example/x.ics"},
		{name: "not a link", url: "calendar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, err := CheckURL(tt.url, hosts)
			if tt.wantHost == "" {
				assert.ErrorIs(t, err, ErrHostNotAllowed)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, host)
		})
	}
}

func TestPublicAddr(t *testing.T) {
	for addr, want := range map[string]bool{
		"8.8.8.8": true, "2001:4860:4860::8888": true,
		"127.0.0.1": false, "10.1.2.3": false, "172.16.0.1": false, "192.168.1.1": false, "169.254.169.254": false,
		"100.64.0.1": false, "0.0.0.0": false, "::1": false, "fd00::1": false, "fe80::1": false,
	} {
		assert.Equal(t, want, publicAddr(netip.MustParseAddr(addr)), addr)
	}
}
```

- [ ] **Step 2: Write the failing sealer test.** Create `server/internal/integration/calendar/sealer_test.go`:

```go
package calendar

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testLink = "https://calendar.google.com/calendar/ical/a%40b/private-xyz/basic.ics"

func testKey(fill byte) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Repeat(string(fill), keyBytes)))
}

func TestSealer(t *testing.T) {
	sealer, err := NewSealer(testKey('k'))
	require.NoError(t, err)

	sealed, err := sealer.Seal(testLink)
	require.NoError(t, err)
	assert.NotContains(t, string(sealed), "private-xyz")

	again, err := sealer.Seal(testLink)
	require.NoError(t, err)
	assert.NotEqual(t, sealed, again, "each seal uses a fresh nonce")

	opened, err := sealer.Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, testLink, opened)

	other, err := NewSealer(testKey('o'))
	require.NoError(t, err)
	_, err = other.Open(sealed)
	assert.Error(t, err, "another key cannot open it")
}

func TestNewSealer_BadKey(t *testing.T) {
	for _, key := range []string{"", "not base64!", base64.StdEncoding.EncodeToString([]byte("short"))} {
		_, err := NewSealer(key)
		assert.ErrorIs(t, err, ErrBadKey, key)
	}
}
```

- [ ] **Step 3: Run them to make sure they fail**

  Run: `cd server && go test ./internal/integration/calendar`
  Expected: FAIL to compile, with `undefined: newHTTPFetcher` and `undefined: NewSealer`.

- [ ] **Step 4: Write the fetcher.** Create `server/internal/integration/calendar/fetch.go`:

```go
package calendar

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	stderrors "errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strings"
	"syscall"
	"time"
)

// schemeHTTPS is the only scheme a calendar link may use.
const schemeHTTPS = "https"

// maxRedirects is how many redirects one fetch follows, each to an allowed host.
const maxRedirects = 3

// blockedPrefixes are address ranges that are not on the public internet, beyond the ones
// net/netip already names (loopback, private, link-local, multicast, unspecified).
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("240.0.0.0/4"),
}

// Fetcher downloads a calendar. Tests pass a fake, so no test reaches the network.
type Fetcher interface {
	Fetch(ctx context.Context, rawURL string) ([]byte, error)
}

// FetchPolicy limits what a fetch may reach and read.
type FetchPolicy struct {
	AllowedHosts []string
	Timeout      time.Duration
	MaxBytes     int64
}

type httpFetcher struct {
	policy FetchPolicy
	client *http.Client
}

// NewHTTPFetcher fetches over https from allowed hosts only, and refuses to connect to an
// address that is not public, whatever the host name resolves to (DNS rebinding included).
func NewHTTPFetcher(policy FetchPolicy) Fetcher {
	return newHTTPFetcher(policy, publicAddr, nil)
}

// newHTTPFetcher lets tests allow the loopback test server and trust its certificate.
func newHTTPFetcher(policy FetchPolicy, allowAddr func(netip.Addr) bool, roots *x509.CertPool) *httpFetcher {
	dialer := &net.Dialer{
		Timeout: policy.Timeout,
		// Control runs after DNS, on the address actually dialled.
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return ErrHostNotAllowed
			}
			addr, err := netip.ParseAddr(host)
			if err != nil || !allowAddr(addr.Unmap()) {
				return ErrHostNotAllowed
			}
			return nil
		},
	}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots},
		TLSHandshakeTimeout:   policy.Timeout,
		ResponseHeaderTimeout: policy.Timeout,
	}
	f := &httpFetcher{policy: policy}
	f.client = &http.Client{Transport: transport, Timeout: policy.Timeout, CheckRedirect: f.checkRedirect}
	return f
}

func (f *httpFetcher) checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return ErrUnreachable
	}
	if _, err := CheckURL(req.URL.String(), f.policy.AllowedHosts); err != nil {
		return err
	}
	return nil
}

func (f *httpFetcher) Fetch(ctx context.Context, rawURL string) ([]byte, error) {
	if _, err := CheckURL(rawURL, f.policy.AllowedHosts); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, ErrUnreachable
	}
	resp, err := f.client.Do(req)
	if err != nil {
		// The error text holds the link, so only the kind of failure goes back.
		if stderrors.Is(err, ErrHostNotAllowed) {
			return nil, ErrHostNotAllowed
		}
		return nil, ErrUnreachable
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnreachable
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, f.policy.MaxBytes+1))
	if err != nil {
		return nil, ErrUnreachable
	}
	if int64(len(data)) > f.policy.MaxBytes {
		return nil, ErrTooLarge
	}
	return data, nil
}

// CheckURL accepts an https link without credentials to an allowed host, and returns the host.
func CheckURL(rawURL string, allowedHosts []string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Scheme != schemeHTTPS || u.User != nil || u.Hostname() == "" {
		return "", ErrHostNotAllowed
	}
	host := strings.ToLower(u.Hostname())
	if !slices.ContainsFunc(allowedHosts, func(h string) bool { return strings.EqualFold(strings.TrimSpace(h), host) }) {
		return "", ErrHostNotAllowed
	}
	return host, nil
}

// publicAddr reports an address on the public internet.
func publicAddr(addr netip.Addr) bool {
	if !addr.IsValid() || addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() ||
		addr.IsInterfaceLocalMulticast() || addr.IsMulticast() || addr.IsUnspecified() {
		return false
	}
	return !slices.ContainsFunc(blockedPrefixes, func(p netip.Prefix) bool { return p.Contains(addr) })
}
```

- [ ] **Step 5: Write the sealer.** Create `server/internal/integration/calendar/sealer.go`:

```go
package calendar

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// keyBytes is the AES-256 key length CALENDAR_FEED_KEY must decode to.
const keyBytes = 32

// ErrBadKey means CALENDAR_FEED_KEY is not 32 bytes of base64.
var ErrBadKey = errors.New("CALENDAR_FEED_KEY must be 32 bytes, base64-encoded")

// errSealed means a stored link could not be opened: a wrong key or damaged data.
var errSealed = errors.New("calendar link could not be decrypted")

// Sealer encrypts calendar links at rest with AES-GCM. A sealed link is nonce ‖ ciphertext.
type Sealer struct {
	aead cipher.AEAD
}

// NewSealer reads a base64 key of 32 bytes.
func NewSealer(keyBase64 string) (*Sealer, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil || len(key) != keyBytes {
		return nil, ErrBadKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrBadKey
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrBadKey
	}
	return &Sealer{aead: aead}, nil
}

// Seal encrypts a link with a fresh random nonce.
func (s *Sealer) Seal(link string) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, []byte(link), nil), nil
}

// Open decrypts a sealed link.
func (s *Sealer) Open(sealed []byte) (string, error) {
	n := s.aead.NonceSize()
	if len(sealed) < n {
		return "", errSealed
	}
	plain, err := s.aead.Open(nil, sealed[:n], sealed[n:], nil)
	if err != nil {
		return "", errSealed
	}
	return string(plain), nil
}
```

- [ ] **Step 6: Write the FX provider.** Create `server/internal/integration/calendar/provide.go`:

```go
package calendar

import "the-slot-finder/server/internal/config"

// ProvideFetcher builds the HTTP fetcher from the CALENDAR_* settings.
func ProvideFetcher(cfg *config.Config) Fetcher {
	return NewHTTPFetcher(FetchPolicy{AllowedHosts: cfg.CalendarAllowedHosts, Timeout: cfg.CalendarFetchTimeout,
		MaxBytes: cfg.CalendarMaxBytes})
}
```

- [ ] **Step 7: Run the tests to make sure they pass**

  Run: `cd server && go test ./internal/integration/calendar -v`
  Expected: every test passes, including `TestHTTPFetcher_Fetch/an_allowed_name_that_resolves_to_loopback_is_refused_when_dialling` and `TestSealer`.

- [ ] **Step 8: Run the server checks, then commit**

```bash
cd server && go build ./... && go vet ./... && go test ./... && make lint
git add internal/integration/calendar/fetch.go internal/integration/calendar/fetch_test.go \
  internal/integration/calendar/sealer.go internal/integration/calendar/sealer_test.go internal/integration/calendar/provide.go
git commit -m "feat(server): fetch calendar links safely and encrypt them at rest" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Absences know their source, and calendar feeds have a table

**Files:**
- Modify: `server/internal/constants/absence.go`
- Modify: `server/internal/models/absence.go`
- Create: `server/internal/models/calendarfeed.go`
- Modify: `server/internal/repositories/absence.go`
- Create: `server/internal/repositories/calendarfeed.go`
- Modify: `server/internal/services/absence.go` (`Create`, `CreateHoliday`)
- Modify: `server/internal/services/absence_test.go` (test, mock)
- Modify: `server/internal/dtos/absence.go`
- Create (generated): `server/cmd/migrations/sql/<timestamp>_calendar_feeds.sql`, and `atlas.sum` updated by Atlas

**Interfaces:**
- Consumes: nothing new.
- Produces:
  - `constants.AbsenceSourceManual = "manual"` and `constants.AbsenceSourceCalendar = "calendar"`
  - `models.Absence.Source string` and `models.Absence.ExternalUID *string`
  - `models.CalendarFeed{PersonID string; URLCiphertext []byte; Host string; LastSyncedAt *time.Time; LastError *string; EventsFound int; NextSyncAt time.Time}`
  - `AbsenceRepository.ListCalendarLeave(ctx context.Context, personID string) ([]models.Absence, error)`
  - `CalendarFeedRepository`: `FindByPerson`, `ClaimDue(ctx, now, next time.Time, limit int)`, `RecordSync(ctx, id string, at time.Time, eventsFound int, errorCode *string)`, `DeleteByPerson`, plus the shared `AbstractRepository` methods
  - `ProvideCalendarFeedRepository`
  - `dtos.AbsenceResponse.Source` (JSON `source`)

`NextSyncAt` is not in the spec's column list. The runner needs it to claim a calendar for one interval and release the row lock at once, so no lock is held during a network fetch.

- [ ] **Step 1: Write the failing test.** In `server/internal/services/absence_test.go`, `TestAvailabilityService_Create`, change the `Create` matcher so it also asks for the manual source:

```go
				repo.On("Create", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
					return a.Status == tt.wantStatus && *a.PersonID == tt.in.PersonID && a.Type == tt.in.Type &&
						a.Source == constants.AbsenceSourceManual
				})).Return(nil)
```

- [ ] **Step 2: Run it to make sure it fails**

  Run: `cd server && go test ./internal/services -run TestAvailabilityService_Create`
  Expected: FAIL to compile, with `a.Source undefined`.

- [ ] **Step 3: Add the sources.** In `server/internal/constants/absence.go`, add above `// PersonalAbsenceTypes are`:

```go
// Absence sources: entered by hand, or read from the person's calendar feed.
const (
	AbsenceSourceManual   = "manual"
	AbsenceSourceCalendar = "calendar"
)
```

- [ ] **Step 4: Change the absence model.** Replace `server/internal/models/absence.go` with:

```go
package models

import "time"

// Absence is time a person cannot work: leave, training, or a holiday.
// A holiday has no person: it applies to everyone (company-wide).
type Absence struct {
	BaseModel
	// PersonID is null for a company-wide holiday.
	PersonID    *string   `gorm:"column:person_id;type:uuid;index;index:idx_absences_person_external_uid,unique,priority:1,where:external_uid IS NOT NULL AND deleted_at IS NULL"`
	FromDate    time.Time `gorm:"column:from_date;type:date;not null;index:idx_absences_dates"`
	ToDate      time.Time `gorm:"column:to_date;type:date;not null;index:idx_absences_dates"`
	HoursPerDay float64   `gorm:"column:hours_per_day;type:numeric(4,2);not null"`
	Type        string    `gorm:"column:type;not null"`
	Status      string    `gorm:"column:status;not null"`
	// Source is manual, or calendar for leave read from the person's calendar feed.
	Source string `gorm:"column:source;not null;default:manual"`
	// ExternalUID is the calendar event's UID; nil for an absence entered by hand.
	ExternalUID *string `gorm:"column:external_uid;index:idx_absences_person_external_uid,unique,priority:2,where:external_uid IS NOT NULL AND deleted_at IS NULL"`

	Person *Person `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
}
```

  The partial unique index lets two people share one event UID (a team offsite), and lets a deleted absence's UID come back. Soft-deleted rows have `deleted_at` set, so they fall outside the index.

- [ ] **Step 5: Add the calendar feed model.** Create `server/internal/models/calendarfeed.go`:

```go
package models

import "time"

// CalendarFeed is a person's secret calendar link (ICS), stored encrypted, with how its last
// sync went. Leave read from it is stored as absences with source calendar.
type CalendarFeed struct {
	BaseModel
	PersonID      string     `gorm:"column:person_id;type:uuid;not null;uniqueIndex"`
	URLCiphertext []byte     `gorm:"column:url_ciphertext;type:bytea;not null"`
	Host          string     `gorm:"column:host;not null"`
	LastSyncedAt  *time.Time `gorm:"column:last_synced_at;type:timestamptz"`
	LastError     *string    `gorm:"column:last_error"`
	EventsFound   int        `gorm:"column:events_found;not null;default:0"`
	// NextSyncAt is when the runner syncs it next; claiming a feed moves it one interval ahead.
	NextSyncAt time.Time `gorm:"column:next_sync_at;type:timestamptz;not null;index"`

	Person *Person `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
}
```

- [ ] **Step 6: Save the source on new absences.**
  - In `server/internal/services/absence.go`, add `Source: constants.AbsenceSourceManual,` to the `models.Absence` literal in `Create`, after `Status: status,`.
  - Add it to the one in `CreateHoliday` too, after `Status: constants.AbsenceStatusApproved,`.
  - In `server/internal/dtos/absence.go`, add this to `AbsenceResponse`, after `Status`:

```go
	Source      string  `json:"source" example:"manual" enums:"manual,calendar"`
```

  and to `NewAbsenceResponse`, after `Status: a.Status,`:

```go
		Source:      a.Source,
```

  Leave read from a calendar must show as such on the member calendar too. The schedule types come from member-calendar:
  - In `server/internal/models/schedule.go`, add to `ScheduleAbsence`, after `HoursPerDay`:

```go
	// Source is manual, or calendar for leave read from the person's calendar feed.
	Source string
```

  - In `server/internal/services/schedule.go`, where the day's absences are built, add the source to the literal:

```go
day.Absences = append(day.Absences, ScheduleAbsence{ID: a.ID, Type: a.Type, Status: a.Status, HoursPerDay: a.HoursPerDay, Source: a.Source})
```

  - In `server/internal/dtos/schedule.go`, add to `ScheduleAbsenceResponse`, after `HoursPerDay`:

```go
	Source      string  `json:"source" enums:"manual,calendar"`
```

    and `Source: a.Source` to its literal in the mapping:

```go
absences = append(absences, ScheduleAbsenceResponse{ID: a.ID, Type: a.Type, Status: a.Status, HoursPerDay: a.HoursPerDay, Source: a.Source})
```

- [ ] **Step 7: Add the repositories.**
  - In `server/internal/repositories/absence.go`, add `"the-slot-finder/server/internal/constants"` to the imports.
  - Add this method to the `AbsenceRepository` interface, after `ListOverlapping`:

```go
	// ListCalendarLeave returns a person's absences read from their calendar, oldest first.
	ListCalendarLeave(ctx context.Context, personID string) ([]models.Absence, error)
```

  - Append its implementation:

```go
func (r *absenceRepository) ListCalendarLeave(ctx context.Context, personID string) ([]models.Absence, error) {
	var absences []models.Absence
	err := r.db.From(ctx).
		Where("person_id = ? AND source = ?", personID, constants.AbsenceSourceCalendar).
		Order("from_date ASC").Order("id ASC").
		Find(&absences).Error
	return absences, err
}
```

  - Create `server/internal/repositories/calendarfeed.go`:

```go
package repositories

import (
	"context"
	"time"

	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/models"

	"gorm.io/gorm/clause"
)

// CalendarFeedRepository stores people's calendar links.
type CalendarFeedRepository interface {
	AbstractRepository[models.CalendarFeed]
	// FindByPerson returns the person's calendar, or gorm.ErrRecordNotFound.
	FindByPerson(ctx context.Context, personID string) (*models.CalendarFeed, error)
	// ClaimDue locks up to limit calendars due at now (FOR UPDATE SKIP LOCKED) and moves their
	// next sync to next, so another server skips them. Call it inside a transaction.
	ClaimDue(ctx context.Context, now, next time.Time, limit int) ([]models.CalendarFeed, error)
	// RecordSync stores how a sync went: on success the time and the events found, on failure
	// only the error code, so the last good sync stays visible.
	RecordSync(ctx context.Context, id string, at time.Time, eventsFound int, errorCode *string) error
	// DeleteByPerson removes the person's calendar for good: its link is a secret.
	DeleteByPerson(ctx context.Context, personID string) error
}

type calendarFeedRepository struct {
	abstractRepository[models.CalendarFeed]
}

// ProvideCalendarFeedRepository creates the calendar feed repository.
func ProvideCalendarFeedRepository(db *db.PostgresDB) CalendarFeedRepository {
	return &calendarFeedRepository{abstractRepository: abstractRepository[models.CalendarFeed]{db: db}}
}

func (r *calendarFeedRepository) FindByPerson(ctx context.Context, personID string) (*models.CalendarFeed, error) {
	var feed models.CalendarFeed
	if err := r.db.From(ctx).Where("person_id = ?", personID).First(&feed).Error; err != nil {
		return nil, err
	}
	return &feed, nil
}

func (r *calendarFeedRepository) ClaimDue(ctx context.Context, now, next time.Time, limit int) ([]models.CalendarFeed, error) {
	var feeds []models.CalendarFeed
	err := r.db.From(ctx).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate, Options: clause.LockingOptionsSkipLocked}).
		Where("next_sync_at <= ?", now).
		Order("next_sync_at ASC").Order("id ASC").
		Limit(limit).
		Find(&feeds).Error
	if err != nil || len(feeds) == 0 {
		return feeds, err
	}
	ids := make([]string, len(feeds))
	for i, f := range feeds {
		ids[i] = f.ID
	}
	return feeds, r.db.From(ctx).Model(&models.CalendarFeed{}).Where("id IN ?", ids).Update("next_sync_at", next).Error
}

func (r *calendarFeedRepository) RecordSync(ctx context.Context, id string, at time.Time, eventsFound int, errorCode *string) error {
	updates := map[string]any{"last_error": errorCode}
	if errorCode == nil {
		updates["last_synced_at"] = at
		updates["events_found"] = eventsFound
	}
	return r.db.From(ctx).Model(&models.CalendarFeed{}).Where("id = ?", id).Updates(updates).Error
}

func (r *calendarFeedRepository) DeleteByPerson(ctx context.Context, personID string) error {
	return r.db.From(ctx).Unscoped().Where("person_id = ?", personID).Delete(&models.CalendarFeed{}).Error
}
```

  - In `server/internal/services/absence_test.go`, give the mock the new method, just above `// stubPeople knows one person`:

```go
func (m *MockAbsenceRepository) ListCalendarLeave(ctx context.Context, personID string) ([]models.Absence, error) {
	args := m.Called(ctx, personID)
	return args.Get(0).([]models.Absence), args.Error(1)
}
```

- [ ] **Step 8: Run the tests to make sure they pass**

  Run: `cd server && go build ./... && go test ./internal/services ./internal/dtos`
  Expected: `ok` for both. The member-calendar schedule tests still pass, because their absences have no source and compare as `""`.

- [ ] **Step 9: Generate the migration**

  Run: `cd server && make migrate-generate name=calendar_feeds`. This needs the Atlas CLI and Docker; if either is missing, stop and say so.
  Expected: a new `cmd/migrations/sql/<timestamp>_calendar_feeds.sql` with exactly:

```sql
-- Modify "absences" table
ALTER TABLE "public"."absences" ADD COLUMN "source" text NOT NULL DEFAULT 'manual', ADD COLUMN "external_uid" text NULL;
-- Create index "idx_absences_person_external_uid" to table: "absences"
CREATE UNIQUE INDEX "idx_absences_person_external_uid" ON "public"."absences" ("person_id", "external_uid") WHERE ((external_uid IS NOT NULL) AND (deleted_at IS NULL));
-- Create "calendar_feeds" table
CREATE TABLE "public"."calendar_feeds" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "person_id" uuid NOT NULL,
  "url_ciphertext" bytea NOT NULL,
  "host" text NOT NULL,
  "last_synced_at" timestamptz NULL,
  "last_error" text NULL,
  "events_found" bigint NOT NULL DEFAULT 0,
  "next_sync_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_calendar_feeds_person" FOREIGN KEY ("person_id") REFERENCES "public"."people" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_calendar_feeds_deleted_at" to table: "calendar_feeds"
CREATE INDEX "idx_calendar_feeds_deleted_at" ON "public"."calendar_feeds" ("deleted_at");
-- Create index "idx_calendar_feeds_next_sync_at" to table: "calendar_feeds"
CREATE INDEX "idx_calendar_feeds_next_sync_at" ON "public"."calendar_feeds" ("next_sync_at");
-- Create index "idx_calendar_feeds_person_id" to table: "calendar_feeds"
CREATE UNIQUE INDEX "idx_calendar_feeds_person_id" ON "public"."calendar_feeds" ("person_id");
```

  Run: `cd server && make migrate-up`, against the local stack started with `make infra`.
  Expected: `Migrating to version <timestamp>` with 1 migration applied.

- [ ] **Step 10: Run the server checks, then commit**

```bash
cd server && go build ./... && go vet ./... && go test ./... && make lint && go test -tags=integration ./...
git add internal/constants/absence.go internal/models/absence.go internal/models/calendarfeed.go \
  internal/repositories/absence.go internal/repositories/calendarfeed.go internal/services/absence.go \
  internal/services/absence_test.go internal/dtos/absence.go internal/models/schedule.go internal/services/schedule.go \
  internal/dtos/schedule.go cmd/migrations/sql
git commit -m "feat(server): absence source and the calendar_feeds table" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Apply calendar leave in the availability service

The availability module owns `absences`, so it writes calendar leave, through the same `saveWithEvent` outbox as any absence. Approving calendar leave uses member-calendar's `POST /absences/{id}/approve` (`AvailabilityService.Approve`); nothing here changes it.

**Rule (from the spec):** when a sync finds an event moved or resized, its absence is updated and set back to `planned`, even if a planner had approved it.

**Files:**
- Modify: `server/internal/services/absence.go`
- Create: `server/internal/services/calendarleave_test.go`

**Interfaces:**
- Consumes: Task 5 (`Source`, `ExternalUID`, `ListCalendarLeave`), and the existing `saveWithEvent`, `absenceChange` and `events.Record`.
- Produces:
  - `services.CalendarLeave{UID string; FromDate, ToDate time.Time; HoursPerDay float64}`
  - `services.ApplyCalendarLeaveInput{PersonID string; From, To time.Time; Leave []CalendarLeave}`
  - `services.CalendarLeaveResult{Created, Updated, Deleted int}`
  - `AvailabilityService.ApplyCalendarLeave(ctx, ApplyCalendarLeaveInput) (CalendarLeaveResult, error)`. It checks no access: the calendar feed service does, and its runner acts for the system.

- [ ] **Step 1: Write the failing tests.** Create `server/internal/services/calendarleave_test.go`:

```go
package services

import (
	"context"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func calendarAbsence(uid string, from, to time.Time, hours float64, status string) models.Absence {
	person := otherPersonID
	return models.Absence{BaseModel: models.BaseModel{ID: "abs-" + uid}, PersonID: &person, FromDate: from, ToDate: to,
		HoursPerDay: hours, Type: constants.AbsenceTypeLeave, Status: status, Source: constants.AbsenceSourceCalendar,
		ExternalUID: &uid}
}

func TestAvailabilityService_ApplyCalendarLeave(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, time.October, d, 0, 0, 0, 0, time.UTC) }
	window := ApplyCalendarLeaveInput{PersonID: otherPersonID, From: day(13), To: day(13).AddDate(0, 0, 182)}
	// kept started before the window and is still in the calendar: it stays as it is.
	kept := calendarAbsence("kept", day(9), day(16), 8, constants.AbsenceStatusApproved)
	moved := calendarAbsence("moved", day(21), day(21), 8, constants.AbsenceStatusApproved)
	gone := calendarAbsence("gone", day(26), day(27), 8, constants.AbsenceStatusPlanned)
	past := calendarAbsence("past", time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.September, 2, 0, 0, 0, 0, time.UTC), 8, constants.AbsenceStatusApproved)

	repo := new(MockAbsenceRepository)
	repo.On("ListCalendarLeave", mock.Anything, otherPersonID).Return([]models.Absence{past, kept, moved, gone}, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
		return *a.ExternalUID == "new" && a.Status == constants.AbsenceStatusPlanned && a.Source == constants.AbsenceSourceCalendar &&
			a.Type == constants.AbsenceTypeLeave && a.HoursPerDay == 4
	})).Return(nil)
	repo.On("Save", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
		return a.ID == moved.ID && a.FromDate.Equal(day(22)) && a.Status == constants.AbsenceStatusPlanned
	})).Return(nil)
	repo.On("Delete", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool { return a.ID == gone.ID })).Return(nil)
	events := &recordedEvents{}
	people := stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: selfPersonID}}}
	service := ProvideAvailabilityService(repo, people, ProvideAuthzService(people), passthroughTx{}, events)

	window.Leave = []CalendarLeave{
		{UID: "kept", FromDate: day(9), ToDate: day(16), HoursPerDay: 8},
		{UID: "moved", FromDate: day(22), ToDate: day(22), HoursPerDay: 8},
		{UID: "new", FromDate: day(23), ToDate: day(23), HoursPerDay: 4},
	}
	result, err := service.ApplyCalendarLeave(context.Background(), window)

	require.NoError(t, err)
	assert.Equal(t, CalendarLeaveResult{Created: 1, Updated: 1, Deleted: 1}, result)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool { return a.ID == past.ID }))
	var types []string
	for _, e := range events.events {
		types = append(types, e.Type)
	}
	assert.Equal(t, []string{constants.EventAbsenceRemoved, constants.EventAbsenceAdded, constants.EventAbsenceAdded,
		constants.EventAbsenceRemoved}, types, "moved: removed + added; new: added; gone: removed")
}
```

- [ ] **Step 2: Run them to make sure they fail**

  Run: `cd server && go test ./internal/services -run ApplyCalendarLeave`
  Expected: FAIL to compile, with `undefined: ApplyCalendarLeaveInput`.

- [ ] **Step 3: Write the service code.** In `server/internal/services/absence.go`:
  - add `"slices"` to the imports
  - add `opApplyCalendarLeave = "apply_calendar_leave"` to the `op…` constants
  - add the types above `// AvailabilityService manages absences and holidays.`:

```go
// CalendarLeave is one out-of-office event as leave: dates inclusive, hours per day resolved.
type CalendarLeave struct {
	UID         string
	FromDate    time.Time
	ToDate      time.Time
	HoursPerDay float64
}

// ApplyCalendarLeaveInput is the leave read from one person's calendar for the dates From to To.
type ApplyCalendarLeaveInput struct {
	PersonID string
	From     time.Time
	To       time.Time
	Leave    []CalendarLeave
}

// CalendarLeaveResult counts what one sync changed.
type CalendarLeaveResult struct {
	Created int
	Updated int
	Deleted int
}
```

  - add these to the `AvailabilityService` interface, after `CreateHoliday`:

```go
	// ApplyCalendarLeave makes the person's calendar absences match in.Leave, in one transaction:
	// a new event creates planned leave, a changed one moves it and sets it back to planned, and
	// one that is gone deletes it when it touches in.From to in.To. Each change records its
	// events. Absences entered by hand are never touched. It checks no access: the calendar
	// feed service does, and its runner acts for the system.
	ApplyCalendarLeave(ctx context.Context, in ApplyCalendarLeaveInput) (CalendarLeaveResult, error)
```

  - add these functions just above `func checkAbsenceDates(`:

```go
func (s *availabilityService) ApplyCalendarLeave(ctx context.Context, in ApplyCalendarLeaveInput) (CalendarLeaveResult, error) {
	person, err := s.people.Get(ctx, in.PersonID)
	if err != nil {
		return CalendarLeaveResult{}, err
	}
	var result CalendarLeaveResult
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		result = CalendarLeaveResult{}
		existing, err := s.repo.ListCalendarLeave(ctx, in.PersonID)
		if err != nil {
			return err
		}
		byUID := make(map[string]models.Absence, len(existing))
		for _, a := range existing {
			if a.ExternalUID != nil {
				byUID[*a.ExternalUID] = a
			}
		}
		for _, leave := range in.Leave {
			old, known := byUID[leave.UID]
			delete(byUID, leave.UID)
			switch {
			case !known:
				err = s.createCalendarLeave(ctx, in.PersonID, person.Name, leave)
				result.Created++
			case !sameLeave(old, leave):
				err = s.moveCalendarLeave(ctx, old, person.Name, leave)
				result.Updated++
			}
			if err != nil {
				return err
			}
		}
		for _, gone := range goneInRange(byUID, in.From, in.To) {
			remove := func(ctx context.Context) error { return s.repo.Delete(ctx, &gone) }
			event := ChangeInput{Type: constants.EventAbsenceRemoved, Payload: absenceChange(&gone, person.Name)}
			if err := s.saveWithEvent(ctx, remove, event); err != nil {
				return err
			}
			result.Deleted++
		}
		return nil
	})
	if err != nil {
		return CalendarLeaveResult{}, asAppError(err, constants.MsgAbsenceCreateFailed, opApplyCalendarLeave, resourceAbsence)
	}
	return result, nil
}

// createCalendarLeave saves a new calendar event as planned leave, with its event.
func (s *availabilityService) createCalendarLeave(ctx context.Context, personID, name string, leave CalendarLeave) error {
	uid := leave.UID
	absence := &models.Absence{BaseModel: models.NewBaseModel(), PersonID: &personID, FromDate: leave.FromDate,
		ToDate: leave.ToDate, HoursPerDay: leave.HoursPerDay, Type: constants.AbsenceTypeLeave,
		Status: constants.AbsenceStatusPlanned, Source: constants.AbsenceSourceCalendar, ExternalUID: &uid}
	create := func(ctx context.Context) error { return s.repo.Create(ctx, absence) }
	return s.saveWithEvent(ctx, create, ChangeInput{Type: constants.EventAbsenceAdded, Payload: absenceChange(absence, name)})
}

// moveCalendarLeave moves calendar leave to the event's new dates and hours, back to planned.
// It records the old dates as removed and the new ones as added, so the re-plan sees both weeks.
func (s *availabilityService) moveCalendarLeave(ctx context.Context, old models.Absence, name string, leave CalendarLeave) error {
	if err := s.events.Record(ctx, ChangeInput{Type: constants.EventAbsenceRemoved, Payload: absenceChange(&old, name)}); err != nil {
		return err
	}
	moved := old
	moved.FromDate, moved.ToDate, moved.HoursPerDay = leave.FromDate, leave.ToDate, leave.HoursPerDay
	moved.Status = constants.AbsenceStatusPlanned
	save := func(ctx context.Context) error { return s.repo.Save(ctx, &moved) }
	return s.saveWithEvent(ctx, save, ChangeInput{Type: constants.EventAbsenceAdded, Payload: absenceChange(&moved, name)})
}

// sameLeave reports calendar leave that already matches the event (status is not compared).
func sameLeave(a models.Absence, leave CalendarLeave) bool {
	return sameDate(a.FromDate, leave.FromDate) && sameDate(a.ToDate, leave.ToDate) &&
		a.HoursPerDay-leave.HoursPerDay < hoursEpsilon && leave.HoursPerDay-a.HoursPerDay < hoursEpsilon
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// goneInRange is the calendar leave whose event is gone and that touches from to to, oldest first.
func goneInRange(gone map[string]models.Absence, from, to time.Time) []models.Absence {
	var out []models.Absence
	for _, a := range gone {
		if !a.ToDate.Before(from) && !a.FromDate.After(to) {
			out = append(out, a)
		}
	}
	slices.SortFunc(out, func(a, b models.Absence) int { return a.FromDate.Compare(b.FromDate) })
	return out
}
```

  The nested `saveWithEvent` calls run inside the outer `tx.Do`: a nested `Do` reuses the open transaction, so one sync commits or rolls back as a whole.

- [ ] **Step 4: Run the tests to make sure they pass**

  Run: `cd server && go test ./internal/services -run 'AvailabilityService' -v`
  Expected: `--- PASS: TestAvailabilityService_ApplyCalendarLeave`, and the older availability tests (member-calendar's `Approve` included) still pass.

- [ ] **Step 5: Run the server checks, then commit**

```bash
cd server && go build ./... && go vet ./... && go test ./... && make lint
git add internal/services/absence.go internal/services/calendarleave_test.go
git commit -m "feat(server): apply calendar leave with change events" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 7: The calendar feed service

**Files:**
- Modify: `server/internal/constants/error_codes.go`
- Modify: `server/internal/constants/messages.go`
- Create: `server/internal/services/calendarfeed.go`
- Create: `server/internal/services/calendarfeed_test.go`

**Interfaces:**
- Consumes:
  - Task 2 settings
  - Task 3: `calendar.Parse`, `calendar.LeaveRules`
  - Task 4: `calendar.Fetcher`, `CheckURL`, `NewSealer`, `Sealer`, `ErrBadKey`, and the errors
  - Task 5: `CalendarFeedRepository`
  - Task 6: `AvailabilityService.ApplyCalendarLeave`
  - existing: `PeopleService.Get`, `AuthzService.RequireSelfOr`, `CanManageAbsences`, `db.TxManager`, `clock.Clock`
- Produces:
  - `services.ConnectFeedInput{PersonID, URL string}`
  - `services.CalendarFeedService` with `Get`, `Connect`, `Disconnect`, `SyncNow`, `RunOnce`, `Run`
  - `services.ProvideCalendarFeedService(feeds, people, availability, authz, tx, clk, cfg, fetcher) (CalendarFeedService, error)`
  - the error codes `constants.CalendarSyncDisabled`, `CalendarFeedNotFound`, `CalendarFeedHostNotAllowed`, `CalendarFeedUnreachable`, `CalendarFeedNotICS` and `CalendarFeedTooLarge`, and the `MsgCalendarFeed…` messages

Rules the code follows:
- Connect fetches **before** its transaction, so a slow calendar holds no database locks.
- A failed sync keeps the absences as they are and stores only the error code.
- The runner claims calendars in a short transaction, then syncs each one outside it.

- [ ] **Step 1: Write the failing tests.** Create `server/internal/services/calendarfeed_test.go`. `testFeedKey` is a fixed 32-byte test key, never a real one; `G101` is excluded in `.golangci.yml`.

```go
package services

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/integration/calendar"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	googleLink  = "https://calendar.google.com/calendar/ical/a%40b/private-secret/basic.ics"
	feedRowID   = "0192f3a4-0000-7000-8000-0000000000f1"
	allDayOOO   = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:ooo\r\nSUMMARY:OOO\r\nDTSTART;VALUE=DATE:20261019\r\nDTEND;VALUE=DATE:20261021\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	halfDayOOO  = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:half\r\nSUMMARY:Out of office\r\nDTSTART:20261019T130000Z\r\nDTEND:20261019T163000Z\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	testFeedKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=" // 32 bytes, for tests only
)

// syncNow is a Wednesday, 13 Oct 2026 09:00 UTC.
var syncNow = time.Date(2026, time.October, 13, 9, 0, 0, 0, time.UTC)

// fakeFeeds keeps calendar feeds in memory.
type fakeFeeds struct {
	repositories.CalendarFeedRepository
	byPerson map[string]*models.CalendarFeed
	claimed  bool
}

func (f *fakeFeeds) FindByPerson(_ context.Context, personID string) (*models.CalendarFeed, error) {
	if feed, ok := f.byPerson[personID]; ok {
		copied := *feed
		return &copied, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeFeeds) Save(_ context.Context, feed *models.CalendarFeed) error {
	copied := *feed
	f.byPerson[feed.PersonID] = &copied
	return nil
}

func (f *fakeFeeds) RecordSync(_ context.Context, id string, at time.Time, eventsFound int, errorCode *string) error {
	for _, feed := range f.byPerson {
		if feed.ID == id {
			feed.LastError = errorCode
			if errorCode == nil {
				feed.LastSyncedAt, feed.EventsFound = &at, eventsFound
			}
		}
	}
	return nil
}

func (f *fakeFeeds) ClaimDue(_ context.Context, now, next time.Time, _ int) ([]models.CalendarFeed, error) {
	f.claimed = true
	var due []models.CalendarFeed
	for _, feed := range f.byPerson {
		if !feed.NextSyncAt.After(now) {
			feed.NextSyncAt = next
			due = append(due, *feed)
		}
	}
	return due, nil
}

func (f *fakeFeeds) DeleteByPerson(_ context.Context, personID string) error {
	delete(f.byPerson, personID)
	return nil
}

// fakeFetcher returns body for any link and remembers the links it was asked for.
type fakeFetcher struct {
	body  string
	err   error
	links []string
}

func (f *fakeFetcher) Fetch(_ context.Context, link string) ([]byte, error) {
	f.links = append(f.links, link)
	return []byte(f.body), f.err
}

// fakeLeave records the leave a sync applies.
type fakeLeave struct {
	AvailabilityService
	applied []ApplyCalendarLeaveInput
}

func (f *fakeLeave) ApplyCalendarLeave(_ context.Context, in ApplyCalendarLeaveInput) (CalendarLeaveResult, error) {
	f.applied = append(f.applied, in)
	return CalendarLeaveResult{Created: len(in.Leave)}, nil
}

// hoursPeople returns a person with weeklyHours for any id and links the specialist to selfPersonID.
type hoursPeople struct {
	stubPeople
	weeklyHours float64
}

func (p hoursPeople) Get(_ context.Context, id string) (*models.Person, error) {
	return &models.Person{BaseModel: models.BaseModel{ID: id}, Name: "Alex", WeeklyHours: p.weeklyHours}, nil
}

// fullTime is the weekly hours of the people in these tests, unless a test says otherwise.
const fullTime = 40

func feedConfig(key string) *config.Config {
	return &config.Config{CalendarFeedKey: key, CalendarAllowedHosts: []string{"calendar.google.com"},
		CalendarSyncInterval: 15 * time.Minute, CalendarSyncBatchSize: 20, CalendarSyncWeeks: 26,
		CalendarLeaveKeywords: []string{"OOO", "Out of office"},
		WorkingWeekdays:       []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}}
}

func newFeedService(t *testing.T, key string, feeds *fakeFeeds, fetcher *fakeFetcher, leave *fakeLeave) CalendarFeedService {
	t.Helper()
	return newFeedServiceFor(t, key, feeds, fetcher, leave, fullTime)
}

func newFeedServiceFor(t *testing.T, key string, feeds *fakeFeeds, fetcher *fakeFetcher, leave *fakeLeave, weeklyHours float64) CalendarFeedService {
	t.Helper()
	people := hoursPeople{stubPeople: stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: selfPersonID}}}, weeklyHours: weeklyHours}
	s, err := ProvideCalendarFeedService(feeds, people, leave, ProvideAuthzService(people), passthroughTx{},
		clock.NewFixed(syncNow), feedConfig(key), fetcher)
	require.NoError(t, err)
	return s
}

func TestProvideCalendarFeedService_BadKey(t *testing.T) {
	_, err := ProvideCalendarFeedService(nil, nil, nil, nil, nil, clock.NewFixed(syncNow),
		feedConfig(base64.StdEncoding.EncodeToString([]byte("short"))), nil)
	require.ErrorIs(t, err, calendar.ErrBadKey)
}

func TestCalendarFeedService_Connect(t *testing.T) {
	specialist := asRole("kc-specialist", constants.RoleSpecialist)
	planner := asRole("kc-planner", constants.RolePlanner)
	tests := []struct {
		name      string
		key       string
		ctx       context.Context
		personID  string
		link      string
		body      string
		fetchErr  error
		wantCode  string
		wantHTTP  int
		wantHours float64
	}{
		{name: "a person connects their own calendar: all-day OOO counts full days", key: testFeedKey, ctx: specialist,
			personID: selfPersonID, link: googleLink, body: allDayOOO, wantHours: 8},
		{name: "a timed event counts its hours", key: testFeedKey, ctx: planner, personID: otherPersonID, link: googleLink,
			body: halfDayOOO, wantHours: 3.5},
		{name: "off without a key", ctx: planner, personID: otherPersonID, link: googleLink,
			wantCode: constants.CalendarSyncDisabled, wantHTTP: http.StatusServiceUnavailable},
		{name: "someone else's calendar", key: testFeedKey, ctx: specialist, personID: otherPersonID, link: googleLink,
			wantHTTP: http.StatusForbidden},
		{name: "a host outside the list is refused before any fetch", key: testFeedKey, ctx: planner, personID: otherPersonID,
			link: "https://evil.example/basic.ics", wantCode: constants.CalendarFeedHostNotAllowed, wantHTTP: http.StatusBadRequest},
		{name: "http is refused", key: testFeedKey, ctx: planner, personID: otherPersonID,
			link: strings.Replace(googleLink, "https", "http", 1), wantCode: constants.CalendarFeedHostNotAllowed, wantHTTP: http.StatusBadRequest},
		{name: "a link that is not a calendar", key: testFeedKey, ctx: planner, personID: otherPersonID, link: googleLink,
			body: "<html>Sign in</html>", wantCode: constants.CalendarFeedNotICS, wantHTTP: http.StatusBadRequest},
		{name: "a calendar that cannot be fetched", key: testFeedKey, ctx: planner, personID: otherPersonID, link: googleLink,
			fetchErr: calendar.ErrUnreachable, wantCode: constants.CalendarFeedUnreachable, wantHTTP: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
			fetcher := &fakeFetcher{body: tt.body, err: tt.fetchErr}
			leave := &fakeLeave{}
			feed, err := newFeedService(t, tt.key, feeds, fetcher, leave).Connect(tt.ctx, ConnectFeedInput{PersonID: tt.personID, URL: tt.link})

			if tt.wantHTTP != 0 {
				if tt.wantCode != "" {
					requireAppError(t, err, tt.wantCode, tt.wantHTTP)
				} else {
					requireStatus(t, err, tt.wantHTTP)
				}
				assert.NotContains(t, err.Error(), "private-secret", "an error must never hold the link")
				assert.Empty(t, feeds.byPerson, "nothing is saved")
				assert.Empty(t, leave.applied)
				if tt.wantCode == constants.CalendarFeedHostNotAllowed {
					assert.Empty(t, fetcher.links, "a refused link is never fetched")
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "calendar.google.com", feed.Host)
			assert.NotContains(t, string(feed.URLCiphertext), "private-secret", "the link is stored encrypted")
			assert.Equal(t, syncNow.Add(15*time.Minute), feed.NextSyncAt)
			require.NotNil(t, feed.LastSyncedAt)
			assert.Equal(t, 1, feed.EventsFound)
			require.Len(t, leave.applied, 1)
			require.Len(t, leave.applied[0].Leave, 1)
			assert.InDelta(t, tt.wantHours, leave.applied[0].Leave[0].HoursPerDay, 1e-9)
			assert.Equal(t, time.Date(2026, time.October, 13, 0, 0, 0, 0, time.UTC), leave.applied[0].From)
			assert.Equal(t, time.Date(2027, time.April, 13, 0, 0, 0, 0, time.UTC), leave.applied[0].To)
		})
	}
}

func TestCalendarFeedService_SyncKeepsAbsencesWhenTheFetchFails(t *testing.T) {
	feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
	fetcher := &fakeFetcher{body: allDayOOO}
	leave := &fakeLeave{}
	service := newFeedService(t, testFeedKey, feeds, fetcher, leave)
	planner := asRole("kc-planner", constants.RolePlanner)
	_, err := service.Connect(planner, ConnectFeedInput{PersonID: otherPersonID, URL: googleLink})
	require.NoError(t, err)

	fetcher.err = calendar.ErrUnreachable
	feed, err := service.SyncNow(planner, otherPersonID)

	require.NoError(t, err)
	require.NotNil(t, feed.LastError)
	assert.Equal(t, constants.CalendarFeedUnreachable, *feed.LastError)
	assert.Equal(t, 1, feed.EventsFound, "the last good sync stays")
	assert.Len(t, leave.applied, 1, "a failed fetch applies nothing")
	assert.Equal(t, []string{googleLink, googleLink}, fetcher.links, "the stored link opens back to the same link")
}

func TestCalendarFeedService_RunOnce(t *testing.T) {
	t.Run("does nothing while off", func(t *testing.T) {
		feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
		n, err := newFeedService(t, "", feeds, &fakeFetcher{}, &fakeLeave{}).RunOnce(context.Background())
		require.NoError(t, err)
		assert.Zero(t, n)
		assert.False(t, feeds.claimed)
	})

	t.Run("syncs the calendars that are due, for the system", func(t *testing.T) {
		feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
		fetcher := &fakeFetcher{body: allDayOOO}
		leave := &fakeLeave{}
		service := newFeedService(t, testFeedKey, feeds, fetcher, leave)
		_, err := service.Connect(asRole("kc-planner", constants.RolePlanner), ConnectFeedInput{PersonID: otherPersonID, URL: googleLink})
		require.NoError(t, err)
		feeds.byPerson[otherPersonID].ID = feedRowID
		feeds.byPerson[otherPersonID].NextSyncAt = syncNow

		n, err := service.RunOnce(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Len(t, leave.applied, 2, "connect, then the runner")
		assert.Equal(t, syncNow.Add(15*time.Minute), feeds.byPerson[otherPersonID].NextSyncAt)
	})
}

func TestCalendarFeedService_GetAndDisconnect(t *testing.T) {
	feeds := &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}}
	service := newFeedService(t, testFeedKey, feeds, &fakeFetcher{body: allDayOOO}, &fakeLeave{})
	specialist := asRole("kc-specialist", constants.RoleSpecialist)

	_, err := service.Get(specialist, selfPersonID)
	requireAppError(t, err, constants.CalendarFeedNotFound, http.StatusNotFound)

	_, err = service.Connect(specialist, ConnectFeedInput{PersonID: selfPersonID, URL: googleLink})
	require.NoError(t, err)
	_, err = service.Get(asRole("kc-lead", constants.RoleLead), selfPersonID)
	requireStatus(t, err, http.StatusForbidden)

	require.NoError(t, service.Disconnect(specialist, selfPersonID))
	assert.Empty(t, feeds.byPerson)
}

func TestCalendarFeedService_LeaveHours(t *testing.T) {
	const oddLength = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:odd\r\nSUMMARY:OOO dentist\r\nDTSTART:20261019T090000Z\r\nDTEND:20261019T113500Z\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	tests := []struct {
		name        string
		weeklyHours float64
		body        string
		wantHours   []float64
	}{
		{name: "an odd length is rounded to two decimals, as the database stores it", weeklyHours: fullTime, body: oddLength,
			wantHours: []float64{2.58}},
		{name: "a person with no weekly hours gets no zero-hour leave", weeklyHours: 0, body: allDayOOO, wantHours: []float64{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leave := &fakeLeave{}
			service := newFeedServiceFor(t, testFeedKey, &fakeFeeds{byPerson: map[string]*models.CalendarFeed{}},
				&fakeFetcher{body: tt.body}, leave, tt.weeklyHours)
			_, err := service.Connect(asRole("kc-planner", constants.RolePlanner), ConnectFeedInput{PersonID: otherPersonID, URL: googleLink})
			require.NoError(t, err)
			require.Len(t, leave.applied, 1)
			hours := []float64{}
			for _, l := range leave.applied[0].Leave {
				hours = append(hours, l.HoursPerDay)
			}
			assert.Equal(t, tt.wantHours, hours)
		})
	}
}
```

- [ ] **Step 2: Run them to make sure they fail**

  Run: `cd server && go test ./internal/services -run 'CalendarFeed'`
  Expected: FAIL to compile, with `undefined: ProvideCalendarFeedService`.

- [ ] **Step 3: Add the error codes.** In `server/internal/constants/error_codes.go`, after `AbsenceDatesInvalid`:

```go
	// Calendar feed errors. The link is a secret, so a problem is only ever stored as one of these.
	CalendarSyncDisabled       = "CALENDAR_SYNC_DISABLED"
	CalendarFeedNotFound       = "CALENDAR_FEED_NOT_FOUND"
	CalendarFeedHostNotAllowed = "CALENDAR_FEED_HOST_NOT_ALLOWED"
	CalendarFeedUnreachable    = "CALENDAR_FEED_UNREACHABLE"
	CalendarFeedNotICS         = "CALENDAR_FEED_NOT_ICS"
	CalendarFeedTooLarge       = "CALENDAR_FEED_TOO_LARGE"
```

- [ ] **Step 4: Add the messages.** In `server/internal/constants/messages.go`, after the absence block:

```go
// Calendar feed messages. None of them may include the link.
const (
	MsgCalendarFeedRetrieved      = "Calendar retrieved"
	MsgCalendarFeedConnected      = "Calendar connected"
	MsgCalendarFeedDisconnected   = "Calendar disconnected"
	MsgCalendarFeedSynced         = "Calendar synced"
	MsgCalendarSyncDisabled       = "Calendar sync is turned off on this server"
	MsgCalendarFeedNotFound       = "No calendar is connected"
	MsgCalendarFeedHostNotAllowed = "Use the secret iCal (https) address from Google Calendar or Outlook"
	MsgCalendarFeedUnreachable    = "The calendar could not be fetched"
	MsgCalendarFeedNotICS         = "The link does not return an iCal calendar"
	MsgCalendarFeedTooLarge       = "The calendar is larger than the limit"
	MsgCalendarFeedSaveFailed     = "Could not save the calendar"
)
```

- [ ] **Step 5: Write the service.** Create `server/internal/services/calendarfeed.go`:

```go
package services

import (
	"context"
	stderrors "errors"
	"math"
	"net/http"
	"strings"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/integration/calendar"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	opGetCalendarFeed        = "get_calendar_feed"
	opConnectCalendarFeed    = "connect_calendar_feed"
	opDisconnectCalendarFeed = "disconnect_calendar_feed"
	opSyncCalendarFeed       = "sync_calendar_feed"
	resourceCalendarFeed     = "calendar_feed"
)

// hoursPerDayScale rounds leave hours to two decimals, as absences store them (numeric(4,2)),
// so an unchanged event never looks moved after a round trip through the database.
const hoursPerDayScale = 100

// daysPerWeek turns CALENDAR_SYNC_WEEKS into days.
const daysPerWeek = 7

// ConnectFeedInput is a person's secret calendar link.
type ConnectFeedInput struct {
	PersonID string
	URL      string
}

// CalendarFeedService connects people's calendars and turns their out-of-office events into
// planned leave. A person may manage their own calendar; planners anyone's.
type CalendarFeedService interface {
	Get(ctx context.Context, personID string) (*models.CalendarFeed, error)
	// Connect checks the link, syncs it once, and stores it encrypted, replacing an earlier one.
	Connect(ctx context.Context, in ConnectFeedInput) (*models.CalendarFeed, error)
	Disconnect(ctx context.Context, personID string) error
	// SyncNow syncs the person's calendar at once and returns its status.
	SyncNow(ctx context.Context, personID string) (*models.CalendarFeed, error)
	// RunOnce claims the calendars that are due and syncs them; it returns how many it claimed.
	RunOnce(ctx context.Context) (int, error)
	// Run calls RunOnce every CALENDAR_SYNC_INTERVAL until ctx ends. It returns at once while
	// calendar sync is off.
	Run(ctx context.Context)
}

type calendarFeedService struct {
	feeds        repositories.CalendarFeedRepository
	people       PeopleService
	availability AvailabilityService
	authz        AuthzService
	tx           db.TxManager
	clock        clock.Clock
	cfg          *config.Config
	fetcher      calendar.Fetcher
	// sealer is nil while CALENDAR_FEED_KEY is empty: the feature is off.
	sealer *calendar.Sealer
}

// ProvideCalendarFeedService creates the calendar feed service. An empty CALENDAR_FEED_KEY turns
// the feature off; a malformed one stops the server at startup.
func ProvideCalendarFeedService(feeds repositories.CalendarFeedRepository, people PeopleService,
	availability AvailabilityService, authz AuthzService, tx db.TxManager, clk clock.Clock, cfg *config.Config,
	fetcher calendar.Fetcher,
) (CalendarFeedService, error) {
	s := &calendarFeedService{feeds: feeds, people: people, availability: availability, authz: authz, tx: tx,
		clock: clk, cfg: cfg, fetcher: fetcher}
	if cfg.CalendarFeedKey == "" {
		return s, nil
	}
	sealer, err := calendar.NewSealer(cfg.CalendarFeedKey)
	if err != nil {
		return nil, err
	}
	s.sealer = sealer
	return s, nil
}

func (s *calendarFeedService) Get(ctx context.Context, personID string) (*models.CalendarFeed, error) {
	if err := s.check(ctx, personID, opGetCalendarFeed); err != nil {
		return nil, err
	}
	return s.find(ctx, personID, opGetCalendarFeed)
}

func (s *calendarFeedService) Connect(ctx context.Context, in ConnectFeedInput) (*models.CalendarFeed, error) {
	if err := s.check(ctx, in.PersonID, opConnectCalendarFeed); err != nil {
		return nil, err
	}
	person, err := s.people.Get(ctx, in.PersonID)
	if err != nil {
		return nil, err
	}
	link := strings.TrimSpace(in.URL)
	host, err := calendar.CheckURL(link, s.cfg.CalendarAllowedHosts)
	if err != nil {
		return nil, feedError(err, opConnectCalendarFeed)
	}
	// Fetch before the transaction: a slow calendar must not hold database locks.
	leave, found, err := s.read(ctx, link, person)
	if err != nil {
		return nil, feedError(err, opConnectCalendarFeed)
	}
	sealed, err := s.sealer.Seal(link)
	if err != nil {
		return nil, errors.InternalError(constants.MsgCalendarFeedSaveFailed, err).
			WithOperation(opConnectCalendarFeed).WithResource(resourceCalendarFeed)
	}
	now := s.clock.Now().UTC()
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		feed, err := s.feeds.FindByPerson(ctx, in.PersonID)
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			feed, err = &models.CalendarFeed{BaseModel: models.NewBaseModel(), PersonID: in.PersonID}, nil
		}
		if err != nil {
			return err
		}
		feed.URLCiphertext, feed.Host, feed.NextSyncAt = sealed, host, now.Add(s.cfg.CalendarSyncInterval)
		if err := s.feeds.Save(ctx, feed); err != nil {
			return err
		}
		if err := s.apply(ctx, person.ID, leave); err != nil {
			return err
		}
		return s.feeds.RecordSync(ctx, feed.ID, now, found, nil)
	})
	if err != nil {
		return nil, asAppError(err, constants.MsgCalendarFeedSaveFailed, opConnectCalendarFeed, resourceCalendarFeed)
	}
	return s.find(ctx, in.PersonID, opConnectCalendarFeed)
}

func (s *calendarFeedService) Disconnect(ctx context.Context, personID string) error {
	if err := s.check(ctx, personID, opDisconnectCalendarFeed); err != nil {
		return err
	}
	if _, err := s.find(ctx, personID, opDisconnectCalendarFeed); err != nil {
		return err
	}
	if err := s.feeds.DeleteByPerson(ctx, personID); err != nil {
		return errors.DatabaseError(constants.MsgCalendarFeedSaveFailed, err).
			WithOperation(opDisconnectCalendarFeed).WithResource(resourceCalendarFeed)
	}
	return nil
}

func (s *calendarFeedService) SyncNow(ctx context.Context, personID string) (*models.CalendarFeed, error) {
	if err := s.check(ctx, personID, opSyncCalendarFeed); err != nil {
		return nil, err
	}
	feed, err := s.find(ctx, personID, opSyncCalendarFeed)
	if err != nil {
		return nil, err
	}
	if err := s.sync(ctx, feed); err != nil {
		return nil, asAppError(err, constants.MsgCalendarFeedSaveFailed, opSyncCalendarFeed, resourceCalendarFeed)
	}
	return s.find(ctx, personID, opSyncCalendarFeed)
}

func (s *calendarFeedService) RunOnce(ctx context.Context) (int, error) {
	if s.sealer == nil {
		return 0, nil
	}
	now := s.clock.Now().UTC()
	var claimed []models.CalendarFeed
	err := s.tx.Do(ctx, func(ctx context.Context) error {
		feeds, err := s.feeds.ClaimDue(ctx, now, now.Add(s.cfg.CalendarSyncInterval), s.cfg.CalendarSyncBatchSize)
		claimed = feeds
		return err
	})
	if err != nil {
		return 0, err
	}
	var errs []error
	for i := range claimed {
		if err := s.sync(ctx, &claimed[i]); err != nil {
			errs = append(errs, err)
		}
	}
	return len(claimed), stderrors.Join(errs...)
}

func (s *calendarFeedService) Run(ctx context.Context) {
	if s.sealer == nil {
		return
	}
	ticker := time.NewTicker(s.cfg.CalendarSyncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Errors never hold a calendar link, so they are safe to log.
			if n, err := s.RunOnce(ctx); err != nil {
				logger.Log.Warn("Calendar sync failed", zap.Int("calendars", n), zap.Error(err))
			}
		}
	}
}

// check answers 503 while the feature is off, then lets the person or a planner through.
func (s *calendarFeedService) check(ctx context.Context, personID, op string) error {
	if s.sealer == nil {
		return errors.NewAppError(constants.CalendarSyncDisabled, constants.MsgCalendarSyncDisabled, errors.ErrorTypeExternal,
			http.StatusServiceUnavailable).WithOperation(op).WithResource(resourceCalendarFeed)
	}
	_, err := s.authz.RequireSelfOr(ctx, personID, CanManageAbsences)
	return err
}

func (s *calendarFeedService) find(ctx context.Context, personID, op string) (*models.CalendarFeed, error) {
	feed, err := s.feeds.FindByPerson(ctx, personID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WrapError(err, constants.CalendarFeedNotFound, constants.MsgCalendarFeedNotFound,
				errors.ErrorTypeNotFound, http.StatusNotFound).WithOperation(op).WithResource(resourceCalendarFeed)
		}
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceCalendarFeed)
	}
	return feed, nil
}

// sync reads one stored calendar and applies its leave. A calendar that cannot be read keeps
// its absences as they are and stores the error code; only database errors are returned.
func (s *calendarFeedService) sync(ctx context.Context, feed *models.CalendarFeed) error {
	person, err := s.people.Get(ctx, feed.PersonID)
	if err != nil {
		return err
	}
	now := s.clock.Now().UTC()
	link, err := s.sealer.Open(feed.URLCiphertext)
	if err != nil {
		return s.recordFailure(ctx, feed.ID, now, calendar.ErrUnreachable)
	}
	leave, found, err := s.read(ctx, link, person)
	if err != nil {
		return s.recordFailure(ctx, feed.ID, now, err)
	}
	return s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.apply(ctx, person.ID, leave); err != nil {
			return err
		}
		return s.feeds.RecordSync(ctx, feed.ID, now, found, nil)
	})
}

func (s *calendarFeedService) recordFailure(ctx context.Context, feedID string, at time.Time, cause error) error {
	code, _ := feedErrorCode(cause)
	return s.feeds.RecordSync(ctx, feedID, at, 0, &code)
}

// read fetches a calendar and returns its out-of-office events from today to
// CALENDAR_SYNC_WEEKS ahead as leave, and how many events it found.
func (s *calendarFeedService) read(ctx context.Context, link string, person *models.Person) ([]CalendarLeave, int, error) {
	data, err := s.fetcher.Fetch(ctx, link)
	if err != nil {
		return nil, 0, err
	}
	from, to := s.window()
	events, err := calendar.Parse(data, calendar.LeaveRules{Keywords: s.cfg.CalendarLeaveKeywords, From: from, To: to},
		s.clock.Location())
	if err != nil {
		return nil, 0, err
	}
	daily := dailyHours(person.WeeklyHours, s.cfg.WorkingWeekdays)
	leave := make([]CalendarLeave, 0, len(events))
	for _, e := range events {
		hours := daily
		if e.HoursPerDay != nil {
			hours = min(*e.HoursPerDay, daily)
		}
		if hours = math.Round(hours*hoursPerDayScale) / hoursPerDayScale; hours <= 0 {
			continue
		}
		leave = append(leave, CalendarLeave{UID: e.UID, FromDate: e.From, ToDate: e.To, HoursPerDay: hours})
	}
	return leave, len(events), nil
}

func (s *calendarFeedService) apply(ctx context.Context, personID string, leave []CalendarLeave) error {
	from, to := s.window()
	_, err := s.availability.ApplyCalendarLeave(ctx, ApplyCalendarLeaveInput{PersonID: personID, From: from, To: to, Leave: leave})
	return err
}

// window is today, in the organization's time zone, to CALENDAR_SYNC_WEEKS ahead, as UTC dates.
func (s *calendarFeedService) window() (time.Time, time.Time) {
	now := s.clock.Now().In(s.clock.Location())
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return from, from.AddDate(0, 0, s.cfg.CalendarSyncWeeks*daysPerWeek)
}

// dailyHours is a person's working day: weekly hours spread over the working weekdays.
func dailyHours(weeklyHours float64, workingDays []time.Weekday) float64 {
	if len(workingDays) == 0 {
		return 0
	}
	return weeklyHours / float64(len(workingDays))
}

// feedError turns a fetch or parse failure into a 400 with its code. Its message never holds the link.
func feedError(err error, op string) error {
	code, msg := feedErrorCode(err)
	return errors.NewAppError(code, msg, errors.ErrorTypeValidation, http.StatusBadRequest).
		WithOperation(op).WithResource(resourceCalendarFeed)
}

func feedErrorCode(err error) (string, string) {
	switch {
	case stderrors.Is(err, calendar.ErrHostNotAllowed):
		return constants.CalendarFeedHostNotAllowed, constants.MsgCalendarFeedHostNotAllowed
	case stderrors.Is(err, calendar.ErrTooLarge):
		return constants.CalendarFeedTooLarge, constants.MsgCalendarFeedTooLarge
	case stderrors.Is(err, calendar.ErrNotICS):
		return constants.CalendarFeedNotICS, constants.MsgCalendarFeedNotICS
	default:
		return constants.CalendarFeedUnreachable, constants.MsgCalendarFeedUnreachable
	}
}
```

- [ ] **Step 6: Run the tests to make sure they pass**

  Run: `cd server && go test ./internal/services -run 'CalendarFeed' -v`
  Expected: `TestProvideCalendarFeedService_BadKey`, `TestCalendarFeedService_Connect` (8 cases), `TestCalendarFeedService_SyncKeepsAbsencesWhenTheFetchFails`, `TestCalendarFeedService_RunOnce`, `TestCalendarFeedService_GetAndDisconnect` and `TestCalendarFeedService_LeaveHours` all pass.

- [ ] **Step 7: Run the server checks, then commit**

```bash
cd server && go build ./... && go vet ./... && go test ./... && make lint
git add internal/constants/error_codes.go internal/constants/messages.go internal/services/calendarfeed.go \
  internal/services/calendarfeed_test.go
git commit -m "feat(server): calendar feed service turns out-of-office events into planned leave" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 8: Prove a sync on a real database

**Files:**
- Create: `server/internal/services/calendarfeed_integration_test.go`

**Interfaces:**
- Consumes: Tasks 5–7, `dbtest.NewPostgres`, `dbtest.ApplyMigrations`, the real repositories, and `ProvideEventService` as the recorder. From `calendarfeed_test.go` it uses `fakeFetcher`, `hoursPeople`, `feedConfig`, `googleLink`, `testFeedKey`, `syncNow` and `fullTime`.
- Produces: nothing new. It pins the partial unique index, the change events, and the rule that manual absences are never touched.

- [ ] **Step 1: Write the test.** Create `server/internal/services/calendarfeed_integration_test.go`:

```go
//go:build integration

package services

import (
	"testing"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sharedOffsite = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:offsite@google.com\r\nSUMMARY:OOO offsite\r\nDTSTART;VALUE=DATE:20261019\r\nDTEND;VALUE=DATE:20261021\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	movedOffsite  = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:offsite@google.com\r\nSUMMARY:OOO offsite\r\nDTSTART;VALUE=DATE:20261026\r\nDTEND;VALUE=DATE:20261027\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	emptyCalendar = "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"
)

func TestCalendarFeedService_Sync_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	alex := models.Person{BaseModel: models.NewBaseModel(), Name: "Alex", Email: "alex@example.test", Role: constants.RoleSpecialist,
		WeeklyHours: 40, Active: true}
	bao := models.Person{BaseModel: models.NewBaseModel(), Name: "Bao", Email: "bao@example.test", Role: constants.RoleSpecialist,
		WeeklyHours: 40, Active: true}
	require.NoError(t, pdb.Create(&alex).Error)
	require.NoError(t, pdb.Create(&bao).Error)
	manual := models.Absence{BaseModel: models.NewBaseModel(), PersonID: &alex.ID, FromDate: syncNow.AddDate(0, 0, 14),
		ToDate: syncNow.AddDate(0, 0, 14), HoursPerDay: 8, Type: constants.AbsenceTypeTraining, Status: constants.AbsenceStatusApproved,
		Source: constants.AbsenceSourceManual}
	require.NoError(t, pdb.Create(&manual).Error)

	people := hoursPeople{weeklyHours: fullTime}
	authz := ProvideAuthzService(people)
	tx := db.ProvideTxManager(pdb)
	recorder := ProvideEventService(repositories.ProvideChangeEventRepository(pdb), repositories.ProvideProjectRepository(pdb),
		authz, clock.NewFixed(syncNow))
	availability := ProvideAvailabilityService(repositories.ProvideAbsenceRepository(pdb), people, authz, tx, recorder)
	fetcher := &fakeFetcher{body: sharedOffsite}
	service, err := ProvideCalendarFeedService(repositories.ProvideCalendarFeedRepository(pdb), people, availability, authz, tx,
		clock.NewFixed(syncNow), feedConfig(testFeedKey), fetcher)
	require.NoError(t, err)
	planner := asRole("kc-planner", constants.RolePlanner)

	calendarLeave := func(t *testing.T, personID string) []models.Absence {
		t.Helper()
		var absences []models.Absence
		require.NoError(t, pdb.Where("person_id = ? AND source = ?", personID, constants.AbsenceSourceCalendar).
			Order("from_date").Find(&absences).Error)
		return absences
	}
	eventTypes := func(t *testing.T) []string {
		t.Helper()
		var events []models.ChangeEvent
		require.NoError(t, pdb.Order("occurred_at, id").Find(&events).Error)
		types := make([]string, len(events))
		for i, e := range events {
			types[i] = e.Type
		}
		return types
	}

	t.Run("two people may share one event", func(t *testing.T) {
		_, err := service.Connect(planner, ConnectFeedInput{PersonID: alex.ID, URL: googleLink})
		require.NoError(t, err)
		_, err = service.Connect(planner, ConnectFeedInput{PersonID: bao.ID, URL: googleLink})
		require.NoError(t, err)
		require.Len(t, calendarLeave(t, alex.ID), 1)
		require.Len(t, calendarLeave(t, bao.ID), 1)
		assert.Equal(t, constants.AbsenceStatusPlanned, calendarLeave(t, alex.ID)[0].Status)
		assert.Equal(t, []string{constants.EventAbsenceAdded, constants.EventAbsenceAdded}, eventTypes(t))
	})

	t.Run("a moved event moves the absence and records both weeks", func(t *testing.T) {
		require.NoError(t, pdb.Model(&models.Absence{}).Where("person_id = ? AND source = ?", alex.ID, constants.AbsenceSourceCalendar).
			Update("status", constants.AbsenceStatusApproved).Error)
		fetcher.body = movedOffsite
		_, err := service.SyncNow(planner, alex.ID)
		require.NoError(t, err)
		leave := calendarLeave(t, alex.ID)
		require.Len(t, leave, 1)
		assert.Equal(t, 26, leave[0].FromDate.Day())
		assert.Equal(t, constants.AbsenceStatusPlanned, leave[0].Status, "a moved event needs approval again")
		types := eventTypes(t)
		assert.Equal(t, []string{constants.EventAbsenceRemoved, constants.EventAbsenceAdded}, types[len(types)-2:])
	})

	t.Run("a deleted event deletes the absence, and it can come back", func(t *testing.T) {
		fetcher.body = emptyCalendar
		_, err := service.SyncNow(planner, alex.ID)
		require.NoError(t, err)
		assert.Empty(t, calendarLeave(t, alex.ID))

		fetcher.body = sharedOffsite
		_, err = service.SyncNow(planner, alex.ID)
		require.NoError(t, err, "a deleted absence does not block the same event UID")
		assert.Len(t, calendarLeave(t, alex.ID), 1)
	})

	t.Run("a sync never touches an absence entered by hand", func(t *testing.T) {
		var kept models.Absence
		require.NoError(t, pdb.First(&kept, "id = ?", manual.ID).Error)
		assert.Equal(t, constants.AbsenceStatusApproved, kept.Status)
		assert.Equal(t, constants.AbsenceSourceManual, kept.Source)
	})
}
```

- [ ] **Step 2: Run it.** This needs Docker.

  Run: `cd server && go test -tags=integration ./internal/services -run TestCalendarFeedService_Sync_Postgres -v`
  Expected: `--- PASS: TestCalendarFeedService_Sync_Postgres` with its four subtests.

  If it fails on "two people may share one event" with a duplicate-key error, the migration does not have the partial unique index from Task 5 Step 9. Regenerate it; never edit the SQL by hand.

- [ ] **Step 3: Run the server checks, then commit**

```bash
cd server && go vet -tags=integration ./... && go test -tags=integration ./... && make lint
git add internal/services/calendarfeed_integration_test.go
git commit -m "test(server): calendar sync on Postgres: shared events, moves, deletes" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 9: HTTP routes, wiring and the runner hook

**Files:**
- Modify: `server/internal/constants/routes.go`
- Modify: `server/internal/constants/roles.go`
- Create: `server/internal/dtos/calendarfeed.go`
- Create: `server/internal/handlers/calendarfeed.go`
- Create: `server/internal/handlers/calendarfeed_test.go`
- Modify: `server/cmd/server/routes/router.go`
- Modify: `server/cmd/server/routes/router_test.go`
- Modify: `server/cmd/server/main.go`
- Regenerate: `server/docs/*`

**Interfaces:**
- Consumes:
  - Task 4: `calendar.ProvideFetcher`
  - Task 5: `ProvideCalendarFeedRepository`
  - Task 7: `CalendarFeedService`, `ProvideCalendarFeedService`, `ConnectFeedInput`
  - existing: `BaseHandler` (`pathID`, `SuccessResponse`, `HandleError`, `ValidationErrorResponse`), the `serve` and `newTestValidator` test helpers in `internal/handlers`
- Produces:
  - `constants.PathPersonCalendarFeed = "/people/:id/calendar-feed"` and `constants.PathPersonCalendarFeedSync = "/people/:id/calendar-feed/sync"`
  - `constants.CalendarFeedRoles`
  - `dtos.ConnectCalendarFeedRequest`, `dtos.CalendarFeedResponse` and `dtos.NewCalendarFeedResponse`
  - `handlers.CalendarFeedHandler` with `Get`, `Connect`, `Disconnect` and `Sync`, and `ProvideCalendarFeedHandler`
  - the four routes:

| Method | Path | Handler |
|---|---|---|
| GET | `/people/{id}/calendar-feed` | `Get` |
| PUT | `/people/{id}/calendar-feed` | `Connect` |
| DELETE | `/people/{id}/calendar-feed` | `Disconnect` |
| POST | `/people/{id}/calendar-feed/sync` | `Sync` |

- [ ] **Step 1: Write the failing handler test.** Create `server/internal/handlers/calendarfeed_test.go`:

```go
package handlers

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const calendarLink = "https://calendar.google.com/calendar/ical/a%40b/private-secret/basic.ics"

// fakeFeeds records the link it is given and answers with a connected calendar.
type fakeFeeds struct {
	services.CalendarFeedService
	connected *services.ConnectFeedInput
}

func (f *fakeFeeds) Connect(_ context.Context, in services.ConnectFeedInput) (*models.CalendarFeed, error) {
	f.connected = &in
	return &models.CalendarFeed{Host: "calendar.google.com", URLCiphertext: []byte(in.URL), EventsFound: 2}, nil
}

func TestCalendarFeedHandler_Connect(t *testing.T) {
	route := constants.PathPersonCalendarFeed
	target := strings.Replace(route, ":id", testPersonID, 1)
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "connects a trimmed link", body: `{"url":"  ` + calendarLink + `  "}`, wantStatus: http.StatusOK},
		{name: "not a link", body: `{"url":"calendar"}`, wantStatus: http.StatusBadRequest},
		{name: "empty", body: `{}`, wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feeds := &fakeFeeds{}
			h := ProvideCalendarFeedHandler(feeds, newTestValidator(t))
			rec, _ := serve(t, route, http.MethodPut, target, tt.body, h.Connect)

			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantStatus != http.StatusOK {
				assert.Nil(t, feeds.connected)
				return
			}
			require.NotNil(t, feeds.connected)
			assert.Equal(t, calendarLink, feeds.connected.URL)
			assert.Equal(t, testPersonID, feeds.connected.PersonID)
			assert.NotContains(t, rec.Body.String(), "private-secret", "the response never holds the link")
		})
	}
}
```

- [ ] **Step 2: Write the failing RBAC rows.** In `server/cmd/server/routes/router_test.go`, `TestRouter_People_RBAC`, add these rows to the table:

```go
		{name: "no role cannot read a calendar", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/calendar-feed", token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "no token cannot connect a calendar", method: http.MethodPut, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/calendar-feed", wantStatus: http.StatusUnauthorized},
		{name: "no role cannot sync a calendar", method: http.MethodPost, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/calendar-feed/sync", token: "no-role-token", wantStatus: http.StatusForbidden},
```

  In `newTestRouter`:
  - add `calendarFeeds := handlers.ProvideCalendarFeedHandler(nil, nil)` after the other handlers
  - pass `calendarFeeds` to `Router(...)` right before `verifier`

  The service checks "self or planner", so a lead or specialist on someone else's calendar gets `403` from the service. Task 7's `TestCalendarFeedService_GetAndDisconnect` pins that. These rows pin the route-level role check and the `401`.

- [ ] **Step 3: Run them to make sure they fail**

  Run: `cd server && go test ./internal/handlers ./cmd/server/routes`
  Expected: FAIL to compile, with `undefined: ProvideCalendarFeedHandler` and `constants.PathPersonCalendarFeed`.

- [ ] **Step 4: Add the names.**
  - In `server/internal/constants/routes.go`, add after `PathAbsence`:

```go
	PathPersonCalendarFeed     = "/people/:id/calendar-feed"
	PathPersonCalendarFeedSync = "/people/:id/calendar-feed/sync"
```

  - In `server/internal/constants/roles.go`, add above `// HolidayViewRoles`:

```go
	// CalendarFeedRoles may reach the calendar-feed routes; the service lets a person act only
	// on their own calendar unless they hold AbsenceManageRoles.
	CalendarFeedRoles = []string{RolePlanner, RoleLead, RoleSpecialist}
```

- [ ] **Step 5: Write the DTOs.** Create `server/internal/dtos/calendarfeed.go`. The response has no field for the link, by design:

```go
package dtos

import (
	"strings"
	"time"

	"the-slot-finder/server/internal/models"
)

// ConnectCalendarFeedRequest is the body of PUT /people/{id}/calendar-feed: the secret iCal
// address of the person's Google or Outlook calendar.
type ConnectCalendarFeedRequest struct {
	URL string `json:"url" validate:"required,url,max=2048" example:"https://calendar.google.com/calendar/ical/name%40example.com/private-abc/basic.ics"`
}

// Normalize trims the link.
func (r *ConnectCalendarFeedRequest) Normalize() {
	r.URL = strings.TrimSpace(r.URL)
}

// CalendarFeedResponse is a calendar's status. It never holds the link, which is a secret.
type CalendarFeedResponse struct {
	Host         string     `json:"host" example:"calendar.google.com"`
	LastSyncedAt *time.Time `json:"last_synced_at"`
	LastError    *string    `json:"last_error" example:"CALENDAR_FEED_UNREACHABLE"`
	EventsFound  int        `json:"events_found" example:"2"`
}

// NewCalendarFeedResponse maps a calendar feed without its link.
func NewCalendarFeedResponse(f models.CalendarFeed) CalendarFeedResponse {
	return CalendarFeedResponse{Host: f.Host, LastSyncedAt: f.LastSyncedAt, LastError: f.LastError, EventsFound: f.EventsFound}
}
```

- [ ] **Step 6: Write the handler.** Create `server/internal/handlers/calendarfeed.go`:

```go
package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// CalendarFeedHandler serves a person's calendar link. The calendar feed service checks access
// (the person themselves, or a planner) and answers 503 while calendar sync is off.
type CalendarFeedHandler struct {
	BaseHandler
	feeds    services.CalendarFeedService
	validate *validator.Validate
}

// ProvideCalendarFeedHandler creates the calendar feed handler.
func ProvideCalendarFeedHandler(feeds services.CalendarFeedService, validate *validator.Validate) *CalendarFeedHandler {
	return &CalendarFeedHandler{BaseHandler: *NewBaseHandler(), feeds: feeds, validate: validate}
}

// Get godoc
// @Summary A person's calendar
// @Description The status of the person's connected calendar: host, last sync, last error and events found. Never the link.
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CalendarFeedResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Failure 503 {object} object{meta=dtos.Meta}
// @Router /people/{id}/calendar-feed [get]
func (h *CalendarFeedHandler) Get(c *echo.Context) error {
	personID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	feed, err := h.feeds.Get(c.Request().Context(), personID)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCalendarFeedRetrieved, dtos.NewCalendarFeedResponse(*feed), nil)
}

// Connect godoc
// @Summary Connect a calendar
// @Description Checks the secret iCal link (https, allowed host), syncs it once and stores it encrypted, replacing an earlier one.
// @Tags Availability
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Param body body dtos.ConnectCalendarFeedRequest true "Calendar link"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CalendarFeedResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 503 {object} object{meta=dtos.Meta}
// @Router /people/{id}/calendar-feed [put]
func (h *CalendarFeedHandler) Connect(c *echo.Context) error {
	personID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.ConnectCalendarFeedRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	feed, err := h.feeds.Connect(c.Request().Context(), services.ConnectFeedInput{PersonID: personID, URL: req.URL})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCalendarFeedConnected, dtos.NewCalendarFeedResponse(*feed), nil)
}

// Disconnect godoc
// @Summary Disconnect a calendar
// @Description Removes the link. Absences it created stay, as ordinary absences.
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Success 200 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /people/{id}/calendar-feed [delete]
func (h *CalendarFeedHandler) Disconnect(c *echo.Context) error {
	personID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	if err := h.feeds.Disconnect(c.Request().Context(), personID); err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCalendarFeedDisconnected, nil, nil)
}

// Sync godoc
// @Summary Sync a calendar now
// @Description Reads the calendar at once. A calendar that cannot be read keeps its absences and returns its error code in last_error.
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CalendarFeedResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /people/{id}/calendar-feed/sync [post]
func (h *CalendarFeedHandler) Sync(c *echo.Context) error {
	personID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	feed, err := h.feeds.SyncNow(c.Request().Context(), personID)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCalendarFeedSynced, dtos.NewCalendarFeedResponse(*feed), nil)
}
```

- [ ] **Step 7: Register the routes.** In `server/cmd/server/routes/router.go`:
  - add the parameter `calendarFeedHandler *handlers.CalendarFeedHandler,` to `Router`, right before `verifier auth.TokenVerifier,`
  - call `registerCalendarFeedRoutes(v1, authenticated, calendarFeedHandler)` right after `registerPeopleRoutes(...)`
  - add this function above `// registerProgramRoutes`:

```go
// registerCalendarFeedRoutes adds a person's calendar link (self or planner, checked in the service).
func registerCalendarFeedRoutes(v1 *echo.Group, authenticated echo.MiddlewareFunc, h *handlers.CalendarFeedHandler) {
	roles := middlewares.RequireRole(constants.CalendarFeedRoles...)
	v1.GET(constants.PathPersonCalendarFeed, h.Get, authenticated, roles)
	v1.PUT(constants.PathPersonCalendarFeed, h.Connect, authenticated, roles)
	v1.DELETE(constants.PathPersonCalendarFeed, h.Disconnect, authenticated, roles)
	v1.POST(constants.PathPersonCalendarFeedSync, h.Sync, authenticated, roles)
}
```

- [ ] **Step 8: Wire FX and the runner.** In `server/cmd/server/main.go`:
  - import `"the-slot-finder/server/internal/integration/calendar"`
  - in `NewHTTPServer`, add the parameter `calendarFeedHandler *handlers.CalendarFeedHandler,` right before `verifier`, and pass it to `routes.Router(...)` in the same position
  - in `appOptions`, add these providers, each next to its layer: `repositories.ProvideCalendarFeedRepository`, `calendar.ProvideFetcher`, `services.ProvideCalendarFeedService` and `handlers.ProvideCalendarFeedHandler`
  - add `fx.Invoke(registerCalendarSyncRunner),` after `fx.Invoke(registerReplanRunner),`
  - append:

```go
// registerCalendarSyncRunner syncs people's calendars in the background from startup until
// shutdown. While CALENDAR_FEED_KEY is empty the runner returns at once.
func registerCalendarSyncRunner(lc fx.Lifecycle, feeds services.CalendarFeedService) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				feeds.Run(ctx)
			}()
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			select {
			case <-done:
			case <-stopCtx.Done():
			}
			return nil
		},
	})
}
```

- [ ] **Step 9: Run the tests to make sure they pass**

  Run: `cd server && go test ./internal/handlers ./cmd/server/... -v -run 'CalendarFeed|RBAC|AppOptions'`
  Expected:
  - `--- PASS: TestCalendarFeedHandler_Connect`
  - the new `TestRouter_People_RBAC` rows pass
  - `--- PASS: TestAppOptions_Validate`. An empty `CALENDAR_FEED_KEY` is fine; the feature stays off.

- [ ] **Step 10: Regenerate Swagger**

  Run: `cd server && make swagger-load`. Without `swag`, run `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init` with the Makefile's flags.
  Then run: `grep -c "calendar-feed" docs/swagger.yaml`
  Expected: `2` or more (both paths). `grep -n "url_ciphertext" docs/swagger.yaml` prints nothing.

- [ ] **Step 11: Run the server checks, then commit**

```bash
cd server && go build ./... && go vet ./... && go test ./... && make lint && go test -tags=integration ./...
git add internal/constants/routes.go internal/constants/roles.go internal/dtos/calendarfeed.go \
  internal/handlers/calendarfeed.go internal/handlers/calendarfeed_test.go cmd/server/routes cmd/server/main.go docs
git commit -m "feat(server): calendar feed routes and the sync runner" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 10: Client: mark absences read from a calendar

**Files:**
- Modify: `client/src/shared/people/types.ts`, `constants.ts` and `index.ts` (from member-calendar)
- Create: `client/src/shared/components/absence-source-badge/` (`index.ts`, `component.tsx`, `types.ts`, `component.test.tsx`)
- Modify: `client/src/features/absences/types.ts`
- Modify: `client/src/features/absences/components/absence-row/component.tsx`, `component.test.tsx`
- Modify: `client/src/features/person-calendar/components/week-row/components/day-cell/component.tsx`
- Modify: `client/src/features/person-calendar/component.test.tsx`, `utils.test.ts`
- Modify: `client/src/shared/i18n/language-key.ts`

**Interfaces:**
- Consumes:
  - Task 5: the `source` field (`manual` | `calendar`) in `GET /people/{id}/absences` and in `days[].absences[]` of `GET /people/{id}/schedule`
  - member-calendar: `@/shared/people` (`ScheduleAbsence`, absence labels), `DayCell`, `AbsenceRow`
- Produces:
  - `@/shared/people`: `type AbsenceSource = 'manual' | 'calendar'`, `ScheduleAbsence.source: AbsenceSource`, `CALENDAR_SOURCE`, `ABSENCE_SOURCE_BADGE: Record<AbsenceSource, BadgeVariant>`
  - `Absence.source` in `features/absences`
  - `AbsenceSourceBadge` with props `{ source: AbsenceSource; compact?: boolean }`. It shows nothing for manual absences.

- [ ] **Step 1: Write the failing badge test.** Create `client/src/shared/components/absence-source-badge/component.test.tsx`:

```tsx
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { translate } from '@/shared/i18n'
import { AbsenceSourceBadge } from './component'

afterEach(cleanup)

describe('AbsenceSourceBadge', () => {
  it('marks leave read from a calendar', () => {
    render(<AbsenceSourceBadge source="calendar" />)
    expect(screen.getByText(translate('From calendar'))).toBeTruthy()
  })

  it('shows only a labelled icon when compact', () => {
    render(<AbsenceSourceBadge source="calendar" compact />)
    expect(screen.getByRole('img', { name: translate('From calendar') })).toBeTruthy()
    expect(screen.queryByText(translate('From calendar'))).toBeNull()
  })

  it('shows nothing for an absence entered by hand', () => {
    const { container } = render(<AbsenceSourceBadge source="manual" />)
    expect(container.innerHTML).toBe('')
  })
})
```

- [ ] **Step 2: Write the failing calendar and row cases**
  - In `client/src/features/person-calendar/component.test.tsx`, give the existing `a1` absence `source: 'manual'`, then add this case inside `describe('PersonCalendar', …)`:

```tsx
  it('marks leave read from a calendar', () => {
    useSchedule.mockReturnValue({
      data: {
        ...october,
        days: [
          {
            date: '2026-10-22',
            absences: [{ id: 'a1', type: 'leave', status: 'planned', hours_per_day: 8, source: 'calendar' }],
            holidays: [],
          },
        ],
      },
      isPending: false,
    })
    renderCalendar()
    expect(screen.getByRole('img', { name: 'From calendar' })).toBeTruthy()
  })
```

  - In `client/src/features/absences/components/absence-row/component.test.tsx`, add `source: 'manual'` to the `planned` fixture, and add:

```tsx
  it('marks an absence read from a calendar', () => {
    renderRow({ ...planned, source: 'calendar' })
    expect(screen.getByText('From calendar')).toBeTruthy()
  })
```

  - Add `source: 'manual'` to every other absence literal in the client tests:

```bash
cd client && grep -rln "hours_per_day" src | grep -E "[.]test[.]tsx?$"
```

  Expected files include `features/person-calendar/utils.test.ts`. Add `source: 'manual'` to each absence object found there.

- [ ] **Step 3: Run them to make sure they fail**

  Run: `cd client && npx tsc -b`
  Expected: errors such as `Object literal may only specify known properties, and 'source' does not exist in type 'ScheduleAbsence'`, and `Cannot find module '@/shared/components/absence-source-badge'`.

- [ ] **Step 4: Add the source to `@/shared/people`.**
  - In `client/src/shared/people/types.ts`, add:

```ts
/** Where an absence comes from: entered by hand, or read from the person's calendar. */
export type AbsenceSource = 'manual' | 'calendar'
```

    and `source: AbsenceSource` as the last field of `ScheduleAbsence`.
  - In `client/src/shared/people/constants.ts`, add `import type { BadgeVariant } from '@/shared/ui/badge'`, add `AbsenceSource` to the `./types` import, and append:

```ts
/** The source that gets a mark; it matches the server's constants/absence.go. */
export const CALENDAR_SOURCE = 'calendar' satisfies AbsenceSource

export const ABSENCE_SOURCE_BADGE: Record<AbsenceSource, BadgeVariant> = {
  manual: 'gray',
  calendar: 'teal',
}
```

  - In `client/src/shared/people/index.ts`, export `ABSENCE_SOURCE_BADGE`, `CALENDAR_SOURCE` and `type AbsenceSource`.

- [ ] **Step 5: Write the badge.** Create the folder `client/src/shared/components/absence-source-badge/`.

  `types.ts`:

```ts
import type { AbsenceSource } from '@/shared/people'

export interface AbsenceSourceBadgeProps {
  source: AbsenceSource
  /** Only the icon, for inside another badge such as a calendar day's absence. */
  compact?: boolean
}
```

  `component.tsx`:

```tsx
import { CalendarSync } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { ABSENCE_SOURCE_BADGE, CALENDAR_SOURCE } from '@/shared/people'
import { Badge } from '@/shared/ui/badge'
import type { AbsenceSourceBadgeProps } from './types'

/** Marks an absence read from the person's calendar; one entered by hand shows nothing. */
export function AbsenceSourceBadge({ source, compact }: AbsenceSourceBadgeProps) {
  const { t } = useTranslation()
  if (source !== CALENDAR_SOURCE) return null
  if (compact) return <CalendarSync data-icon="inline-start" role="img" aria-label={t('From calendar')} />

  return (
    <Badge variant={ABSENCE_SOURCE_BADGE[source]}>
      <CalendarSync data-icon="inline-start" aria-hidden />
      {t('From calendar')}
    </Badge>
  )
}
```

  `index.ts`:

```ts
export { AbsenceSourceBadge } from './component'
export type { AbsenceSourceBadgeProps } from './types'
```

- [ ] **Step 6: Use it.**
  - In `client/src/features/absences/types.ts`, add `import type { AbsenceSource } from '@/shared/people'` and the field `source: AbsenceSource` after `status` in `Absence`.
  - In `client/src/features/absences/components/absence-row/component.tsx`, import `AbsenceSourceBadge` from `@/shared/components/absence-source-badge`, and make the type cell:

```tsx
      <TableCell>
        <div className="flex items-center gap-2">
          {t(ABSENCE_TYPE_LABELS[absence.type])}
          <AbsenceSourceBadge source={absence.source} />
        </div>
      </TableCell>
```

  - In `client/src/features/person-calendar/components/week-row/components/day-cell/component.tsx`, import the badge the same way, and put the compact mark first inside each absence's `Badge`:

```tsx
        {day.absences.map((absence) => (
          <Badge
            key={absence.id}
            variant={absence.status === PLANNED_STATUS ? PLANNED_ABSENCE_BADGE : ABSENCE_BADGE[absence.type]}
          >
            <AbsenceSourceBadge source={absence.source} compact />
            {absence.status === PLANNED_STATUS
              ? t('{type} (planned)', { type: t(ABSENCE_TYPE_LABELS[absence.type]) })
              : t(ABSENCE_TYPE_LABELS[absence.type])}
          </Badge>
        ))}
```

  - In `client/src/shared/i18n/language-key.ts`, add `'From calendar': string` in its sorted place.

- [ ] **Step 7: Run the tests to make sure they pass**

  Run: `cd client && npm run format && npx tsc -b && npx vitest run src/shared/components/absence-source-badge src/features/absences src/features/person-calendar`
  Expected: all pass, including "marks leave read from a calendar" and "marks an absence read from a calendar".

- [ ] **Step 8: Run the client checks, then commit**

```bash
cd client && npx tsc -b && npm run lint && npm test && npm run build
git add src/shared/people src/shared/components/absence-source-badge src/features/absences src/features/person-calendar \
  src/shared/i18n/language-key.ts
git commit -m "feat(client): mark absences read from a calendar" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 11: Client: the calendar panel on My schedule and the person page

**Files:**
- Modify: `client/src/resources/endpoints/people.ts`
- Modify: `client/src/utils/constants/query.ts`
- Modify: `client/src/features/absences/constants.ts`
- Modify: `client/src/resources/rbac/permissions.ts`
- Create: `client/src/features/calendar-feed/` with `index.ts`, `component.tsx`, `component.test.tsx`, `types.ts`, `constants.ts`, `api.ts`, `utils.ts`, `utils.test.ts` and `use-calendar-feed.ts`
- Create: `client/src/features/calendar-feed/components/calendar-feed-form/` with `index.ts`, `component.tsx`, `types.ts` and `utils.ts`
- Create: `client/src/features/calendar-feed/components/calendar-feed-status/` with `index.ts`, `component.tsx` and `types.ts`
- Modify: `client/src/pages/my-schedule/component.tsx`
- Modify: `client/src/pages/person-detail/component.tsx`
- Modify: `client/src/shared/i18n/language-key.ts`

**Interfaces:**
- Consumes:
  - Task 9 routes: `GET`/`PUT`/`DELETE /people/{id}/calendar-feed` and `POST …/sync`. The response is `{ host, last_synced_at, last_error, events_found }`; the error codes are those of Task 7.
  - `INSIGHTS_QUERY_KEY`: member-calendar's `scheduleQueryKey` sits under it, so invalidating it refreshes the calendar.
  - `Section`, `Button`, `Input`, `Label`, `toast`, `formatDateTime`, `can`
- Produces:
  - `CalendarFeed` with props `{ personId: string }`, exported from `@/features/calendar-feed`
  - `PEOPLE_ENDPOINTS.calendarFeed(id)` and `PEOPLE_ENDPOINTS.calendarFeedSync(id)`
  - `ABSENCES_QUERY_KEY`
  - permission `'calendar-feeds:manage'` (planner)

**What the panel shows:**
- A "no calendar" answer (`404 CALENDAR_FEED_NOT_FOUND`) and a "sync is off" answer (`503 CALENDAR_SYNC_DISABLED`) are states, not errors. `api.ts` turns them into states, so no error toast shows.
- The connect form shows its own link errors: `meta.quietErrorCodes` keeps the shared toast quiet for those codes.

- [ ] **Step 1: Write the failing tests.**

  `client/src/features/calendar-feed/utils.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { ApiError } from '@/shared/api/types'
import { CALENDAR_FEED_NOT_FOUND, CALENDAR_SYNC_DISABLED } from './constants'
import { stateFromError } from './utils'

const HTTP_NOT_FOUND = 404
const HTTP_UNAVAILABLE = 503
const HTTP_SERVER_ERROR = 500

function apiError(status: number, code: string) {
  return new ApiError({ status, code, message: code })
}

describe('stateFromError', () => {
  it('treats "no calendar" and "sync is off" as states', () => {
    expect(stateFromError(apiError(HTTP_NOT_FOUND, CALENDAR_FEED_NOT_FOUND))).toEqual({ kind: 'none' })
    expect(stateFromError(apiError(HTTP_UNAVAILABLE, CALENDAR_SYNC_DISABLED))).toEqual({ kind: 'disabled' })
  })

  it('leaves other errors to the shared toast', () => {
    expect(stateFromError(apiError(HTTP_SERVER_ERROR, 'INTERNAL_ERROR'))).toBeNull()
    expect(stateFromError(new Error('network'))).toBeNull()
  })
})
```

  `client/src/features/calendar-feed/component.test.tsx`:

```tsx
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/shared/api/types'
import { translate } from '@/shared/i18n'
import { CalendarFeed } from './component'
import { CALENDAR_FEED_ERROR_LABELS } from './constants'
import type { CalendarFeedState } from './types'
import { useCalendarFeed, useConnectCalendarFeed } from './use-calendar-feed'

vi.mock('./use-calendar-feed', () => ({
  useCalendarFeed: vi.fn(),
  useConnectCalendarFeed: vi.fn(),
  useSyncCalendarFeed: () => ({ mutate: vi.fn(), isPending: false }),
  useDisconnectCalendarFeed: () => ({ mutate: vi.fn(), isPending: false }),
}))

const PERSON_ID = 'p1'
const HOST = 'calendar.google.com'
const LINK = 'https://calendar.google.com/calendar/ical/a%40b/private-x/basic.ics'
const HTTP_BAD_REQUEST = 400
const connect = vi.fn()

function showState(state: CalendarFeedState) {
  vi.mocked(useCalendarFeed).mockReturnValue({ data: state, isPending: false } as ReturnType<typeof useCalendarFeed>)
}

function connectFails(code: string) {
  vi.mocked(useConnectCalendarFeed).mockReturnValue({
    mutate: connect,
    isPending: false,
    error: new ApiError({ status: HTTP_BAD_REQUEST, code, message: code }),
  } as unknown as ReturnType<typeof useConnectCalendarFeed>)
}

beforeEach(() => {
  vi.mocked(useConnectCalendarFeed).mockReturnValue({
    mutate: connect,
    isPending: false,
    error: null,
  } as unknown as ReturnType<typeof useConnectCalendarFeed>)
})

afterEach(() => {
  cleanup()
  connect.mockReset()
})

describe('CalendarFeed', () => {
  it('says when sync is turned off on the server', () => {
    showState({ kind: 'disabled' })
    render(<CalendarFeed personId={PERSON_ID} />)
    expect(screen.getByText(translate('Calendar sync is turned off on this server.'))).toBeTruthy()
  })

  it('connects a pasted link, trimmed', () => {
    showState({ kind: 'none' })
    render(<CalendarFeed personId={PERSON_ID} />)
    fireEvent.change(screen.getByLabelText(translate('Secret iCal address')), { target: { value: `  ${LINK}  ` } })
    fireEvent.click(screen.getByRole('button', { name: translate('Connect calendar') }))
    expect(connect).toHaveBeenCalledWith(LINK, expect.anything())
  })

  it('explains a link the organization blocks', () => {
    showState({ kind: 'none' })
    connectFails('CALENDAR_FEED_NOT_ICS')
    render(<CalendarFeed personId={PERSON_ID} />)
    expect(screen.getByText(translate(CALENDAR_FEED_ERROR_LABELS.CALENDAR_FEED_NOT_ICS))).toBeTruthy()
  })

  it('shows the connected calendar and its last problem, never the link', () => {
    showState({
      kind: 'connected',
      calendar: { host: HOST, last_synced_at: null, last_error: 'CALENDAR_FEED_UNREACHABLE', events_found: 2 },
    })
    render(<CalendarFeed personId={PERSON_ID} />)
    expect(screen.getByText(HOST)).toBeTruthy()
    expect(screen.getByText(translate('Not yet'))).toBeTruthy()
    expect(screen.getByText(translate(CALENDAR_FEED_ERROR_LABELS.CALENDAR_FEED_UNREACHABLE))).toBeTruthy()
    expect(screen.getByRole('button', { name: translate('Sync now') })).toBeTruthy()
  })
})
```

- [ ] **Step 2: Run them to make sure they fail**

  Run: `cd client && npx vitest run src/features/calendar-feed`
  Expected: FAIL with `Failed to resolve import "./utils"` and `"./component"`.

- [ ] **Step 3: Add the endpoint and the query-key prefix.**
  - In `client/src/resources/endpoints/people.ts`, add to `PEOPLE_ENDPOINTS` after `assignments`:

```ts
  calendarFeed: (id: string) => `${person(id)}/calendar-feed`,
  calendarFeedSync: (id: string) => `${person(id)}/calendar-feed/sync`,
```

  - Append to `client/src/utils/constants/query.ts`:

```ts
/** Prefix of every person's absence list, refreshed when a calendar sync changes absences. */
export const ABSENCES_QUERY_KEY = ['absences'] as const
```

  - In `client/src/features/absences/constants.ts`, import `ABSENCES_QUERY_KEY` from `@/utils/constants` and build the key on it:

```ts
export const absencesQueryKey = (personId: string) => [...ABSENCES_QUERY_KEY, personId] as const
```

- [ ] **Step 4: Write the feature.**

  `client/src/features/calendar-feed/types.ts`:

```ts
/** Problems the server stores or answers with; they match its constants/error_codes.go. */
export type CalendarFeedErrorCode =
  'CALENDAR_FEED_HOST_NOT_ALLOWED' | 'CALENDAR_FEED_UNREACHABLE' | 'CALENDAR_FEED_NOT_ICS' | 'CALENDAR_FEED_TOO_LARGE'

/** A connected calendar as the server returns it. The link itself never comes back. */
export interface ConnectedCalendar {
  host: string
  /** ISO timestamp of the last good sync, or null before the first one. */
  last_synced_at: string | null
  last_error: CalendarFeedErrorCode | null
  events_found: number
}

/** What the panel shows: sync is off on the server, nothing connected yet, or a calendar. */
export type CalendarFeedState =
  { kind: 'disabled' } | { kind: 'none' } | { kind: 'connected'; calendar: ConnectedCalendar }

export interface CalendarFeedProps {
  personId: string
}
```

  `client/src/features/calendar-feed/constants.ts`:

```ts
import type { PlainKey } from '@/shared/i18n'
import type { CalendarFeedErrorCode } from './types'

export const calendarFeedQueryKey = (personId: string) => ['calendar-feed', personId] as const

/** Answers that are states, not errors: no calendar yet, or sync turned off on the server. */
export const CALENDAR_FEED_NOT_FOUND = 'CALENDAR_FEED_NOT_FOUND'
export const CALENDAR_SYNC_DISABLED = 'CALENDAR_SYNC_DISABLED'

/** What to tell the person for each problem with their link. */
export const CALENDAR_FEED_ERROR_LABELS: Record<CalendarFeedErrorCode, PlainKey> = {
  CALENDAR_FEED_HOST_NOT_ALLOWED:
    'Use the secret iCal address from Google Calendar or Outlook. It starts with https://.',
  CALENDAR_FEED_UNREACHABLE: 'The calendar could not be fetched. Your organization may block secret calendar links.',
  CALENDAR_FEED_NOT_ICS: 'The link does not return a calendar. Your organization may block secret calendar links.',
  CALENDAR_FEED_TOO_LARGE: 'The calendar is too large to read.',
}

/** The codes the connect form shows itself, so the shared error toast skips them. */
export const CALENDAR_FEED_ERROR_CODES = [
  'CALENDAR_FEED_HOST_NOT_ALLOWED',
  'CALENDAR_FEED_UNREACHABLE',
  'CALENDAR_FEED_NOT_ICS',
  'CALENDAR_FEED_TOO_LARGE',
] as const satisfies readonly CalendarFeedErrorCode[]
```

  `client/src/features/calendar-feed/utils.ts`:

```ts
import { ApiError } from '@/shared/api/types'
import { CALENDAR_FEED_NOT_FOUND, CALENDAR_SYNC_DISABLED } from './constants'
import type { CalendarFeedState } from './types'

/** "No calendar yet" and "sync is off" are answers, not errors. Anything else returns null. */
export function stateFromError(error: unknown): CalendarFeedState | null {
  if (!(error instanceof ApiError)) return null
  if (error.code === CALENDAR_FEED_NOT_FOUND) return { kind: 'none' }
  if (error.code === CALENDAR_SYNC_DISABLED) return { kind: 'disabled' }
  return null
}
```

  `client/src/features/calendar-feed/api.ts`:

```ts
import { PEOPLE_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { CalendarFeedState, ConnectedCalendar } from './types'
import { stateFromError } from './utils'

/** The panel's state. "Not connected" and "turned off" come back as states, so no toast shows. */
export async function getCalendarFeed(personId: string): Promise<CalendarFeedState> {
  try {
    const calendar = await request<ConnectedCalendar>({ url: PEOPLE_ENDPOINTS.calendarFeed(personId) })
    return { kind: 'connected', calendar }
  } catch (error) {
    const state = stateFromError(error)
    if (state) return state
    throw error
  }
}

export function connectCalendarFeed(personId: string, url: string) {
  return request<ConnectedCalendar>({ url: PEOPLE_ENDPOINTS.calendarFeed(personId), method: 'PUT', data: { url } })
}

export function syncCalendarFeed(personId: string) {
  return request<ConnectedCalendar>({ url: PEOPLE_ENDPOINTS.calendarFeedSync(personId), method: 'POST' })
}

export function disconnectCalendarFeed(personId: string) {
  return request<null>({ url: PEOPLE_ENDPOINTS.calendarFeed(personId), method: 'DELETE' })
}
```

  `client/src/features/calendar-feed/use-calendar-feed.ts`:

```ts
import { type QueryClient, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ABSENCES_QUERY_KEY, INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { connectCalendarFeed, disconnectCalendarFeed, getCalendarFeed, syncCalendarFeed } from './api'
import { CALENDAR_FEED_ERROR_CODES, calendarFeedQueryKey } from './constants'

/** A sync changes absences, so the panel, the absence lists and the calendar refresh together. */
function refreshAfterSync(queryClient: QueryClient, personId: string) {
  return Promise.all([
    queryClient.invalidateQueries({ queryKey: calendarFeedQueryKey(personId) }),
    queryClient.invalidateQueries({ queryKey: ABSENCES_QUERY_KEY }),
    queryClient.invalidateQueries({ queryKey: INSIGHTS_QUERY_KEY }),
  ])
}

export function useCalendarFeed(personId: string) {
  return useQuery({ queryKey: calendarFeedQueryKey(personId), queryFn: () => getCalendarFeed(personId) })
}

export function useConnectCalendarFeed(personId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (url: string) => connectCalendarFeed(personId, url),
    onSuccess: () => refreshAfterSync(queryClient, personId),
    meta: { quietErrorCodes: CALENDAR_FEED_ERROR_CODES },
  })
}

export function useSyncCalendarFeed(personId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => syncCalendarFeed(personId),
    onSuccess: () => refreshAfterSync(queryClient, personId),
  })
}

export function useDisconnectCalendarFeed(personId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => disconnectCalendarFeed(personId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: calendarFeedQueryKey(personId) }),
  })
}
```

  `client/src/features/calendar-feed/component.tsx`:

```tsx
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { CalendarFeedForm } from './components/calendar-feed-form'
import { CalendarFeedStatus } from './components/calendar-feed-status'
import type { CalendarFeedProps } from './types'
import { useCalendarFeed } from './use-calendar-feed'

/** A person's Google or Outlook calendar: connect it, or see how its last sync went. */
export function CalendarFeed({ personId }: CalendarFeedProps) {
  const { t } = useTranslation()
  const { data: state, isPending } = useCalendarFeed(personId)

  return (
    <Section
      title={t('Calendar')}
      description={t(
        'Out-of-office events in a Google or Outlook calendar become planned leave for a planner to approve.',
      )}
      card
    >
      {isPending && <p className="text-sm text-muted-foreground">{t('Loading…')}</p>}
      {state?.kind === 'disabled' && (
        <p className="text-sm text-muted-foreground">{t('Calendar sync is turned off on this server.')}</p>
      )}
      {state?.kind === 'none' && <CalendarFeedForm personId={personId} />}
      {state?.kind === 'connected' && <CalendarFeedStatus personId={personId} calendar={state.calendar} />}
    </Section>
  )
}
```

  `client/src/features/calendar-feed/index.ts`:

```ts
export { CalendarFeed } from './component'
export type { CalendarFeedProps } from './types'
```

  `client/src/features/calendar-feed/components/calendar-feed-form/types.ts`:

```ts
export interface CalendarFeedFormProps {
  personId: string
}
```

  `client/src/features/calendar-feed/components/calendar-feed-form/utils.ts`:

```ts
import type { ApiError } from '@/shared/api/types'
import { CALENDAR_FEED_ERROR_CODES } from '../../constants'
import type { CalendarFeedErrorCode } from '../../types'

/** The link problem to show under the field, or null for none (other errors use the toast). */
export function linkProblem(error: ApiError | null): CalendarFeedErrorCode | null {
  return CALENDAR_FEED_ERROR_CODES.find((code) => code === error?.code) ?? null
}
```

  `client/src/features/calendar-feed/components/calendar-feed-form/component.tsx`:

```tsx
import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { CALENDAR_FEED_ERROR_LABELS } from '../../constants'
import { useConnectCalendarFeed } from '../../use-calendar-feed'
import type { CalendarFeedFormProps } from './types'
import { linkProblem } from './utils'

/** Paste the secret iCal address; the server checks it and syncs it once. */
export function CalendarFeedForm({ personId }: CalendarFeedFormProps) {
  const { t } = useTranslation()
  const [link, setLink] = useState('')
  const connect = useConnectCalendarFeed(personId)
  const problem = linkProblem(connect.error)
  const inputId = `calendar-link-${personId}`

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => setLink(event.target.value)
  const handleConnected = () => {
    toast.success(t('Calendar connected.'))
    setLink('')
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    connect.mutate(link.trim(), { onSuccess: handleConnected })
  }

  return (
    <form className="flex flex-col gap-3" onSubmit={handleSubmit}>
      <div className="flex flex-col gap-2">
        <Label htmlFor={inputId}>{t('Secret iCal address')}</Label>
        <Input
          id={inputId}
          type="url"
          value={link}
          onChange={handleChange}
          autoComplete="off"
          aria-invalid={problem !== null}
          required
        />
        <p className="text-sm text-muted-foreground">
          {t('Google: Settings and sharing → Integrate calendar → Secret address in iCal format.')}
        </p>
        <p className="text-sm text-muted-foreground">
          {t('Outlook: Settings → Calendar → Shared calendars → Publish a calendar → ICS link.')}
        </p>
        {problem && <p className="text-sm text-destructive">{t(CALENDAR_FEED_ERROR_LABELS[problem])}</p>}
      </div>
      <Button type="submit" className="self-start" disabled={connect.isPending}>
        {t('Connect calendar')}
      </Button>
    </form>
  )
}
```

  `client/src/features/calendar-feed/components/calendar-feed-form/index.ts`:

```ts
export { CalendarFeedForm } from './component'
```

  `client/src/features/calendar-feed/components/calendar-feed-status/types.ts`:

```ts
import type { ConnectedCalendar } from '../../types'

export interface CalendarFeedStatusProps {
  personId: string
  calendar: ConnectedCalendar
}
```

  `client/src/features/calendar-feed/components/calendar-feed-status/component.tsx`:

```tsx
import { RefreshCw } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { formatDateTime } from '@/utils/functions'
import { CALENDAR_FEED_ERROR_LABELS } from '../../constants'
import { useDisconnectCalendarFeed, useSyncCalendarFeed } from '../../use-calendar-feed'
import type { CalendarFeedStatusProps } from './types'

/** Where the calendar is, how its last sync went, and the actions on it. */
export function CalendarFeedStatus({ personId, calendar }: CalendarFeedStatusProps) {
  const { t } = useTranslation()
  const sync = useSyncCalendarFeed(personId)
  const disconnect = useDisconnectCalendarFeed(personId)

  const handleSync = () => sync.mutate()
  const handleDisconnect = () => {
    if (!window.confirm(t('Disconnect this calendar? Leave it already added stays.'))) return
    disconnect.mutate()
  }

  return (
    <div className="flex flex-col gap-3">
      <dl className="grid gap-x-6 gap-y-1 text-sm sm:grid-cols-[auto_1fr]">
        <dt className="text-muted-foreground">{t('Connected to')}</dt>
        <dd>{calendar.host}</dd>
        <dt className="text-muted-foreground">{t('Last sync')}</dt>
        <dd>{calendar.last_synced_at ? formatDateTime(calendar.last_synced_at) : t('Not yet')}</dd>
        <dt className="text-muted-foreground">{t('Out-of-office events')}</dt>
        <dd>{calendar.events_found}</dd>
      </dl>
      {calendar.last_error && (
        <p className="text-sm text-destructive">{t(CALENDAR_FEED_ERROR_LABELS[calendar.last_error])}</p>
      )}
      <div className="flex gap-2">
        <Button variant="outline" onClick={handleSync} disabled={sync.isPending}>
          <RefreshCw />
          {t('Sync now')}
        </Button>
        <Button variant="ghost" onClick={handleDisconnect} disabled={disconnect.isPending}>
          {t('Disconnect')}
        </Button>
      </div>
    </div>
  )
}
```

  `client/src/features/calendar-feed/components/calendar-feed-status/index.ts`:

```ts
export { CalendarFeedStatus } from './component'
```

- [ ] **Step 5: Add the i18n keys.** Add each key below to `client/src/shared/i18n/language-key.ts`, in its sorted place (case-insensitive), as `'Key': string`, or as a bare `Key: string` for one word:

```text
Calendar
Calendar connected.
Calendar sync is turned off on this server.
Connect calendar
Connected to
Disconnect
Disconnect this calendar? Leave it already added stays.
Google: Settings and sharing → Integrate calendar → Secret address in iCal format.
Last sync
Not yet
Out-of-office events
Out-of-office events in a Google or Outlook calendar become planned leave for a planner to approve.
Outlook: Settings → Calendar → Shared calendars → Publish a calendar → ICS link.
Secret iCal address
Sync now
The calendar could not be fetched. Your organization may block secret calendar links.
The calendar is too large to read.
The link does not return a calendar. Your organization may block secret calendar links.
Use the secret iCal address from Google Calendar or Outlook. It starts with https://.
```

- [ ] **Step 6: Put the panel on the pages.**
  - Add `'calendar-feeds:manage': [ROLES.PLANNER],` to `PERMISSIONS` in `client/src/resources/rbac/permissions.ts`.
  - In `client/src/pages/my-schedule/component.tsx`, import `CalendarFeed` from `@/features/calendar-feed`, and render `<CalendarFeed personId={user.person_id} />` right after the absences `Section`, inside the branch that has a `person_id`. A person always manages their own calendar.
  - In `client/src/pages/person-detail/component.tsx`, import `CalendarFeed`, `can` (if not imported yet) and the current user the way the page already gets it. Render this right after the absences `Section`:

```tsx
{can(user?.role, 'calendar-feeds:manage') && <CalendarFeed personId={id} />}
```

    A lead reads the person page but may not see someone else's calendar: the server would answer 403.

- [ ] **Step 7: Run the tests to make sure they pass**

  Run: `cd client && npm run format && npx tsc -b && npx vitest run src/features/calendar-feed src/pages`
  Expected:
  - `utils.test.ts`: 2 tests pass
  - `component.test.tsx`: 4 tests pass
  - the existing page tests still pass. Their `vi.mock` of features covers only what they mock; if a page test now fails on `useCalendarFeed`, add `vi.mock('@/features/calendar-feed', () => ({ CalendarFeed: () => null }))` to it, next to its `AbsenceList` mock.

- [ ] **Step 8: Run the client checks, then commit**

```bash
cd client && npx tsc -b && npm run lint && npm test && npm run build
git add src/resources src/utils/constants/query.ts src/features/absences/constants.ts src/features/calendar-feed \
  src/pages/my-schedule src/pages/person-detail src/shared/i18n/language-key.ts
git commit -m "feat(client): connect a calendar on My schedule and the person page" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 12: Check it end to end, then update the docs

**Files:**
- Modify: `docs/features/calendar-sync.md` (status, an **As built** section, ticked tasks)
- Modify: `docs/features/README.md` (row 13 status)
- Modify: `server/README.md` §3, only if the built API differs from the docs PR

**Interfaces:**
- Consumes: everything above.
- Produces: a working demo path, and docs that match it.

- [ ] **Step 1: Turn the feature on locally**

  Make a key and put it in `server/cmd/server/.env`, never in `.env.example`:

```bash
cd server && printf 'CALENDAR_FEED_KEY=%s\n' "$(openssl rand -base64 32)" >> cmd/server/.env
make migrate-up && make up
```

  For a quick demo cycle, also set `CALENDAR_SYNC_INTERVAL=1m` in `cmd/server/.env`.

- [ ] **Step 2: Walk the demo path.** Sign in as the planner and open a person's page, then:
  1. Paste the secret link from Task 1.
  2. Expected: "Calendar connected.", `calendar.google.com` with the event count, and "Leave (planned)" with the calendar mark on 26–28 Oct, in both the calendar and the absence list.
  3. Approve that leave (member-calendar's Approve).
  4. Expected: within `REPLAN_DEBOUNCE`, `/events` shows the change and a re-plan draft if the person's hours no longer fit.
  5. Move the event in Google to 2–4 Nov, then press **Sync now**.
  6. Expected: the absence moves and is "planned" again.
  7. Delete the event and sync.
  8. Expected: the absence is gone.

  Then check the server never logged the link:

```bash
cd server && make logs 2>&1 | grep -c 'calendar/ical' || true
```

  Expected: `0`.

  Then check the guard:
  - Paste `http://calendar.google.com/x.ics`. Expected: the "Use the secret iCal address…" message.
  - Paste a link on another host. Expected: the same message.

- [ ] **Step 3: Update the spec.** In `docs/features/calendar-sync.md`:
  - set `**Status:** Done (PR #<n>)`
  - tick every task
  - add an **As built** list:
    - `next_sync_at` on `calendar_feeds`, for claiming without holding locks during a fetch
    - leave hours rounded to two decimals
    - a moved event records `AbsenceRemoved` and `AbsenceAdded`
    - approving goes through member-calendar's `POST /absences/{id}/approve`
    - the library chosen: `github.com/arran4/golang-ical` v0.3.6
    - what Task 1 found about the tenant's links
  
  In `docs/features/README.md`, set row 13 to **Done**.

- [ ] **Step 4: Commit the docs, push, and open the PR**

```bash
git add docs/features/calendar-sync.md docs/features/README.md
git commit -m "docs: calendar sync as built" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git push -u origin feat/calendar-sync
gh pr create --base main --head feat/calendar-sync --title "Calendar sync: out-of-office events become planned leave" --body-file <(printf '%s\n' \
  "Implements docs/features/calendar-sync.md." "" \
  "- Secret ICS link (Google, Outlook), stored encrypted, never logged or returned" \
  "- Out-of-office events become planned leave that follows the calendar, with change events" \
  "- SSRF guard: https, allowed hosts, public IPs only, redirect checks, timeout, size cap" \
  "- Calendar panel on My schedule and the person page; From calendar marks" "" \
  "🤖 Generated with [Claude Code](https://claude.com/claude-code)")
```

  Wait for CI. Do not merge until the user says "merge"; then use rebase merge and keep the branch.

---

## Self-review

- **Spec coverage:**

| Spec item | Task |
|---|---|
| Connect, status, disconnect | 7 and 9 (server), 11 (client) |
| Sync every interval, and Sync now | 7 (`RunOnce`, `Run`, `SyncNow`), 9 (hook) |
| Which events count (OOF flag, keywords, not cancelled, window) | 3 |
| From event to absence (all-day, timed, multi-day, time zone) | 3 and 7 |
| Absences follow the calendar, each with a change event; manual ones untouched | 6 and 8 |
| UI marks | 10 |
| Tables | 5 |
| Security: encryption, https, allowed hosts, private IPs, redirects, timeout, size, never logging the link | 4, 7 and 9 |
| Runner with `SKIP LOCKED` and a fake fetcher in tests | 5, 7 and 9 |
| Settings | 2 |
| Acceptance 1–7 | Task 12 Step 2 (1–5), Task 7 and Task 9 RBAC (6), Task 8 (7) |
| Checking real Google and Outlook links before building | Task 1 |

- **Type consistency:** these names are the same in every task that uses them:
  - `calendar.LeaveEvent`, `calendar.Parse` and `services.CalendarLeave`, `ApplyCalendarLeaveInput`
  - `CalendarFeedRepository.RecordSync(ctx, id, at, eventsFound, errorCode)`
  - `ConnectFeedInput{PersonID, URL}`
  - the client `ConnectedCalendar` and `CalendarFeedState`
- **Deviations from the spec, on purpose:**
  - `calendar_feeds.next_sync_at` (Task 5)
  - Approve comes from member-calendar, not this feature
