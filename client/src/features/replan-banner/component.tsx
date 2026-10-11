import { TriangleAlert } from 'lucide-react'
import { generatePath, Link } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { useEvents } from '@/shared/events'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { BANNER_EVENTS_PAGE_SIZE } from './constants'
import type { ReplanBannerProps } from './types'
import { openReplan } from './utils'

/** Says when a change (leave, a slipped phase or deadline) produced a staffing draft for the program. */
export function ReplanBanner({ programId }: ReplanBannerProps) {
  const { t } = useTranslation()
  const { data } = useEvents(programId, 1, BANNER_EVENTS_PAGE_SIZE)
  const open = openReplan(data?.data ?? [])

  if (!open) return null

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-warning bg-warning/10 p-4">
      <div className="flex items-center gap-3">
        <TriangleAlert className="size-5 shrink-0 text-warning" />
        <div className="flex flex-col">
          <span className="font-medium">{t('A change produced a staffing draft for this program.')}</span>
          <span className="text-sm text-muted-foreground">
            {open.changes === 1
              ? t('{count} change is waiting for review.', { count: open.changes })
              : t('{count} changes are waiting for review.', { count: open.changes })}
          </span>
        </div>
      </div>
      <Button size="sm" nativeButton={false} render={<Link to={generatePath(ROUTES.staffingDraft, { id: open.draftId })} />}>
        {t('Review draft')}
      </Button>
    </div>
  )
}
