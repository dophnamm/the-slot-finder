import type { Role } from '@/resources/rbac'
import type { Person } from '../../types'

export interface PersonFormDialogProps {
  open: boolean
  /** The person to edit, or null to add one. */
  person: Person | null
  onOpenChange: (open: boolean) => void
}

/** Form fields as the inputs hold them (text); converted when saving. */
export interface PersonFormValues {
  name: string
  email: string
  role: Role
  weekly_hours: string
  non_project_hours: string
  start_date: string
  end_date: string
  active: boolean
}

/** Text inputs of the form, matched by their name attribute. */
export type PersonTextField = Exclude<keyof PersonFormValues, 'role' | 'active'>
