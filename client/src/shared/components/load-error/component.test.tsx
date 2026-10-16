import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { LoadError } from './component'

const MESSAGE = 'Could not read from the database'

afterEach(cleanup)

describe('LoadError', () => {
  it('says why the request failed and retries on click', () => {
    const onRetry = vi.fn()
    render(<LoadError message={MESSAGE} onRetry={onRetry} />)
    expect(screen.getByText(translate('Could not load this.'))).toBeTruthy()
    expect(screen.getByText(MESSAGE)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: translate('Try again') }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })
})
