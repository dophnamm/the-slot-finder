import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { pageCount } from '@/utils/functions'
import type { PagerProps } from './types'

/** "Page 2 of 5" with previous and next buttons for a paged list. */
export function Pager({ page, total, pageSize, onPageChange }: PagerProps) {
  const { t } = useTranslation()
  const pages = pageCount(total, pageSize)

  const handlePrevious = () => onPageChange(Math.max(1, page - 1))
  const handleNext = () => onPageChange(Math.min(pages, page + 1))

  return (
    <div className="flex items-center gap-2">
      <span className="text-sm text-muted-foreground">{t('Page {page} of {pages}', { page, pages })}</span>
      <Button variant="outline" size="sm" onClick={handlePrevious} disabled={page <= 1}>
        {t('Previous')}
      </Button>
      <Button variant="outline" size="sm" onClick={handleNext} disabled={page >= pages}>
        {t('Next')}
      </Button>
    </div>
  )
}
