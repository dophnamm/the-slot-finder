import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { INSIGHTS_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import { useCapacityForecast } from './use-capacity-forecast'

vi.mock('@/shared/api/http', () => ({ request: vi.fn().mockResolvedValue({ weeks: [], expertises: [] }) }))

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

afterEach(() => {
  cleanup()
  vi.mocked(request).mockClear()
})

describe('useCapacityForecast', () => {
  it('asks for the week range', async () => {
    const { result } = renderHook(() => useCapacityForecast({ from: '2026-W42', to: '2026-W53' }), { wrapper })

    await waitFor(() => expect(result.current.data).toEqual({ weeks: [], expertises: [] }))
    expect(vi.mocked(request).mock.calls[0][0]).toMatchObject({
      url: INSIGHTS_ENDPOINTS.capacityForecast,
      params: { from: '2026-W42', to: '2026-W53' },
    })
  })
})
