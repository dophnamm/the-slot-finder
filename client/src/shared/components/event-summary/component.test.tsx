import { cleanup, render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import type { ChangeEvent, ChangeEventPayload, ChangeEventType } from '@/shared/events'
import { EventSummary } from './component'

afterEach(cleanup)

function textOf(type: ChangeEventType, payload: ChangeEventPayload) {
  const event = {
    id: '1',
    type,
    payload,
    occurred_at: '',
    status: 'pending',
    attempts: 0,
    last_error: null,
    project_ids: [],
    draft: null,
  } satisfies ChangeEvent
  return render(<EventSummary event={event} />).container.textContent
}

describe('EventSummary', () => {
  it.each([
    ['SkillChanged', { person_name: 'Quinn' }, 'Quinn: skills changed'],
    [
      'ProgramSlipped',
      { program: 'Checkout Revamp', old_end_date: '2026-10-30', new_end_date: '2026-11-13' },
      'Checkout Revamp: end moved from 2026-W44 to 2026-W46',
    ],
    ['DemandChanged', { project: 'Checkout Revamp' }, 'Checkout Revamp: demand changed'],
    [
      'PersonCapacityChanged',
      { person_name: 'Quinn', fields: ['weekly_hours', 'bogus', 'end_date'] },
      'Quinn: weekly hours, end date changed',
    ],
    [
      'ProjectStatusChanged',
      { name: 'Checkout Revamp', old_status: 'planned', new_status: 'active' },
      'Checkout Revamp: Planned → Active',
    ],
    [
      'TeamChanged',
      { person_name: 'Quinn', program: 'Checkout Revamp', change: 'added' },
      'Quinn added to the Checkout Revamp team',
    ],
    [
      'TeamChanged',
      { person_name: 'Quinn', program: 'Checkout Revamp', change: 'removed' },
      'Quinn removed from the Checkout Revamp team',
    ],
    ['TeamChanged', { change: 'added' }, 'Team member added'],
    ['TeamChanged', { change: 'removed' }, 'Team member removed'],
  ] as [ChangeEventType, ChangeEventPayload, string][])('%s renders its line', (type, payload, expected) => {
    expect(textOf(type, payload)).toBe(expected)
  })

  it('renders without payload fields', () => {
    expect(() => textOf('ProgramSlipped', {})).not.toThrow()
  })
})
