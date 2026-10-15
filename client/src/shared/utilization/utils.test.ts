import { describe, expect, it } from 'vitest'
import type { UtilizationBand, UtilizationBands } from './types'
import { utilizationBand } from './utils'

const BANDS: UtilizationBands = { underused_below: 0.5, near_full_from: 0.85 }

const CASES: { name: string; load: number; capacity: number; band: UtilizationBand }[] = [
  { name: 'no capacity and no load', load: 0, capacity: 0, band: 'unavailable' },
  { name: 'no capacity but some load', load: 4, capacity: 0, band: 'overloaded' },
  { name: 'under half of capacity', load: 10, capacity: 40, band: 'underused' },
  { name: 'exactly the underused limit', load: 20, capacity: 40, band: 'healthy' },
  { name: 'just under the near full limit', load: 33, capacity: 40, band: 'healthy' },
  { name: 'exactly the near full limit', load: 34, capacity: 40, band: 'near_full' },
  { name: 'exactly full', load: 40, capacity: 40, band: 'near_full' },
  { name: 'over capacity', load: 41, capacity: 40, band: 'overloaded' },
]

describe('utilizationBand', () => {
  it.each(CASES)('$name', ({ load, capacity, band }) => {
    expect(utilizationBand(load, capacity, BANDS)).toBe(band)
  })
})
