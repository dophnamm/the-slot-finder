import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { absencesQueryKey } from './constants'
import { useApproveAbsence } from './use-absences'

const { approveAbsence } = vi.hoisted(() => ({ approveAbsence: vi.fn() }))
vi.mock('./api', () => ({ approveAbsence, createAbsence: vi.fn(), deleteAbsence: vi.fn(), listAbsences: vi.fn() }))

describe('useApproveAbsence', () => {
  it('approves, then refreshes the absences and the schedule', async () => {
    approveAbsence.mockResolvedValue({ id: 'a1', status: 'approved' })
    const queryClient = new QueryClient()
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries')
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )

    const { result } = renderHook(() => useApproveAbsence('p1'), { wrapper })
    await act(() => result.current.mutateAsync('a1'))

    expect(approveAbsence).toHaveBeenCalledWith('a1')
    expect(invalidate).toHaveBeenCalledWith({ queryKey: absencesQueryKey('p1') })
    // The schedule lives under the insights prefix, so this refreshes the calendar too.
    expect(invalidate).toHaveBeenCalledWith({ queryKey: INSIGHTS_QUERY_KEY })
  })
})

describe('useApproveAbsence after a refusal', () => {
  it('refreshes the absences, so a row approved elsewhere stops offering Approve', async () => {
    approveAbsence.mockRejectedValue(new Error('ABSENCE_NOT_PLANNED'))
    const queryClient = new QueryClient()
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries')
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )

    const { result } = renderHook(() => useApproveAbsence('p1'), { wrapper })
    await act(() => result.current.mutateAsync('a1').catch(() => undefined))

    expect(invalidate).toHaveBeenCalledWith({ queryKey: absencesQueryKey('p1') })
  })
})
