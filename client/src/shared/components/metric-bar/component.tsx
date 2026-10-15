import { Term } from '@/shared/components/term'
import { useTranslation } from '@/shared/i18n'
import { Progress } from '@/shared/ui/progress'
import { cn } from '@/utils/functions'
import { FULL_PERCENT } from './constants'
import type { MetricBarProps } from './types'

/** A labelled percentage with its bar; red when the server raised a warning, "—" when unknown. */
export function MetricBar({ label, term, value, warning = false, compact = false }: MetricBarProps) {
  const { t } = useTranslation()
  const shown = value === null ? t('—') : t('{percent}%', { percent: Math.round(value) })

  return (
    <div className={cn('flex min-w-24 flex-col', compact ? 'gap-0.5 text-xs' : 'gap-1 text-sm')}>
      <div className="flex items-center justify-between gap-2">
        <span className="min-w-0 text-muted-foreground">
          <Term term={term}>{label}</Term>
        </span>
        <span className={cn('shrink-0 font-medium tabular-nums', warning && 'text-destructive')}>{shown}</span>
      </div>
      <Progress
        value={value === null ? 0 : Math.min(value, FULL_PERCENT)}
        variant={warning ? 'destructive' : 'default'}
        aria-label={label}
      />
    </div>
  )
}
