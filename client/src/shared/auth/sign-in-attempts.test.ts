import { afterEach, describe, expect, it } from 'vitest'
import { AUTO_SIGN_IN_LIMIT, AUTO_SIGN_IN_STORAGE_KEY, AUTO_SIGN_IN_WINDOW_MS } from './constants'
import { clearAutoSignIns, recordAutoSignIn } from './sign-in-attempts'

const start = 1_000_000

afterEach(() => window.sessionStorage.clear())

describe('recordAutoSignIn', () => {
  it('allows the limit within the window, then stops', () => {
    for (let i = 0; i < AUTO_SIGN_IN_LIMIT; i++) expect(recordAutoSignIn(start + i)).toBe(true)
    expect(recordAutoSignIn(start + AUTO_SIGN_IN_LIMIT)).toBe(false)
  })

  it('allows again once the earlier tries are older than the window', () => {
    for (let i = 0; i < AUTO_SIGN_IN_LIMIT; i++) recordAutoSignIn(start)
    expect(recordAutoSignIn(start + AUTO_SIGN_IN_WINDOW_MS)).toBe(true)
  })

  it('starts over after clearAutoSignIns', () => {
    for (let i = 0; i < AUTO_SIGN_IN_LIMIT; i++) recordAutoSignIn(start)
    clearAutoSignIns()
    expect(recordAutoSignIn(start)).toBe(true)
  })

  it('ignores a damaged record', () => {
    window.sessionStorage.setItem(AUTO_SIGN_IN_STORAGE_KEY, '{oops')
    expect(recordAutoSignIn(start)).toBe(true)
  })
})
