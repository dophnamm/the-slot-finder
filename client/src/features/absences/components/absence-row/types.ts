import type { Absence } from '../../types'

export interface AbsenceRowProps {
  absence: Absence
  onDelete: (id: string) => void
}
