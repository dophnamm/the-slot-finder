import { cleanup, render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { MembersTable } from './component'

afterEach(cleanup)

describe('MembersTable', () => {
  it('is the table itself, so inside a card section it runs from edge to edge like every other table', () => {
    const { container } = render(<MembersTable members={[]} isPending={false} programId="g1" />)
    expect((container.firstElementChild as HTMLElement).dataset.slot).toBe('table-frame')
  })
})
