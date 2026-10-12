import { useTranslation } from '@/shared/i18n'
import { cn } from '@/utils/functions'
import type { FigureProps } from './types'

/** One cost figure in hours. */
export function Figure({ label, hours, warning = false }: FigureProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className={cn('text-lg font-semibold tabular-nums', warning && 'text-destructive')}>
        {t('{hours} h', { hours: Math.round(hours) })}
      </dd>
    </div>
  )
}
