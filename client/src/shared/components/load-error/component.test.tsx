import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/shared/api/types'
import { translate } from '@/shared/i18n'
import { LoadError } from './component'

const MESSAGE = 'Could not read from the database'
const failure = (status: number, details?: unknown) => new ApiError({ status, code: 'ERR', message: MESSAGE, details })

afterEach(() => {
  cleanup()
  vi.useRealTimers()
})

describe('LoadError', () => {
  it('says why the request failed and retries on click', () => {
    const onRetry = vi.fn()
    render(<LoadError error={failure(500)} onRetry={onRetry} />)
    expect(screen.getByText(translate('Could not load this.'))).toBeTruthy()
    expect(screen.getByText(MESSAGE)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: translate('Try again') }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })

  it('holds the retry for the seconds a 429 asks, and says so', () => {
    vi.useFakeTimers()
    const onRetry = vi.fn()
    render(<LoadError error={failure(429, { retry_after: 2 })} onRetry={onRetry} />)
    const button = screen.getByRole('button', { name: translate('Try again') }) as HTMLButtonElement
    expect(button.disabled).toBe(true)
    expect(
      screen.getByText(translate('Too many requests. Try again in {seconds} seconds.', { seconds: 2 })),
    ).toBeTruthy()

    act(() => {
      vi.advanceTimersByTime(2000)
    })
    expect(button.disabled).toBe(false)
    expect(screen.getByText(MESSAGE)).toBeTruthy()
  })
})
