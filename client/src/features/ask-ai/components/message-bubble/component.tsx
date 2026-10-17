import { cn } from 'cn'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { MarkdownText } from '../markdown-text'
import { BUBBLE_CLASS, BUBBLE_ROLE_CLASSES } from './constants'
import type { MessageBubbleProps } from './types'

export function MessageBubble({ message, onRetry }: MessageBubbleProps) {
  const { t } = useTranslation()
  const isAnswer = message.role === 'assistant'
  return (
    <div
      role={message.role === 'error' ? 'alert' : undefined}
      className={cn(BUBBLE_CLASS, BUBBLE_ROLE_CLASSES[message.role])}
    >
      {isAnswer ? <MarkdownText text={message.text} /> : message.text}
      {message.role === 'error' && onRetry && (
        <Button variant="outline" size="sm" className="mt-2 flex cursor-pointer" onClick={onRetry}>
          {t('Try again')}
        </Button>
      )}
    </div>
  )
}
