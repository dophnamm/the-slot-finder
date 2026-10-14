import { describe, expect, it } from 'vitest'
import { toImportForm } from './utils'

const base = { from_date: '2026-01-01', to_date: '2027-12-31', hours_per_day: 8 }

describe('toImportForm', () => {
  it('sends a file with the dates and hours', () => {
    const file = new File(['BEGIN:VCALENDAR'], 'holidays.ics', { type: 'text/calendar' })
    const form = toImportForm({ ...base, source: 'file', file, url: '' })
    expect(form.get('file')).toBe(file)
    expect(form.get('url')).toBeNull()
    expect(form.get('from_date')).toBe('2026-01-01')
    expect(form.get('to_date')).toBe('2027-12-31')
    expect(form.get('hours_per_day')).toBe('8')
  })

  it('sends a trimmed link and no file', () => {
    const form = toImportForm({ ...base, source: 'link', file: null, url: '  https://calendar.google.com/x.ics ' })
    expect(form.get('url')).toBe('https://calendar.google.com/x.ics')
    expect(form.get('file')).toBeNull()
  })
})
