import { describe, expect, it } from 'vitest'
import { can } from './permissions'

describe('specialist permissions', () => {
  it('reads programs and their teams', () => {
    expect(can('specialist', 'programs:view')).toBe(true)
    expect(can('specialist', 'program-teams:view')).toBe(true)
  })

  it('cannot manage, view people, or see the overview and demands', () => {
    expect(can('specialist', 'programs:manage')).toBe(false)
    expect(can('specialist', 'program-teams:manage')).toBe(false)
    expect(can('specialist', 'people:view')).toBe(false)
    expect(can('specialist', 'program-overview:view')).toBe(false)
    expect(can('specialist', 'project-demands:view')).toBe(false)
  })

  it('keeps the overview for planners and managers', () => {
    expect(can('planner', 'program-overview:view')).toBe(true)
    expect(can('manager', 'program-overview:view')).toBe(true)
  })
})
