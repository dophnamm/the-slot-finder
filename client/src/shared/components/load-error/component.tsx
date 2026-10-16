import { TriangleAlert } from 'lucide-react'
import { EmptyState } from '@/shared/components/empty-state'
import { TableEmpty } from '@/shared/components/table-empty'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import type { LoadErrorProps } from './types'

/** What to show where a request failed, instead of an empty list or "not found": the reason and a retry. */
export function LoadError({ message, onRetry, framed, columns }: LoadErrorProps) {
  const { t } = useTranslation()
  const state = {
    icon: TriangleAlert,
    title: t('Could not load this.'),
    description: message,
    framed,
    action: (
      <Button variant="outline" size="sm" onClick={onRetry}>
        {t('Try again')}
      </Button>
    ),
  }

  return columns ? <TableEmpty columns={columns} {...state} /> : <EmptyState {...state} />
}
