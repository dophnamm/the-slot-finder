import { type ChangeEvent, type FormEvent, useState } from 'react'
import { toast } from 'sonner'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { CALENDAR_FEED_ERROR_LABELS } from '../../constants'
import { useConnectCalendarFeed } from '../../use-calendar-feed'
import type { CalendarFeedFormProps } from './types'
import { linkProblem } from './utils'

/** Paste the secret iCal address; the server checks it and syncs it once. */
export function CalendarFeedForm({ personId }: CalendarFeedFormProps) {
  const { t } = useTranslation()
  const [link, setLink] = useState('')
  const connect = useConnectCalendarFeed(personId)
  const problem = linkProblem(connect.error)
  const inputId = `calendar-link-${personId}`

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => setLink(event.target.value)
  const handleConnected = () => {
    toast.success(t('Calendar connected.'))
    setLink('')
  }
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    connect.mutate(link.trim(), { onSuccess: handleConnected })
  }

  return (
    <form className="flex flex-col gap-3" onSubmit={handleSubmit}>
      <div className="flex flex-col gap-2">
        <Label htmlFor={inputId}>{t('Secret iCal address')}</Label>
        <Input
          id={inputId}
          type="url"
          value={link}
          onChange={handleChange}
          autoComplete="off"
          aria-invalid={problem !== null}
          required
        />
        <p className="text-sm text-muted-foreground">
          {t('Google: Settings and sharing → Integrate calendar → Secret address in iCal format.')}
        </p>
        <p className="text-sm text-muted-foreground">
          {t('Outlook: Settings → Calendar → Shared calendars → Publish a calendar → ICS link.')}
        </p>
        {problem && <p className="text-sm text-destructive">{t(CALENDAR_FEED_ERROR_LABELS[problem])}</p>}
      </div>
      <Button type="submit" className="self-start" disabled={connect.isPending}>
        {t('Connect calendar')}
      </Button>
    </form>
  )
}
