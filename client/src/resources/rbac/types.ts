import type { PERMISSIONS } from './permissions'
import type { ROLES } from './roles'

export type Role = (typeof ROLES)[keyof typeof ROLES]

export type Permission = keyof typeof PERMISSIONS
