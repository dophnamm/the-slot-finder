import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { TextLink } from './component'
import { TEXT_LINK_CLASS } from './constants'

const APP_LINK = 'Acme'
const OUTSIDE_LINK = 'Charter'

afterEach(cleanup)

describe('TextLink', () => {
  it('styles an app route and an outside page the same way', () => {
    render(
      <MemoryRouter>
        <TextLink to="/programs/g1">{APP_LINK}</TextLink>
        <TextLink to="https://example.test/charter" external>
          {OUTSIDE_LINK}
        </TextLink>
      </MemoryRouter>,
    )
    const app = screen.getByRole('link', { name: APP_LINK })
    expect(app.getAttribute('href')).toBe('/programs/g1')
    expect(app.className).toContain(TEXT_LINK_CLASS)
    const outside = screen.getByRole('link', { name: OUTSIDE_LINK })
    expect(outside.getAttribute('target')).toBe('_blank')
    expect(outside.getAttribute('rel')).toBe('noreferrer')
    expect(outside.className).toContain(TEXT_LINK_CLASS)
  })
})
