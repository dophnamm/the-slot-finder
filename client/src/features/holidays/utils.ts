import { HOLIDAY_IMPORT_FIELDS } from './constants'
import type { ImportHolidaysInput } from './types'

/** The multipart form of an import: the file or the trimmed link, the dates and the hours. */
export function toImportForm(input: ImportHolidaysInput): FormData {
  const form = new FormData()
  if (input.source === 'file' && input.file) form.append(HOLIDAY_IMPORT_FIELDS.file, input.file)
  if (input.source === 'link') form.append(HOLIDAY_IMPORT_FIELDS.url, input.url.trim())
  form.append(HOLIDAY_IMPORT_FIELDS.from, input.from_date)
  form.append(HOLIDAY_IMPORT_FIELDS.to, input.to_date)
  form.append(HOLIDAY_IMPORT_FIELDS.hours, String(input.hours_per_day))
  return form
}
