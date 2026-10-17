import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/resources/routes'
import { visibleNavItems } from './utils'

describe('visibleNavItems', () => {
  it('shows People to a manager, who may read it, but not to a specialist', () => {
    expect(visibleNavItems('manager').map((item) => item.to)).toContain(ROUTES.people)
    expect(visibleNavItems('specialist').map((item) => item.to)).not.toContain(ROUTES.people)
  })
})
