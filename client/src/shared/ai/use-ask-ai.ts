import { useMutation } from '@tanstack/react-query'
import { askAi } from './api'
import { AI_ERROR_CODES } from './constants'

export function useAskAi() {
  return useMutation({ mutationFn: askAi, meta: { quietErrorCodes: AI_ERROR_CODES } })
}
