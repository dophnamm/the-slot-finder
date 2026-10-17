import { ROLES } from './roles'
import type { Permission, Role } from './types'

/**
 * Which roles may see or use each part of the UI. The server enforces the same rules;
 * per-program rules (manager of this program, internal assistant) come from the API.
 */
export const PERMISSIONS = {
  'dashboard:view': [ROLES.PLANNER, ROLES.MANAGER],
  'programs:view': [ROLES.PLANNER, ROLES.MANAGER],
  'programs:manage': [ROLES.PLANNER, ROLES.MANAGER],
  'programs:manage-all': [ROLES.PLANNER],
  // A manager whose programs list holds a single program opens it instead of the list.
  'programs:open-only-program': [ROLES.MANAGER],
  'intake:view': [ROLES.PLANNER, ROLES.MANAGER],
  'staffing-drafts:view': [ROLES.PLANNER, ROLES.MANAGER],
  'staffing-drafts:manage': [ROLES.PLANNER, ROLES.MANAGER],
  'staffing-drafts:manage-all': [ROLES.PLANNER],
  'program-teams:view': [ROLES.PLANNER, ROLES.MANAGER],
  'program-teams:manage': [ROLES.PLANNER],
  'scenarios:view': [ROLES.PLANNER],
  'people:view': [ROLES.PLANNER, ROLES.MANAGER],
  'people:manage': [ROLES.PLANNER],
  'people:import': [ROLES.PLANNER],
  'timesheets:import': [ROLES.PLANNER],
  'absences:approve': [ROLES.PLANNER],
  'calendar-feeds:manage': [ROLES.PLANNER],
  'holidays:manage': [ROLES.PLANNER],
  'ai:ask': [ROLES.PLANNER, ROLES.MANAGER],
  'events:view': [ROLES.PLANNER],
  'my-schedule:view': [ROLES.SPECIALIST],
} as const satisfies Record<string, readonly Role[]>

/** `role` is empty or missing while the user is unknown or holds no application role. */
export function can(role: Role | '' | null | undefined, permission: Permission): boolean {
  const allowed: readonly Role[] = PERMISSIONS[permission]
  return !!role && allowed.includes(role)
}
