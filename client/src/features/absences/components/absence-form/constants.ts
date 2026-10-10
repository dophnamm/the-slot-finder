import { DEFAULT_HOURS_PER_DAY } from '@/utils/constants'
import type { AbsenceFormValues } from './types'

export const EMPTY_ABSENCE: AbsenceFormValues = {
  from_date: '',
  to_date: '',
  hours_per_day: String(DEFAULT_HOURS_PER_DAY),
  type: 'leave',
  status: 'planned',
}
