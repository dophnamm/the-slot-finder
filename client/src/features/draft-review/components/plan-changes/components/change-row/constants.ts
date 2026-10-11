import type { PlanChangeKind } from '@/shared/staffing'
import type { BadgeVariant } from '@/shared/ui/badge'

export const CHANGE_BADGE_VARIANT: Record<PlanChangeKind, BadgeVariant> = {
  added: 'green',
  changed: 'amber',
  removed: 'red',
}
