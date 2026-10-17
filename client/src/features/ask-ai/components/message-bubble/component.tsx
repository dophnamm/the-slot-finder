import { cn } from 'cn'
import { MarkdownText } from '../markdown-text'
import { BUBBLE_CLASS, BUBBLE_ROLE_CLASSES } from './constants'
import type { MessageBubbleProps } from './types'

export function MessageBubble({ message }: MessageBubbleProps) {
  const isAnswer = message.role === 'assistant'
  return (
    <div
      role={message.role === 'error' ? 'alert' : undefined}
      className={cn(BUBBLE_CLASS, BUBBLE_ROLE_CLASSES[message.role])}
    >
      {isAnswer ? <MarkdownText text={message.text} /> : message.text}
    </div>
  )
}
