import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { PageSkeleton } from './component'

afterEach(cleanup)

describe('PageSkeleton', () => {
  it('draws the page header, tabs and a card in skeleton shapes', () => {
    const { container } = render(<PageSkeleton tabs={3} />)
    expect(screen.getByRole('status')).toBeTruthy()
    expect(container.querySelector('h1 [data-slot=skeleton]')).toBeTruthy()
    expect(container.querySelectorAll('[data-slot=skeleton]').length).toBeGreaterThanOrEqual(3 + 2 + 1)
  })
})
