import { useTranslation } from '@/shared/i18n'
import { LOADING_REVEAL_CLASS } from '@/utils/constants'
import { cn } from '@/utils/functions'
import type { LoadingBarProps } from './types'

/**
 * A thin bar running along the top edge of a card while its data loads again (another page, filter or
 * month). The data already on screen stays as it is. Its parent must be `relative`.
 */
export function LoadingBar({ className }: LoadingBarProps) {
  const { t } = useTranslation()

  return (
    <div
      data-slot="loading-bar"
      className={cn(
        'pointer-events-none absolute inset-x-0 top-0 z-30 h-0.5 overflow-hidden',
        LOADING_REVEAL_CLASS,
        className,
      )}
    >
      <div className="h-full w-1/3 animate-progress-slide bg-primary motion-reduce:animate-pulse" />
      <span className="sr-only" role="status">
        {t('Loading…')}
      </span>
    </div>
  )
}
