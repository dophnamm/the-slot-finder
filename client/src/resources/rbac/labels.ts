import type { PlainKey } from '@/shared/i18n'
import { ROLES } from './roles'
import type { Role } from './types'

/** Display label of each role, for t(). */
export const ROLE_LABELS: Record<Role, PlainKey> = {
  [ROLES.PLANNER]: 'Planner',
  [ROLES.LEAD]: 'Manager',
  [ROLES.SPECIALIST]: 'Specialist',
}
