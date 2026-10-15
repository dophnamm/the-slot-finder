import { afterEach, describe, expect, it, vi } from 'vitest'
import { FORBIDDEN_CODE, FORBIDDEN_STATUS, TOO_MANY_REQUESTS_STATUS } from './constants'
import { queryClient } from './query-client'
import { ApiError } from './types'

const { toastError } = vi.hoisted(() => ({ toastError: vi.fn() }))
vi.mock('sonner', () => ({ toast: { error: toastError } }))
vi.mock('@/shared/auth/session', () => ({ signOutLocally: vi.fn() }))

const forbidden = () =>
  Promise.reject(new ApiError({ status: FORBIDDEN_STATUS, code: FORBIDDEN_CODE, message: 'Insufficient permissions' }))

const tooManyRequests = (details?: unknown) =>
  new ApiError({
    status: TOO_MANY_REQUESTS_STATUS,
    code: 'RATE_LIMIT_EXCEEDED',
    message: 'Rate limit exceeded',
    details,
  })

const failWith = (error: ApiError) => () => Promise.reject(error)

afterEach(() => {
  queryClient.clear()
  toastError.mockReset()
})

describe('queryClient', () => {
  it('shows one toast for a failed query', async () => {
    await queryClient.fetchQuery({ queryKey: ['loud'], queryFn: forbidden }).catch(() => undefined)
    expect(toastError).toHaveBeenCalledOnce()
  })

  it('stays quiet for the codes a query shows itself', async () => {
    await queryClient
      .fetchQuery({ queryKey: ['quiet'], queryFn: forbidden, meta: { quietErrorCodes: [FORBIDDEN_CODE] } })
      .catch(() => undefined)
    expect(toastError).not.toHaveBeenCalled()
  })

  it.each([
    [6, 'Too many requests. Try again in 6 seconds.'],
    [1, 'Too many requests. Try again in 1 second.'],
  ])('says how long to wait after too many requests (%i s)', async (seconds, message) => {
    const error = tooManyRequests({ limit: 10, window: '1m0s', retry_after: seconds })
    await queryClient.fetchQuery({ queryKey: ['busy', seconds], queryFn: failWith(error) }).catch(() => undefined)
    expect(toastError).toHaveBeenCalledWith(message)
  })

  it('shows the server message when it does not say how long to wait', async () => {
    await queryClient.fetchQuery({ queryKey: ['busy'], queryFn: failWith(tooManyRequests()) }).catch(() => undefined)
    expect(toastError).toHaveBeenCalledWith('Rate limit exceeded')
  })

  it('does not retry a query after too many requests', () => {
    const { retry } = queryClient.getDefaultOptions().queries ?? {}
    expect(typeof retry).toBe('function')
    if (typeof retry !== 'function') return
    expect(retry(0, tooManyRequests())).toBe(false)
    expect(retry(0, new ApiError({ status: 500, code: 'INTERNAL_ERROR', message: 'Failed' }))).toBe(true)
  })
})
