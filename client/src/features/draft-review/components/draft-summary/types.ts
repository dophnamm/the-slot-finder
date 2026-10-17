import type { Draft } from '@/shared/staffing'
import type { DraftSummary } from '../../types'

export interface DraftSummaryProps {
  /** The summary is dropped whenever this draft changes, so a stale text never shows. */
  draft: Draft
}

/** A summary kept with the draft it was written for. */
export interface SummaryAnswer {
  draft: Draft
  result: DraftSummary
}
