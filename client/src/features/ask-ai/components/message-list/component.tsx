import { useEffect, useRef } from 'react'
import { useTranslation } from '@/shared/i18n'
import { MessageBubble } from '../message-bubble'
import { TypingIndicator } from '../typing-indicator'
import type { MessageListProps } from './types'

export function MessageList({ messages, isPending, onRetry }: MessageListProps) {
  const { t } = useTranslation()
  const endRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    endRef.current?.scrollIntoView?.({ block: 'end' })
  }, [messages.length, isPending])

  return (
    <div className="flex min-h-40 flex-1 flex-col gap-2 overflow-y-auto p-4">
      {messages.length === 0 && (
        <p className="text-sm text-muted-foreground">
          {t('Hi, I can answer questions about the page you are on. What would you like to know?')}
        </p>
      )}
      {messages.map((message, index) => (
        <MessageBubble
          key={index}
          message={message}
          onRetry={index === messages.length - 1 && !isPending ? onRetry : undefined}
        />
      ))}
      {isPending && <TypingIndicator />}
      <div ref={endRef} />
    </div>
  )
}
