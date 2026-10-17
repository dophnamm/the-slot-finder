import type { Person } from '../../types'

export interface PersonRowProps {
  person: Person
  /** Without it the row has no Edit button (a manager reads the list only). */
  onEdit?: (person: Person) => void
  onShowAbsences?: (person: Person) => void
}
