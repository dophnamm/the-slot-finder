import { describe, expect, it } from 'vitest'
import type { Program } from '@/shared/programs'
import { onlyProgramId } from './utils'

const program = (id: string) => ({ id }) as unknown as Program

describe('onlyProgramId', () => {
  it('is the program of a manager whose whole list is that one program', () => {
    expect(onlyProgramId('manager', [program('g1')], 1)).toBe('g1')
  })

  it('is null for a longer list, even on a last page holding one program', () => {
    expect(onlyProgramId('manager', [program('g1'), program('g2')], 2)).toBeNull()
    expect(onlyProgramId('manager', [program('g21')], 21)).toBeNull()
  })

  it('is null for a planner, an empty list or a list still loading', () => {
    expect(onlyProgramId('planner', [program('g1')], 1)).toBeNull()
    expect(onlyProgramId('manager', [], 0)).toBeNull()
    expect(onlyProgramId('manager', [], undefined)).toBeNull()
  })
})
