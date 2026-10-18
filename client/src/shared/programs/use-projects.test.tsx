import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { useCreateProject } from './use-projects'

const { createProject, replaceDemands } = vi.hoisted(() => ({ createProject: vi.fn(), replaceDemands: vi.fn() }))
vi.mock('./api', () => ({ createProject, replaceDemands }))

const PROJECT = { name: 'A', start_earliest: '2026-10-19', deadline: '2026-11-19', scope_points: 1, done_points: 0 }
const DEMANDS = [{ expertise_id: 'e1', phase_id: null, min_level: 3, headcount: 1, total_hours: 40 }]

function run() {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={new QueryClient()}>{children}</QueryClientProvider>
  )
  return renderHook(() => useCreateProject('g1'), { wrapper })
}

describe('useCreateProject', () => {
  it('creates the project, then saves its demands', async () => {
    createProject.mockResolvedValue({ id: 'p1' })
    replaceDemands.mockResolvedValue({ id: 'p1' })
    const { result } = run()
    result.current.mutate({ project: PROJECT, demands: DEMANDS })
    await waitFor(() => expect(result.current.data).toEqual({ project: { id: 'p1' }, demandsSaved: true }))
    expect(createProject).toHaveBeenCalledWith('g1', PROJECT)
    expect(replaceDemands).toHaveBeenCalledWith('p1', DEMANDS)
    expect(createProject.mock.invocationCallOrder[0]).toBeLessThan(replaceDemands.mock.invocationCallOrder[0])
  })

  it('reports the project without its demands when the second call fails', async () => {
    createProject.mockResolvedValue({ id: 'p1' })
    replaceDemands.mockRejectedValue(new Error('boom'))
    const { result } = run()
    result.current.mutate({ project: PROJECT, demands: DEMANDS })
    await waitFor(() => expect(result.current.data).toEqual({ project: { id: 'p1' }, demandsSaved: false }))
  })
})
