import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/shared/api/types'
import { TimesheetImport } from './component'

const hooks = vi.hoisted(() => ({
  download: vi.fn(),
  upload: vi.fn(),
  reset: vi.fn(),
  error: null as ApiError | null,
}))

vi.mock('./use-timesheet-import', () => ({
  useDownloadTemplate: () => ({ mutate: hooks.download, isPending: false }),
  useImportTimesheet: () => ({ mutate: hooks.upload, reset: hooks.reset, error: hooks.error, isPending: false }),
}))

const CSV = 'person_email,project_id,week,actual_hours\n'
const FILE_NAME = 'timesheet.csv'

beforeEach(() => {
  hooks.error = null
  vi.clearAllMocks()
})
afterEach(cleanup)

describe('TimesheetImport', () => {
  it('imports the chosen file', async () => {
    render(<TimesheetImport />)
    fireEvent.click(screen.getByRole('button', { name: 'Import timesheet' }))
    const file = new File([CSV], FILE_NAME, { type: 'text/csv' })
    fireEvent.change(await screen.findByLabelText('Timesheet file'), { target: { files: [file] } })
    fireEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(hooks.upload).toHaveBeenCalledWith(file, expect.anything())
  })

  it('empties the file field after a refused import, so the fixed file can be picked again', async () => {
    const refused = new ApiError({ status: 400, code: 'TIMESHEET_INVALID', message: 'TIMESHEET_INVALID' })
    hooks.upload.mockImplementation((_file: File, options: { onError: (error: ApiError) => void }) =>
      options.onError(refused),
    )
    render(<TimesheetImport />)
    fireEvent.click(screen.getByRole('button', { name: 'Import timesheet' }))
    const file = new File([CSV], FILE_NAME, { type: 'text/csv' })
    const field = await screen.findByLabelText('Timesheet file')
    fireEvent.change(field, { target: { files: [file] } })
    fireEvent.click(screen.getByRole('button', { name: 'Import' }))

    expect(screen.getByLabelText('Timesheet file')).not.toBe(field)
    expect(screen.getByRole<HTMLButtonElement>('button', { name: 'Import' }).disabled).toBe(true)
    fireEvent.change(screen.getByLabelText('Timesheet file'), { target: { files: [file] } })
    expect(screen.getByRole<HTMLButtonElement>('button', { name: 'Import' }).disabled).toBe(false)
  })

  it('downloads the template for the chosen weeks', async () => {
    render(<TimesheetImport />)
    fireEvent.click(screen.getByRole('button', { name: 'Import timesheet' }))
    fireEvent.click(await screen.findByRole('button', { name: 'Download template' }))
    expect(hooks.download).toHaveBeenCalledWith({ from: expect.any(String), to: expect.any(String) })
  })

  it('lists the row problems instead of saving', async () => {
    hooks.error = new ApiError({
      status: 400,
      code: 'TIMESHEET_INVALID',
      message: 'Some rows of the timesheet have problems; nothing was saved',
      details: {
        problems: [
          { line: 5, column: 'person_email', code: 'PERSON_UNKNOWN' },
          { line: 7, column: '', code: 'ROW_DUPLICATE' },
        ],
      },
    })
    render(<TimesheetImport />)
    fireEvent.click(screen.getByRole('button', { name: 'Import timesheet' }))
    expect(await screen.findByText('No person has this email.')).toBeTruthy()
    expect(screen.getByText('The same person, project and week appear more than once.')).toBeTruthy()
    expect(screen.getByText('person_email')).toBeTruthy()
  })
})
