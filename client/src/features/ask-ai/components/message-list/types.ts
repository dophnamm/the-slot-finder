import type { AskAiMessage } from '../../types'

export interface MessageListProps {
  messages: AskAiMessage[]
  isPending: boolean
  onRetry: () => void
}
