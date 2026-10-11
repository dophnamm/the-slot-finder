import type { PlanChangeKind } from '@/shared/staffing'

export const CHANGE_BADGE_VARIANT: Record<PlanChangeKind, 'secondary' | 'outline' | 'destructive'> = {
  added: 'secondary',
  changed: 'outline',
  removed: 'destructive',
}
