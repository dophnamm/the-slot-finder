import { useTranslation } from '@/shared/i18n'
import { cn } from '@/utils/functions'
import { TYPING_DOT_CLASSES } from './constants'

/** Three bouncing dots in an assistant bubble while the answer is on its way. */
export function TypingIndicator() {
  const { t } = useTranslation()

  return (
    <div aria-busy className="flex items-center gap-1 self-start rounded-lg bg-muted px-3 py-3">
      <span role="status" className="sr-only">
        {t('Thinking...')}
      </span>
      {TYPING_DOT_CLASSES.map((delayClass, index) => (
        <span
          key={index}
          aria-hidden
          className={cn(
            'size-1.5 animate-bounce rounded-full bg-muted-foreground motion-reduce:animate-none',
            delayClass,
          )}
        />
      ))}
    </div>
  )
}
