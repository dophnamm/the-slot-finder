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
  it('starts on all programs and lists the programs', async () => {
    render(<HeatmapFilters />)

    expect(screen.getByLabelText('Program').textContent).toBe(ALL_PROGRAMS)
    fireEvent.click(screen.getByLabelText('Program'))
    expect(await screen.findByRole('option', { name: ALL_PROGRAMS })).toBeTruthy()
    expect(screen.getByRole('option', { name: PROGRAM_NAME })).toBeTruthy()
  })

  it('keeps the chosen program in the filters, and clears it again', async () => {
    render(<HeatmapFilters />)

    fireEvent.click(screen.getByLabelText('Program'))
    const option = await screen.findByRole('option', { name: PROGRAM_NAME })
    fireEvent.pointerDown(option)
    fireEvent.click(option)
    expect(useHeatmapFilters.getState().programId).toBe(PROGRAM_ID)

    fireEvent.click(screen.getByLabelText('Program'))
    const allOption = await screen.findByRole('option', { name: ALL_PROGRAMS })
    fireEvent.pointerDown(allOption)
    fireEvent.click(allOption)
    expect(useHeatmapFilters.getState().programId).toBe('')
  })
})
