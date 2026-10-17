import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { toast } from 'sonner'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/shared/api/types'
import { PeopleImport } from './component'

const hooks = vi.hoisted(() => ({
  download: vi.fn(),
  upload: vi.fn(),
  reset: vi.fn(),
  error: null as ApiError | null,
}))

vi.mock('sonner', () => ({ toast: { success: vi.fn() } }))
vi.mock('./use-people-import', () => ({
  useDownloadPeopleTemplate: () => ({ mutate: hooks.download, isPending: false }),
  useImportPeople: () => ({ mutate: hooks.upload, reset: hooks.reset, error: hooks.error, isPending: false }),
}))

const FILE_NAME = 'people.csv'

beforeEach(() => {
  hooks.error = null
  vi.clearAllMocks()
})
afterEach(cleanup)

describe('PeopleImport', () => {
  it('opens a dialog with the template, a file picker and Import', async () => {
    render(<PeopleImport />)
    fireEvent.click(screen.getByRole('button', { name: 'Import people' }))
    fireEvent.click(await screen.findByRole('button', { name: 'Download template' }))
    expect(hooks.download).toHaveBeenCalled()
    expect(screen.getByLabelText('People file')).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Import' })).toBeTruthy()
  })

  it('lists the problems with line, column and a readable label', async () => {
    hooks.error = new ApiError({
      status: 400,
      code: 'PEOPLE_IMPORT_INVALID',
      message: 'PEOPLE_IMPORT_INVALID',
      details: { problems: [{ line: 5, column: 'email', code: 'EMAIL_INVALID' }] },
    })
    render(<PeopleImport />)
    fireEvent.click(screen.getByRole('button', { name: 'Import people' }))
    expect(await screen.findByText('5')).toBeTruthy()
    expect(screen.getByText('email')).toBeTruthy()
    expect(screen.getByText('Not a valid email address.')).toBeTruthy()
  })

  it('shows the counts in a toast on success', async () => {
    hooks.upload.mockImplementation(
      (
        _file: File,
        options: { onSuccess: (result: { created: number; updated: number; unchanged: number }) => void },
      ) => options.onSuccess({ created: 3, updated: 2, unchanged: 14 }),
    )
    render(<PeopleImport />)
    fireEvent.click(screen.getByRole('button', { name: 'Import people' }))
    const file = new File(['name\n'], FILE_NAME, { type: 'text/csv' })
    fireEvent.change(await screen.findByLabelText('People file'), { target: { files: [file] } })
    fireEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(toast.success).toHaveBeenCalledWith('Created 3, updated 2, unchanged 14')
  })
})
