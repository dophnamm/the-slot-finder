import { describe, expect, it } from 'vitest'
import { readBlobEnvelope } from './blob'

const ENVELOPE = {
  meta: { error_code: 'WEEK_RANGE_TOO_LONG', message: 'The range may cover at most 12 weeks' },
  data: null,
}
const CSV = 'person_email,week\nada@example.test,2026-W41\n'

describe('readBlobEnvelope', () => {
  it('reads a JSON envelope that arrived as a Blob', async () => {
    const blob = new Blob([JSON.stringify(ENVELOPE)], { type: 'application/json' })
    expect(await readBlobEnvelope(blob)).toEqual(ENVELOPE)
  })

  it('gives undefined for a Blob that is not JSON', async () => {
    expect(await readBlobEnvelope(new Blob([CSV]))).toBeUndefined()
  })

  it('leaves other data unchanged', async () => {
    expect(await readBlobEnvelope(ENVELOPE)).toBe(ENVELOPE)
  })
})
