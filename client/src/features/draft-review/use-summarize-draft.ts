import { useMutation } from '@tanstack/react-query'
import { summarizeDraft } from './api'

/** Writes the summary of one staffing draft on request. */
export function useSummarizeDraft(draftId: string) {
  return useMutation({ mutationFn: () => summarizeDraft(draftId) })
}
