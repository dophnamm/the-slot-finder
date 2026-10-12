import type { PlainKey } from '@/shared/i18n'
import type { ScenarioChange } from '@/shared/scenarios'
import { PLAN_CHANGE_BADGE, PLAN_CHANGE_LABELS } from '@/shared/staffing'
import type { BadgeVariant } from '@/shared/ui/badge'
import { formatIsoWeekSpan } from '@/utils/functions'
import { MOVED_BADGE, MOVED_LABEL } from './constants'

/** A change with the same hours in other weeks is "Moved"; the rest are added, removed or changed. */
export function changeBadge(change: ScenarioChange): { label: PlainKey; variant: BadgeVariant } {
  if (change.change === 'changed' && change.before_hours === change.after_hours)
    return { label: MOVED_LABEL, variant: MOVED_BADGE }
  return { label: PLAN_CHANGE_LABELS[change.change], variant: PLAN_CHANGE_BADGE[change.change] }
}

/** The first to last week with hours, or '' when there are none. */
export function weekSpan(from: string, to: string): string {
  return from ? formatIsoWeekSpan(from, to) : ''
}
