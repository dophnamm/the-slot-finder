import { useTranslation } from '@/shared/i18n'
import { ASSIGNMENT_KIND_BADGE, ASSIGNMENT_KIND_LABELS } from '@/shared/staffing'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import type { AlternativeRowProps } from './types'

/** One person who could take over: kind, free hours over the window, why they fit, and a Swap in button. */
export function AlternativeRow({ alternative, busy, onChoose }: AlternativeRowProps) {
  const { t } = useTranslation()
  const { person, kind, free_hours, capacity_hours, reason } = alternative

  const handleChoose = () => onChoose(person.id)

  return (
    <li className="flex items-center justify-between gap-3 rounded-lg border p-3">
      <div className="flex min-w-0 flex-col gap-0.5">
        <div className="flex flex-wrap items-center gap-2">
          <span className="font-medium">{person.name}</span>
          <Badge variant={ASSIGNMENT_KIND_BADGE[kind]}>{t(ASSIGNMENT_KIND_LABELS[kind])}</Badge>
          <span className="text-xs text-muted-foreground">
            {t('{free} of {capacity} hrs free in the window', { free: free_hours, capacity: capacity_hours })}
          </span>
        </div>
        <span className="truncate text-xs text-muted-foreground" title={reason.join(' · ')}>
          {reason.join(' · ')}
        </span>
      </div>
      <Button size="sm" disabled={busy} onClick={handleChoose}>
        {t('Swap in')}
      </Button>
    </li>
  )
}
