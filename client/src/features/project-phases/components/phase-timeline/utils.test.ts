import { describe, expect, it } from 'vitest'
import type { TimelineRowData } from './types'
import { legendItems } from './utils'

const row = (cells: TimelineRowData['cells']): TimelineRowData => ({ key: 'r', name: 'Build', cells })

describe('legendItems', () => {
  it('explains only the states the timeline shows', () => {
    expect(legendItems([row(['planned', 'empty'])]).map((item) => item.state)).toEqual(['planned'])
    expect(legendItems([row(['vacated', 'moved']), row(['planned'])]).map((item) => item.state)).toEqual([
      'planned',
      'moved',
      'vacated',
    ])
  })
})
