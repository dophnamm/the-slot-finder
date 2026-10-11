import type { AssignmentKind } from '@/shared/staffing'

/** Badge variant per assignment kind: stretch stands out. */
export const KIND_BADGE_VARIANT: Record<AssignmentKind, 'secondary' | 'outline' | 'destructive'> = {
  primary: 'secondary',
  cross_expertise: 'outline',
  stretch: 'destructive',
}
