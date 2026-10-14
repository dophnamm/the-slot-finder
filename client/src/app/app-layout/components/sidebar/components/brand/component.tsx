import { useTranslation } from '@/shared/i18n'
import { cn } from '@/utils/functions'
import { BRAND_CELLS } from './constants'

/** The product mark and name at the top of the sidebar. */
export function Brand() {
  const { t } = useTranslation()

  return (
    <span className="flex items-center gap-2.5">
      <span aria-hidden className="grid size-7 shrink-0 grid-cols-2 gap-0.5 rounded-lg bg-primary p-1.5">
        {BRAND_CELLS.map((isFilled, index) => (
          <span
            key={index}
            className={cn('rounded-[2px]', isFilled ? 'bg-primary-foreground' : 'bg-primary-foreground/40')}
          />
        ))}
      </span>
      <span className="truncate text-sm font-semibold">{t('The Slot Finder')}</span>
    </span>
  )
}
