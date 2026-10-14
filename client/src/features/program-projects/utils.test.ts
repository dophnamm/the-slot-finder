import { describe, expect, it } from 'vitest'
import type { Project } from '@/shared/programs'
import { moveBy, moveTo, orderedProjects } from './utils'

const project = (id: string) => ({ id }) as unknown as Project

describe('moveTo', () => {
  it('puts the dragged project where the row under it was', () => {
    expect(moveTo(['a', 'b', 'c'], 'c', 'a')).toEqual(['c', 'a', 'b'])
    expect(moveTo(['a', 'b', 'c'], 'a', 'c')).toEqual(['b', 'c', 'a'])
    expect(moveTo(['a', 'b', 'c'], 'b', 'b')).toEqual(['a', 'b', 'c'])
  })
})

describe('moveBy', () => {
  it('moves one place up or down and stops at the ends', () => {
    expect(moveBy(['a', 'b', 'c'], 'b', -1)).toEqual(['b', 'a', 'c'])
    expect(moveBy(['a', 'b', 'c'], 'b', 1)).toEqual(['a', 'c', 'b'])
    expect(moveBy(['a', 'b', 'c'], 'a', -1)).toEqual(['a', 'b', 'c'])
    expect(moveBy(['a', 'b', 'c'], 'c', 1)).toEqual(['a', 'b', 'c'])
  })
})

describe('orderedProjects', () => {
  it('follows a pending order, or keeps the server order without one', () => {
    const projects = [project('a'), project('b'), project('c')]
    expect(orderedProjects(projects, ['c', 'a', 'b']).map((p) => p.id)).toEqual(['c', 'a', 'b'])
    expect(orderedProjects(projects, null).map((p) => p.id)).toEqual(['a', 'b', 'c'])
  })
})
