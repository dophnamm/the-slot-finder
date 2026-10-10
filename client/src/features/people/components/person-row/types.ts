import type { Person } from '../../types'

export interface PersonRowProps {
  person: Person
  onEdit: (person: Person) => void
  onShowAbsences?: (person: Person) => void
}
