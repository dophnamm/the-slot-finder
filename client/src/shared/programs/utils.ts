import { can, type Role } from '@/resources/rbac'
import { formatIsoWeekSpan } from '@/utils/functions'
import { FINAL_STATUSES, PROGRESS_WARNINGS } from './constants'
import type { EditorUser, PersonRef, ProgressMetrics, ProjectDetail } from './types'

/** Planners edit everything; a lead edits what one of `leads` names them as lead of. */
export function canEdit(user: EditorUser | undefined, leads: (PersonRef | null | undefined)[]): boolean {
  if (!user) return false
  if (can(user.role as Role, 'programs:manage-all')) return true
  return user.person_id !== null && leads.some((lead) => lead?.id === user.person_id)
}

/** Whether a status still allows edits. */
export function isOpenStatus(status: string): boolean {
  return !FINAL_STATUSES.includes(status)
}

/** A project's form, phases and demand are editable by a planner or either lead while it is open. */
export function canEditProject(user: EditorUser | undefined, project: ProjectDetail): boolean {
  return isOpenStatus(project.status) && canEdit(user, [project.lead, project.program_lead])
}

/** "2026-W42 – 2026-W47" with the dates of those weeks. */
export function weekSpan(from: string, to: string): string {
  return `${from} – ${to} (${formatIsoWeekSpan(from, to)})`
}

/** Actions from a table whose `from` statuses include the current one. */
export function availableActions<A extends string, S extends string>(
  actions: Record<A, { from: readonly S[] }>,
  status: S,
): A[] {
  return (Object.keys(actions) as A[]).filter((action) => actions[action].from.includes(status))
}

/** The warnings the server raised on these metrics, in display order. */
export function activeWarnings(metrics: ProgressMetrics) {
  return PROGRESS_WARNINGS.filter((warning) => metrics[warning.flag])
}
