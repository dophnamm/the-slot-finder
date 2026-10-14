import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { ImportHolidays } from './component'

const { importHolidays } = vi.hoisted(() => ({ importHolidays: vi.fn() }))
vi.mock('../../use-holidays', () => ({ useImportHolidays: () => ({ mutate: importHolidays, isPending: false }) }))

afterEach(() => {
  cleanup()
  importHolidays.mockReset()
})

async function openDialog() {
  render(<ImportHolidays />)
  fireEvent.click(screen.getByRole('button', { name: translate('Import holidays') }))
  return screen.findByRole('dialog')
}

describe('ImportHolidays', () => {
  it('imports an ICS file with this year and next year by default', async () => {
    const dialog = await openDialog()
    const file = new File(['BEGIN:VCALENDAR'], 'holidays.ics', { type: 'text/calendar' })
    fireEvent.change(screen.getByLabelText(translate('ICS file')), { target: { files: [file] } })
    fireEvent.submit(dialog.querySelector('form') as HTMLFormElement)
    const input = importHolidays.mock.calls[0][0]
    expect(input.source).toBe('file')
    expect(input.file).toBe(file)
    expect(input.from_date).toMatch(/^\d{4}-01-01$/)
    expect(input.to_date).toMatch(/^\d{4}-12-31$/)
    expect(Number(input.to_date.slice(0, 4)) - Number(input.from_date.slice(0, 4))).toBe(1)
  })

  it('imports from a calendar link on the Link tab', async () => {
    const dialog = await openDialog()
    fireEvent.click(screen.getByRole('tab', { name: translate('Link') }))
    fireEvent.change(screen.getByLabelText(translate('Calendar link')), {
      target: { value: 'https://calendar.google.com/calendar/ical/x/public/basic.ics' },
    })
    fireEvent.submit(dialog.querySelector('form') as HTMLFormElement)
    expect(importHolidays.mock.calls[0][0]).toMatchObject({
      source: 'link',
      url: 'https://calendar.google.com/calendar/ical/x/public/basic.ics',
    })
  })

  it('keeps Import off until a file or a link is given', async () => {
    await openDialog()
    const submit = screen.getAllByRole('button', { name: translate('Import holidays') }).at(-1) as HTMLButtonElement
    expect(submit.disabled).toBe(true)
  })
})
