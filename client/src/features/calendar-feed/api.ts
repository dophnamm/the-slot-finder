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
