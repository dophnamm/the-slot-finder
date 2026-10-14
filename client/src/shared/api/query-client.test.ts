import { afterEach, describe, expect, it, vi } from 'vitest'
import { FORBIDDEN_CODE, FORBIDDEN_STATUS } from './constants'
import { queryClient } from './query-client'
import { ApiError } from './types'

const { toastError } = vi.hoisted(() => ({ toastError: vi.fn() }))
vi.mock('sonner', () => ({ toast: { error: toastError } }))
vi.mock('@/shared/auth/session', () => ({ signOutLocally: vi.fn() }))

const forbidden = () =>
  Promise.reject(new ApiError({ status: FORBIDDEN_STATUS, code: FORBIDDEN_CODE, message: 'Insufficient permissions' }))

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
})
