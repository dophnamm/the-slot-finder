import type { Holiday } from '../../types'

export interface HolidayRowProps {
  holiday: Holiday
  canDelete: boolean
  onDelete: (id: string) => void
}
