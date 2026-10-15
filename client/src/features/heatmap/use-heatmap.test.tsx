import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { request } from '@/shared/api/http'
import { useHeatmap } from './use-heatmap'

vi.mock('@/shared/api/http', () => ({ request: vi.fn().mockResolvedValue({ weeks: [], rows: [] }) }))

const RANGE = { from: '2026-W42', to: '2026-W53', expertise: '' }

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

afterEach(() => {
  cleanup()
  vi.mocked(request).mockClear()
})

describe('useHeatmap', () => {
  it('asks for the team of the chosen program', async () => {
    renderHook(() => useHeatmap({ ...RANGE, programId: 'prog-1' }), { wrapper })

    await waitFor(() => expect(request).toHaveBeenCalled())
    expect(vi.mocked(request).mock.calls[0][0].params).toMatchObject({ program_id: 'prog-1' })
  })

  it('leaves the program out for all programs', async () => {
    renderHook(() => useHeatmap({ ...RANGE, programId: '' }), { wrapper })

    await waitFor(() => expect(request).toHaveBeenCalled())
    expect(vi.mocked(request).mock.calls[0][0].params?.program_id).toBeUndefined()
  })
})
