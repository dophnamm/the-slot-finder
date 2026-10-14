import type { ExpertiseStaffing as ExpertiseRow } from '@/shared/programs'

export interface ExpertiseStaffingProps {
  rows: ExpertiseRow[]
  /** The overview is still loading: skeleton rows instead of the empty message. */
  isPending: boolean
}
