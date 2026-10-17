import { retryAfterSeconds } from '@/shared/api/retry-after'
import type { ApiError } from '@/shared/api/types'
import { translate } from '@/shared/i18n'
import { AI_RATE_LIMITED_CODE } from './constants'

/** The text of a failed question in the chat; at the usage limit it says how long to wait. */
export function askErrorMessage(error: ApiError) {
  if (error.code !== AI_RATE_LIMITED_CODE) return translate('The assistant could not answer. Try again.')
  const seconds = retryAfterSeconds(error)
  if (seconds === undefined) return translate('The assistant is at its usage limit. Try again in a minute.')
  if (seconds === 1)
    return translate('The assistant is at its usage limit. Try again in {seconds} second.', { seconds })
  return translate('The assistant is at its usage limit. Try again in {seconds} seconds.', { seconds })
}
