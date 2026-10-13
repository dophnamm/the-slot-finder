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

  it('shows what a secret link looks like', () => {
    showState({ kind: 'none' })
    render(<CalendarFeed personId={PERSON_ID} />)
    expect(screen.getByLabelText(translate('Secret iCal address')).getAttribute('placeholder')).toBe(
      translate('e.g. https://calendar.google.com/calendar/ical/…/private-…/basic.ics'),
    )
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
