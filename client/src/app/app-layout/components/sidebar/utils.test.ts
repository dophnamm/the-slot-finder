import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/resources/routes'
import { visibleNavItems } from './utils'

describe('visibleNavItems', () => {
  it('shows People to a lead, who may read it, but not to a specialist', () => {
    expect(visibleNavItems('lead').map((item) => item.to)).toContain(ROUTES.people)
    expect(visibleNavItems('specialist').map((item) => item.to)).not.toContain(ROUTES.people)
  })
})
