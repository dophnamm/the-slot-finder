import { RefreshCw } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { confirm } from '@/shared/stores/confirm-store'
import { Button } from '@/shared/ui/button'
import { formatDateTime } from '@/utils/functions'
import { CALENDAR_FEED_ERROR_LABELS } from '../../constants'
import { useDisconnectCalendarFeed, useSyncCalendarFeed } from '../../use-calendar-feed'
import type { CalendarFeedStatusProps } from './types'

/** Where the calendar is, how its last sync went, and the actions on it. */
export function CalendarFeedStatus({ personId, calendar }: CalendarFeedStatusProps) {
  const { t } = useTranslation()
  const sync = useSyncCalendarFeed(personId)
  const disconnect = useDisconnectCalendarFeed(personId)

  const handleSync = () => sync.mutate()
  const handleDisconnect = async () => {
    const confirmed = await confirm({
      title: t('Disconnect this calendar?'),
      description: t('Leave that was already added stays.'),
      confirmLabel: t('Disconnect'),
      destructive: true,
    })
    if (!confirmed) return
    disconnect.mutate()
  }

  return (
    <div className="flex flex-col gap-3">
      <dl className="grid gap-x-6 gap-y-1 text-sm sm:grid-cols-[auto_1fr]">
        <dt className="text-muted-foreground">{t('Connected to')}</dt>
        <dd>{calendar.host}</dd>
        <dt className="text-muted-foreground">{t('Last sync')}</dt>
        <dd>{calendar.last_synced_at ? formatDateTime(calendar.last_synced_at) : t('Not yet')}</dd>
        <dt className="text-muted-foreground">{t('Out-of-office events')}</dt>
        <dd>{calendar.events_found}</dd>
      </dl>
      {calendar.last_error && (
        <p className="text-sm text-destructive">{t(CALENDAR_FEED_ERROR_LABELS[calendar.last_error])}</p>
      )}
      <div className="flex gap-2">
        <Button variant="outline" isLoading={sync.isPending} onClick={handleSync}>
          <RefreshCw />
          {t('Sync now')}
        </Button>
        <Button variant="ghost" isLoading={disconnect.isPending} onClick={handleDisconnect}>
          {t('Disconnect')}
        </Button>
      </div>
    </div>
  )
}
