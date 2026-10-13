import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query'
import { cleanup, fireEvent, render, renderHook, screen, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { CalendarFeed } from './component'
import type { CalendarFeedState, ConnectedCalendar } from './types'

const { getCalendarFeed, connectCalendarFeed, success } = vi.hoisted(() => ({
  getCalendarFeed: vi.fn(),
  connectCalendarFeed: vi.fn(),
  success: vi.fn(),
}))
vi.mock('./api', () => ({
  getCalendarFeed,
  connectCalendarFeed,
  syncCalendarFeed: vi.fn(),
  disconnectCalendarFeed: vi.fn(),
}))
vi.mock('sonner', () => ({ toast: { success } }))

const LINK = 'https://calendar.google.com/calendar/ical/a%40b/private-x/basic.ics'
const CALENDAR: ConnectedCalendar = {
  host: 'calendar.google.com',
  last_synced_at: null,
  last_error: null,
  events_found: 1,
}
/** The calendar on the same page refetches slower than the panel, as it does in the browser. */
const SLOW_REFETCH_MS = 300

function slowly() {
  return new Promise((resolve) => setTimeout(resolve, SLOW_REFETCH_MS, null))
}

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

describe('connecting a calendar through the real hooks', () => {
  it('confirms the connection even when the panel switches before the calendar refreshes', async () => {
    let state: CalendarFeedState = { kind: 'none' }
    getCalendarFeed.mockImplementation(() => Promise.resolve(state))
    connectCalendarFeed.mockImplementation(() => {
      state = { kind: 'connected', calendar: CALENDAR }
      return Promise.resolve(CALENDAR)
    })
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )

    const schedule = renderHook(() => useQuery({ queryKey: [...INSIGHTS_QUERY_KEY, 'schedule'], queryFn: slowly }), {
      wrapper,
    })
    await waitFor(() => expect(schedule.result.current.isSuccess).toBe(true))
    render(<CalendarFeed personId="p1" />, { wrapper })
    fireEvent.change(await screen.findByLabelText(translate('Secret iCal address')), { target: { value: LINK } })
    fireEvent.click(screen.getByRole('button', { name: translate('Connect calendar') }))

    expect(await screen.findByText(CALENDAR.host)).toBeTruthy()
    await waitFor(() => expect(success).toHaveBeenCalledWith(translate('Calendar connected.')))
  })
})
