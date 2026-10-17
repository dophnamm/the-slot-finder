import { useMutation } from '@tanstack/react-query'
import { askAi } from './api'

/** A failed question is shown in the chat, so the shared error toast stays quiet. */
export function useAskAi() {
  return useMutation({ mutationFn: askAi, meta: { quietErrors: true } })
}
