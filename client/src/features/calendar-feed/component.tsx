import { toast } from 'sonner'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
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
      {isPending && <p className="text-sm text-muted-foreground">{t('Loading…')}</p>}
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
