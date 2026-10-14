import { can, type Role } from '@/resources/rbac'
import type { Program } from '@/shared/programs'

/**
 * The program to open instead of the list: the only one, for a role that skips a list of one. `total` is
 * the whole list's size from the server, so a last page that happens to hold one program does not count.
 */
export function onlyProgramId(role: string | undefined, programs: Program[], total: number | undefined): string | null {
  if (!can(role as Role, 'programs:open-only-program') || total !== 1 || programs.length !== 1) return null
  return programs[0].id
}
