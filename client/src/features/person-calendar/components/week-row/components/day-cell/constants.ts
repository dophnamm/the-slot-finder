import type { AbsenceType } from '@/shared/people'
import type { BadgeVariant } from '@/shared/ui/badge'

/** Approved absence and holidays by type. */
export const ABSENCE_BADGE: Record<AbsenceType, BadgeVariant> = {
  leave: 'violet',
  training: 'teal',
  holiday: 'pink',
}

/** Planned absence of any type, until a planner approves it. */
export const PLANNED_ABSENCE_BADGE: BadgeVariant = 'amber'

/** Today's cell is outlined, so the current day is easy to find. */
export const TODAY_CELL_CLASS = 'ring-2 ring-primary ring-inset'
