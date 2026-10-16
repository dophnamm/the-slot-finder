import { retryAfterSeconds } from '@/shared/api/retry-after'
import { useTranslation } from '@/shared/i18n'
import { AI_RATE_LIMITED_CODE } from './constants'
import type { AskErrorProps } from './types'

/** The line under the chat when a question failed; at the usage limit it says how long to wait. */
export function AskError({ error }: AskErrorProps) {
  const { t } = useTranslation()
  const seconds = retryAfterSeconds(error)

  let message = t('The assistant could not answer. Try again.')
  if (error.code === AI_RATE_LIMITED_CODE) {
    message = t('The assistant is at its usage limit. Try again in a minute.')
    if (seconds === 1) message = t('The assistant is at its usage limit. Try again in {seconds} second.', { seconds })
    else if (seconds !== undefined)
      message = t('The assistant is at its usage limit. Try again in {seconds} seconds.', { seconds })
  }

  return (
    <p role="alert" className="px-4 pb-2 text-xs text-destructive">
      {message}
    </p>
  )
}
