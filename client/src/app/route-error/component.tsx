import { TriangleAlert } from 'lucide-react'
import { EmptyState } from '@/shared/components/empty-state'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { cn } from '@/utils/functions'
import type { RouteErrorProps } from './types'

/** Shown by the router when a screen throws while rendering, so the user gets a way out instead of a blank page. */
export function RouteError({ fullPage }: RouteErrorProps) {
  const { t } = useTranslation()

  const handleReload = () => window.location.reload()

  return (
    <div className={cn(fullPage && 'flex min-h-svh items-center justify-center bg-canvas p-4')}>
      <EmptyState
        framed
        icon={TriangleAlert}
        title={t('Something went wrong.')}
        description={t('This screen hit an unexpected error. Reload the page; if it happens again, tell an admin.')}
        action={
          <Button variant="outline" size="sm" onClick={handleReload}>
            {t('Reload page')}
          </Button>
        }
      />
    </div>
  )
}
