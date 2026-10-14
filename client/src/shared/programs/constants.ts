import type { PlainKey } from '@/shared/i18n'
import type { BadgeVariant } from '@/shared/ui/badge'
import { MAX_PAGE_SIZE } from '@/utils/constants'
import type { ProgramAction, ProgramStatus, ProgressMetrics, ProjectAction, ProjectStatus } from './types'

export const PROGRAMS_QUERY_KEY = ['programs'] as const
export const programQueryKey = (id: string) => ['programs', id] as const
export const programProjectsQueryKey = (id: string) => ['programs', id, 'projects'] as const
export const programOverviewQueryKey = (id: string) => ['programs', id, 'overview'] as const
export const programMembersQueryKey = (id: string) => ['programs', id, 'members'] as const
export const PROJECTS_QUERY_KEY = ['projects'] as const
export const projectsQueryKey = (status: ProjectStatus | 'all') => ['projects', 'list', status] as const
export const projectQueryKey = (id: string) => ['projects', id] as const
export const openProjectsQueryKey = ['projects', 'open-options'] as const
export const LEAD_OPTIONS_QUERY_KEY = ['lead-options'] as const

/** Programs shown per page of the list. */
export const PROGRAMS_PAGE_SIZE = 20

/** Enough for every lead candidate. */
export const LEAD_OPTIONS_PAGE_SIZE = MAX_PAGE_SIZE

/** Roles that can lead a program or project. */
export const LEAD_ROLES = ['planner', 'lead'] as const

export const PROGRAM_STATUS_LABELS: Record<ProgramStatus, PlainKey> = {
  open: 'Open',
  pending: 'On hold',
  closing: 'Closing',
  closed: 'Closed',
  cancelled: 'Cancelled',
}

export const PROJECT_STATUS_LABELS: Record<ProjectStatus, PlainKey> = {
  incoming: 'Incoming',
  planned: 'Planned',
  active: 'Active',
  on_hold: 'On hold',
  done: 'Done',
  cancelled: 'Cancelled',
}

/** Program actions: their label and the statuses they start from (the server checks again). */
export const PROGRAM_ACTIONS: Record<ProgramAction, { label: PlainKey; from: readonly ProgramStatus[] }> = {
  hold: { label: 'Put on hold', from: ['open'] },
  resume: { label: 'Resume', from: ['pending'] },
  'start-closing': { label: 'Start closing', from: ['open'] },
  close: { label: 'Close', from: ['closing'] },
  cancel: { label: 'Cancel program', from: ['open', 'pending', 'closing'] },
}

export const PROJECT_ACTIONS: Record<ProjectAction, { label: PlainKey; from: readonly ProjectStatus[] }> = {
  hold: { label: 'Put on hold', from: ['incoming', 'planned', 'active'] },
  resume: { label: 'Resume', from: ['on_hold'] },
  complete: { label: 'Mark done', from: ['incoming', 'planned', 'active'] },
  cancel: { label: 'Cancel project', from: ['incoming', 'planned', 'active', 'on_hold'] },
}

/** Statuses that end a program or project: no more edits. */
export const FINAL_STATUSES: readonly string[] = ['closed', 'cancelled', 'done']

/** A warning flag of the metrics and its label, in the order they are shown. */
export const PROGRESS_WARNINGS: readonly {
  flag: keyof Pick<ProgressMetrics, 'at_risk' | 'understaffed' | 'behind_on_delivery' | 'over_planned' | 'past_end'>
  label: PlainKey
  badge: BadgeVariant
}[] = [
  { flag: 'at_risk', label: 'At risk', badge: 'red' },
  { flag: 'understaffed', label: 'Understaffed', badge: 'amber' },
  { flag: 'behind_on_delivery', label: 'Behind on delivery', badge: 'pink' },
  { flag: 'over_planned', label: 'Over-planned', badge: 'violet' },
  { flag: 'past_end', label: 'Past the end date', badge: 'red' },
]

/** Labels of the two burns: from actual hours, or from planned hours until actual hours exist. */
export const BURN_LABELS = {
  actual: 'Burn',
  planned: 'Planned burn',
} as const satisfies Record<string, PlainKey>
