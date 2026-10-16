import { TriangleAlert } from 'lucide-react'
import { EmptyState } from '@/shared/components/empty-state'
import { TableEmpty } from '@/shared/components/table-empty'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import type { LoadErrorProps } from './types'
import { useRetryWait } from './use-retry-wait'

/** What to show where a request failed, instead of an empty list or "not found": the reason and a retry. */
export function LoadError({ error, onRetry, framed, columns }: LoadErrorProps) {
  const { t } = useTranslation()
  const wait = useRetryWait(error)
  const waitText =
    wait === 1
      ? t('Too many requests. Try again in {seconds} second.', { seconds: wait })
      : t('Too many requests. Try again in {seconds} seconds.', { seconds: wait })
  const state = {
    icon: TriangleAlert,
    title: t('Could not load this.'),
    description: wait > 0 ? waitText : error.message,
    framed,
    action: (
      <Button variant="outline" size="sm" disabled={wait > 0} onClick={onRetry}>
        {t('Try again')}
      </Button>
    ),
  }

  return columns ? <TableEmpty columns={columns} {...state} /> : <EmptyState {...state} />
}
