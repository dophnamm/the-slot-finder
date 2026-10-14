import { toast } from 'sonner'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { Skeleton } from '@/shared/ui/skeleton'
import { LOADING_REVEAL_CLASS } from '@/utils/constants'
import { cn } from '@/utils/functions'
import { CalendarFeedForm } from './components/calendar-feed-form'
import { CalendarFeedStatus } from './components/calendar-feed-status'
import type { CalendarFeedProps } from './types'
import { useCalendarFeed, useConnectCalendarFeed } from './use-calendar-feed'

/**
 * A person's Google or Outlook calendar: connect it, or see how its last sync went. The connect
 * mutation lives here, not in the form: the form unmounts once the panel shows the calendar, and
 * a mutation's own callbacks are dropped when its component is gone.
 */
export function CalendarFeed({ personId }: CalendarFeedProps) {
  const { t } = useTranslation()
  const { data: state, isPending } = useCalendarFeed(personId)
  const connect = useConnectCalendarFeed(personId)

  const handleConnected = () => toast.success(t('Calendar connected.'))
  const handleConnect = (link: string) => connect.mutate(link, { onSuccess: handleConnected })

  return (
    <Section
      title={t('Calendar')}
      description={t(
        'Out-of-office events in a Google or Outlook calendar become planned leave for a planner to approve.',
      )}
      card
    >
      {isPending && (
        <div aria-busy className={cn('flex flex-col gap-3', LOADING_REVEAL_CLASS)}>
          <Skeleton className="h-3 w-24" />
          <Skeleton className="h-8 w-full max-w-xl" />
          <Skeleton className="h-8 w-36" />
          <span className="sr-only" role="status">
            {t('Loading…')}
          </span>
        </div>
      )}
      {state?.kind === 'disabled' && (
        <p className="text-sm text-muted-foreground">{t('Calendar sync is turned off on this server.')}</p>
      )}
      {state?.kind === 'none' && (
        <CalendarFeedForm
          personId={personId}
          onConnect={handleConnect}
          pending={connect.isPending}
          error={connect.error}
        />
      )}
      {state?.kind === 'connected' && <CalendarFeedStatus personId={personId} calendar={state.calendar} />}
    </Section>
  )
}
