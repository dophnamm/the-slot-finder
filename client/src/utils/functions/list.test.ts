import { describe, expect, it } from 'vitest'
import { range } from './list'

describe('range', () => {
  it('counts from zero', () => {
    expect(range(3)).toEqual([0, 1, 2])
    expect(range(0)).toEqual([])
    expect(range(-1)).toEqual([])
  })
})
