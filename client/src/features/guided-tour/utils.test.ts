import { describe, expect, it } from 'vitest'
import type { PlainKey } from '@/shared/i18n'
import { NAV_DESCRIPTIONS } from './constants'
import { buildTourSteps, shouldAutoStart } from './utils'

const identity = (key: PlainKey) => key
const ALL = Object.keys(NAV_DESCRIPTIONS) as PlainKey[]

describe('buildTourSteps', () => {
  it('gives a role with fewer links fewer steps', () => {
    expect(buildTourSteps(ALL.slice(0, 2), false, identity)).toHaveLength(3)
    expect(buildTourSteps(ALL, false, identity)).toHaveLength(ALL.length + 1)
  })

  it('adds a step for the Ask AI button only for a role that can ask', () => {
    const anchors = (hasAskAi: boolean) => buildTourSteps(ALL, hasAskAi, identity).map((step) => step.element)
    expect(anchors(true)).toContain('[data-tour="ask-ai"]')
    expect(anchors(false)).not.toContain('[data-tour="ask-ai"]')
  })

  it('gives every step an anchor, a title and a description', () => {
    for (const step of buildTourSteps(ALL, true, identity)) {
      expect(step.element).toMatch(/^\[data-tour="[a-z-]+"\]$/)
      expect(step.popover?.title).toBeTruthy()
      expect(step.popover?.description).toBeTruthy()
    }
  })
})

describe('shouldAutoStart', () => {
  it.each([
    ['unseen', true, true],
    ['seen', true, false],
    ['unavailable', true, false],
    ['unseen', false, false],
  ] as const)('flag %s, anchor visible %s → %s', (flag, visible, expected) => {
    expect(shouldAutoStart(flag, visible)).toBe(expected)
  })
})
