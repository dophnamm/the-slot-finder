import { afterEach, describe, expect, it, vi } from 'vitest'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { getPerson, getSchedule } from './api'
import { PEOPLE_QUERY_KEY, personQueryKey, scheduleQueryKey } from './constants'

const { request } = vi.hoisted(() => ({ request: vi.fn() }))
vi.mock('@/shared/api/http', () => ({ request }))

afterEach(() => {
  request.mockReset()
})

describe('people api', () => {
  it('asks for a schedule by person and week range', async () => {
    await getSchedule('p1', '2026-W40', '2026-W44')
    expect(request).toHaveBeenCalledWith({ url: '/people/p1/schedule', params: { from: '2026-W40', to: '2026-W44' } })
  })

  it('reads one person', async () => {
    await getPerson('p1')
    expect(request).toHaveBeenCalledWith({ url: '/people/p1' })
  })
})

describe('query keys', () => {
  it('keeps a schedule under insights, so accepting a staffing draft refreshes it', () => {
    expect(scheduleQueryKey('p1', '2026-W40', '2026-W44').slice(0, INSIGHTS_QUERY_KEY.length)).toEqual([
      ...INSIGHTS_QUERY_KEY,
    ])
  })

  it('keeps a person under people, so saving a person refreshes the page header', () => {
    expect(personQueryKey('p1').slice(0, PEOPLE_QUERY_KEY.length)).toEqual([...PEOPLE_QUERY_KEY])
  })
})
