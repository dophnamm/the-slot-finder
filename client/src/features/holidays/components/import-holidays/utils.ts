import { DATE_FORMAT, DEFAULT_HOURS_PER_DAY } from '@/utils/constants'
import { dayjs } from '@/utils/functions'
import type { ImportHolidaysInput } from '../../types'
import { IMPORT_YEARS_AHEAD } from './constants'
import type { ImportHolidaysValues } from './types'

/** A file to pick, this year and next year, and a full working day. */
export function initialValues(): ImportHolidaysValues {
  return {
    source: 'file',
    file: null,
    url: '',
    from_date: dayjs().startOf('year').format(DATE_FORMAT.isoDate),
    to_date: dayjs().add(IMPORT_YEARS_AHEAD, 'year').endOf('year').format(DATE_FORMAT.isoDate),
    hours_per_day: String(DEFAULT_HOURS_PER_DAY),
  }
}

/** A file or a link for the chosen source, both dates in order, and positive hours. */
export function canImport(values: ImportHolidaysValues): boolean {
  const hasSource = values.source === 'file' ? values.file !== null : values.url.trim() !== ''
  return (
    hasSource &&
    values.from_date !== '' &&
    values.to_date !== '' &&
    values.to_date >= values.from_date &&
    Number(values.hours_per_day) > 0
  )
}

export function toInput(values: ImportHolidaysValues): ImportHolidaysInput {
  return { ...values, hours_per_day: Number(values.hours_per_day) }
}
