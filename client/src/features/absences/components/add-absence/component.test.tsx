import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { AddAbsence } from './component'

vi.mock('../../use-absences', () => ({ useCreateAbsence: () => ({ mutate: vi.fn(), isPending: false }) }))

afterEach(cleanup)

describe('AddAbsence', () => {
  it('opens the absence form in a dialog', async () => {
    render(<AddAbsence personId="p1" canApprove />)
    expect(screen.queryByLabelText(translate('From'))).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: translate('Add absence') }))
    expect(await screen.findByRole('dialog')).toBeTruthy()
    expect(screen.getByLabelText(translate('Status'))).toBeTruthy()
  })
})
