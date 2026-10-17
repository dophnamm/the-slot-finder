import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ProblemTable } from './component'

describe('ProblemTable', () => {
  it('shows a label for a known code and the raw code otherwise', () => {
    render(
      <ProblemTable
        problems={[
          { line: 2, column: 'email', code: 'PERSON_UNKNOWN' },
          { line: 3, column: '', code: 'SOMETHING_NEW' },
        ]}
        labels={{ PERSON_UNKNOWN: 'No person has this email.' }}
      />,
    )
    expect(screen.getByText('No person has this email.')).toBeTruthy()
    expect(screen.getByText('SOMETHING_NEW')).toBeTruthy()
  })
})
