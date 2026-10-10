import { ROLES } from './roles'
import type { Permission, Role } from './types'

/**
 * Which roles may see or use each part of the UI. The server enforces the same rules;
 * per-program rules (lead of this program, internal assistant) come from the API.
 */
export const PERMISSIONS = {
  'dashboard:view': [ROLES.PLANNER, ROLES.LEAD],
  'programs:view': [ROLES.PLANNER, ROLES.LEAD],
  'programs:manage': [ROLES.PLANNER, ROLES.LEAD],
  'intake:view': [ROLES.PLANNER, ROLES.LEAD],
  'staffing-drafts:review': [ROLES.PLANNER],
  'scenarios:view': [ROLES.PLANNER],
  'people:manage': [ROLES.PLANNER],
  'absences:approve': [ROLES.PLANNER],
  'holidays:manage': [ROLES.PLANNER],
  'events:view': [ROLES.PLANNER],
  'my-schedule:view': [ROLES.SPECIALIST],
} as const satisfies Record<string, readonly Role[]>

/** `role` is empty or missing while the user is unknown or holds no application role. */
export function can(role: Role | '' | null | undefined, permission: Permission): boolean {
  const allowed: readonly Role[] = PERMISSIONS[permission]
  return !!role && allowed.includes(role)
}
