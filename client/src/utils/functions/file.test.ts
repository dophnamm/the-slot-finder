import { afterEach, describe, expect, it, vi } from 'vitest'
import { saveFile } from './file'

const OBJECT_URL = 'blob:http://localhost/template'
const FILE_NAME = 'timesheet-2026-W39-2026-W42.csv'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('saveFile', () => {
  it('clicks a download link for the blob, then releases its URL', () => {
    const blob = new Blob(['a,b\n'])
    URL.createObjectURL = vi.fn(() => OBJECT_URL)
    URL.revokeObjectURL = vi.fn()
    const links: HTMLAnchorElement[] = []
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(function (this: HTMLAnchorElement) {
      links.push(this)
    })

    saveFile(blob, FILE_NAME)

    expect(URL.createObjectURL).toHaveBeenCalledWith(blob)
    expect(links).toHaveLength(1)
    expect(links[0].download).toBe(FILE_NAME)
    expect(links[0].href).toBe(OBJECT_URL)
    expect(URL.revokeObjectURL).toHaveBeenCalledWith(OBJECT_URL)
  })
})
