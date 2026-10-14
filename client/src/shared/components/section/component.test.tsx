import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { Section } from './component'

const TITLE = 'People'
const BODY = 'old rows'

afterEach(cleanup)

describe('Section', () => {
  it('keeps the body and shows a thin bar while it loads again', () => {
    render(
      <Section title={TITLE} refreshing>
        {BODY}
      </Section>,
    )
    expect(screen.getByText(BODY)).toBeTruthy()
    expect(screen.getByRole('status')).toBeTruthy()
  })

  it('shows no bar otherwise', () => {
    render(<Section title={TITLE}>{BODY}</Section>)
    expect(screen.queryByRole('status')).toBeNull()
  })
})
