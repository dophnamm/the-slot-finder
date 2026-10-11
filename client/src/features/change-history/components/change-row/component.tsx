import { generatePath, Link } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { EventSummary } from '@/shared/components/event-summary'
import { EVENT_STATUS_BADGE, EVENT_STATUS_LABELS } from '@/shared/events'
import { useTranslation } from '@/shared/i18n'
import { type PlanStatus, PLAN_STATUS_LABELS } from '@/shared/staffing'
import { Badge } from '@/shared/ui/badge'
import { TableCell, TableRow } from '@/shared/ui/table'
import { formatDateTime } from '@/utils/functions'
import type { ChangeRowProps } from './types'

export function ChangeRow({ event }: ChangeRowProps) {
  const { t } = useTranslation()

  return (
    <TableRow>
      <TableCell className="text-sm whitespace-nowrap text-muted-foreground">{formatDateTime(event.occurred_at)}</TableCell>
      <TableCell className="whitespace-normal">
        <EventSummary event={event} />
      </TableCell>
      <TableCell>
        <div className="flex flex-col gap-1">
          <Badge variant={EVENT_STATUS_BADGE[event.status]}>{t(EVENT_STATUS_LABELS[event.status])}</Badge>
          {event.last_error && (
            <span className="max-w-64 text-xs whitespace-normal text-muted-foreground">
              {t('{count} attempts: {error}', { count: event.attempts, error: event.last_error })}
            </span>
          )}
        </div>
      </TableCell>
      <TableCell>
        {event.draft && (
          <Link className="text-sm font-medium hover:underline" to={generatePath(ROUTES.staffingDraft, { id: event.draft.id })}>
            {t('Open draft ({status})', { status: t(PLAN_STATUS_LABELS[event.draft.status as PlanStatus]) })}
          </Link>
        )}
      </TableCell>
    </TableRow>
  )
}
