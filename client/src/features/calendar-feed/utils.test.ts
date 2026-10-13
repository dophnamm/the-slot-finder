import { describe, expect, it } from 'vitest'
import { ApiError } from '@/shared/api/types'
import { CALENDAR_FEED_NOT_FOUND, CALENDAR_SYNC_DISABLED } from './constants'
import { stateFromError } from './utils'

const HTTP_NOT_FOUND = 404
const HTTP_UNAVAILABLE = 503
const HTTP_SERVER_ERROR = 500

function apiError(status: number, code: string) {
  return new ApiError({ status, code, message: code })
}

describe('stateFromError', () => {
  it('treats "no calendar" and "sync is off" as states', () => {
    expect(stateFromError(apiError(HTTP_NOT_FOUND, CALENDAR_FEED_NOT_FOUND))).toEqual({ kind: 'none' })
    expect(stateFromError(apiError(HTTP_UNAVAILABLE, CALENDAR_SYNC_DISABLED))).toEqual({ kind: 'disabled' })
  })

  it('leaves other errors to the shared toast', () => {
    expect(stateFromError(apiError(HTTP_SERVER_ERROR, 'INTERNAL_ERROR'))).toBeNull()
    expect(stateFromError(new Error('network'))).toBeNull()
  })
})
