import { cn } from 'cn'
import { MarkdownText } from '../markdown-text'
import type { MessageBubbleProps } from './types'

export function MessageBubble({ message }: MessageBubbleProps) {
  const isUser = message.role === 'user'
  return (
    <div
      className={cn(
        'max-w-[85%] rounded-lg px-3 py-2 text-sm has-[table]:w-full has-[table]:max-w-full',
        isUser && 'whitespace-pre-wrap',
        isUser ? 'self-end bg-primary text-primary-foreground' : 'self-start bg-muted text-foreground',
      )}
    >
      {isUser ? message.text : <MarkdownText text={message.text} />}
    </div>
  )
}
