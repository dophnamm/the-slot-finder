import { type ChangeEvent, OPEN_REPLAN_DRAFT_STATUS } from '@/shared/events'
import type { OpenReplan } from './types'

/** The newest open re-plan draft among the changes, with how many of them link to it. */
export function openReplan(events: ChangeEvent[]): OpenReplan | null {
  const open = events.find((event) => event.draft?.status === OPEN_REPLAN_DRAFT_STATUS)
  if (!open?.draft) return null
  const draftId = open.draft.id
  return { draftId, changes: events.filter((event) => event.draft?.id === draftId).length }
}
