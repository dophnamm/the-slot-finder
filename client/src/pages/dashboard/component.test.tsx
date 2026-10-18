import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { INSIGHTS_ENDPOINTS } from '@/resources/endpoints'
import { useWeekRange } from '@/shared/week-range'
import { DashboardPage } from './component'

const WEEK = '2026-W42'
const NEXT_RANGE = '8 weeks'

vi.mock('@/features/risks', () => ({ Risks: () => null }))
vi.mock('@/shared/api/http', () => ({
  request: vi.fn(({ url }: { url: string }) => {
    if (url === INSIGHTS_ENDPOINTS.capacityForecast) {
      return Promise.resolve({
        weeks: [WEEK],
        expertises: [{ expertise_id: 'pm', expertise: 'Analyst', sustained_overload: null, weeks: [] }],
      })
    }
    if (url === INSIGHTS_ENDPOINTS.heatmap) return Promise.resolve({ weeks: [], rows: [] })
    return Promise.resolve([])
  }),
}))

const renderPage = () =>
  render(
    <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
      <MemoryRouter>
        <DashboardPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )

afterEach(() => {
  cleanup()
  useWeekRange.setState({ weeks: 12 })
})

describe('DashboardPage', () => {
  it('opens on the person heatmap with the view toggle', () => {
    renderPage()

    expect(screen.getByText('Team heatmap')).toBeTruthy()
    expect(screen.getByRole('tab', { name: 'By person' }).getAttribute('aria-selected')).toBe('true')
    expect(screen.queryByText('Capacity forecast')).toBeNull()
  })

  it('switches to the forecast by expertise and keeps the week range', async () => {
    renderPage()
    fireEvent.click(screen.getByLabelText('Weeks'))
    const option = await screen.findByRole('option', { name: NEXT_RANGE })
    fireEvent.pointerDown(option)
    fireEvent.click(option)

    fireEvent.click(screen.getByRole('tab', { name: 'By expertise' }))

    expect(screen.getByText('Capacity forecast')).toBeTruthy()
    expect(screen.getByLabelText('Weeks').textContent).toBe(NEXT_RANGE)
    await waitFor(() => expect(screen.getByText('Analyst')).toBeTruthy())
  })
})
