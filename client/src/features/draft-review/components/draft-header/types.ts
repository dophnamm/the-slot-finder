import type { Draft } from '@/shared/staffing'

export interface DraftHeaderProps {
  draft: Draft
  /** The signed-in user may accept or reject this draft. */
  editable: boolean
}
