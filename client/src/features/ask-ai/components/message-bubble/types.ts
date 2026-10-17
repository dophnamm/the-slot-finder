import type { AskAiMessage } from '../../types'

export interface MessageBubbleProps {
  message: AskAiMessage
  /** Set only on the last message when it is an error: shows the Try again button. */
  onRetry?: () => void
}
