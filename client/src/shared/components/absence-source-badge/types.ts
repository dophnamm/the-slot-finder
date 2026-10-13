import type { AbsenceSource } from '@/shared/people'

export interface AbsenceSourceBadgeProps {
  source: AbsenceSource
  /** Only the icon, for inside another badge such as a calendar day's absence. */
  compact?: boolean
}
