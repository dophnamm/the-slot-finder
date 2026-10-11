import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { Table, TableBody, TableCell } from '@/shared/ui/table'
import { LinkRow } from './component'

function renderRow() {
  const router = createMemoryRouter(
    [
      {
        path: '/',
        element: (
          <Table>
            <TableBody>
              <LinkRow to="/detail">
                <TableCell data-testid="cell" />
                <TableCell>
                  <button type="button" data-testid="action" />
                </TableCell>
              </LinkRow>
            </TableBody>
          </Table>
        ),
      },
      { path: '/detail', element: <main data-testid="detail" /> },
    ],
    { initialEntries: ['/'] },
  )
  render(<RouterProvider router={router} />)
  return router
}

afterEach(cleanup)

describe('LinkRow', () => {
  it('opens the detail page when the row is clicked', () => {
    renderRow()
    fireEvent.click(screen.getByTestId('cell'))
    expect(screen.getByTestId('detail')).toBeTruthy()
  })

  it('leaves clicks on its own buttons alone', () => {
    const router = renderRow()
    fireEvent.click(screen.getByTestId('action'))
    expect(router.state.location.pathname).toBe('/')
  })
})
