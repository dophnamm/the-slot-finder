import { ROLES } from '@/resources/rbac'
import type { PersonFormValues } from './types'

/** Hours per week a new person starts with in the form. */
export const DEFAULT_WEEKLY_HOURS = 40

/** Values of an empty form. Non-project hours left empty use the server's default. */
export const EMPTY_FORM: PersonFormValues = {
  name: '',
  email: '',
  role: ROLES.SPECIALIST,
  weekly_hours: String(DEFAULT_WEEKLY_HOURS),
  non_project_hours: '',
  start_date: '',
  end_date: '',
  active: true,
}

/** Values of the active select. */
export const ACTIVE_OPTION = { yes: 'yes', no: 'no' } as const
