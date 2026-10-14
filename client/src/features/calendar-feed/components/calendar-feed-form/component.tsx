import { type ChangeEvent, type FormEvent, useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { CALENDAR_FEED_ERROR_LABELS } from '../../constants'
import type { CalendarFeedFormProps } from './types'
import { linkProblem } from './utils'

/** Paste the secret iCal address; the server checks it and syncs it once. */
export function CalendarFeedForm({ personId, onConnect, pending, error }: CalendarFeedFormProps) {
  const { t } = useTranslation()
  const [link, setLink] = useState('')
  const problem = linkProblem(error)
  const inputId = `calendar-link-${personId}`

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => setLink(event.target.value)
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    onConnect(link.trim())
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
          placeholder={t('e.g. https://calendar.google.com/calendar/ical/…/private-…/basic.ics')}
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
      <Button type="submit" className="self-start" isLoading={pending}>
        {t('Connect calendar')}
      </Button>
    </form>
  )
}
