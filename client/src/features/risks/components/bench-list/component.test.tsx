import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { BenchList } from './component'

const mocks = vi.hoisted(() => ({ useBench: vi.fn() }))
vi.mock('../../use-bench', () => ({ useBench: mocks.useBench }))

afterEach(cleanup)

const renderList = () =>
  render(
    <MemoryRouter>
      <BenchList />
    </MemoryRouter>,
  )

describe('BenchList', () => {
  it('lists people with their free hours', () => {
    mocks.useBench.mockReturnValue({
      data: { weeks: 4, people: [{ person_id: 'p1', name: 'Quinn', primary_expertise: 'QA', free_hours: 104 }] },
      isPending: false,
    })
    renderList()
    expect(screen.getByText('Bench (1)')).toBeTruthy()
    expect(screen.getByRole('link', { name: 'Quinn' }).getAttribute('href')).toBe('/people/p1')
    expect(screen.getByText('QA')).toBeTruthy()
    expect(screen.getByText('104 h free in 4 weeks')).toBeTruthy()
  })

  it('shows an empty state', () => {
    mocks.useBench.mockReturnValue({ data: { weeks: 4, people: [] }, isPending: false })
    renderList()
    expect(screen.getByText('Nobody is on the bench')).toBeTruthy()
  })

  it('shows the load error', () => {
    mocks.useBench.mockReturnValue({
      data: undefined,
      error: new Error('boom'),
      isPending: false,
      refetch: vi.fn(),
    })
    renderList()
    expect(screen.getByRole('button', { name: /retry|try again/i })).toBeTruthy()
  })
})
