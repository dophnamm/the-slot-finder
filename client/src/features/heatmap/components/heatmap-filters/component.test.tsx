import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useHeatmapFilters } from '../../use-heatmap-filters'
import { HeatmapFilters } from './component'

const PROGRAM_ID = 'prog-1'
const PROGRAM_NAME = 'Acme Commerce Platform'
const ALL_PROGRAMS = 'All programs'

vi.mock('@/shared/expertises', () => ({ useExpertises: () => ({ data: [], isPending: false }) }))
vi.mock('@/shared/programs', () => ({
  useProgramOptions: () => ({ data: [{ id: PROGRAM_ID, name: PROGRAM_NAME }], isPending: false }),
}))

afterEach(() => {
  cleanup()
  useHeatmapFilters.setState({ programId: '' })
})

describe('HeatmapFilters program select', () => {
  it('starts on all programs and lists the programs', () => {
    render(<HeatmapFilters />)

    expect((screen.getByLabelText('Program') as HTMLSelectElement).value).toBe('')
    expect(screen.getByRole('option', { name: ALL_PROGRAMS })).toBeTruthy()
    expect(screen.getByRole('option', { name: PROGRAM_NAME })).toBeTruthy()
  })

  it('keeps the chosen program in the filters, and clears it again', () => {
    render(<HeatmapFilters />)

    fireEvent.change(screen.getByLabelText('Program'), { target: { value: PROGRAM_ID } })
    expect(useHeatmapFilters.getState().programId).toBe(PROGRAM_ID)

    fireEvent.change(screen.getByLabelText('Program'), { target: { value: '' } })
    expect(useHeatmapFilters.getState().programId).toBe('')
  })
})
