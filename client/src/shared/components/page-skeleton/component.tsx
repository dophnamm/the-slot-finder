import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { Skeleton } from '@/shared/ui/skeleton'
import { LOADING_REVEAL_CLASS } from '@/utils/constants'
import { cn, range } from '@/utils/functions'
import { PAGE_SKELETON_FACTS } from './constants'
import type { PageSkeletonProps } from './types'

/** A record page while it loads: its header, optional tabs and a card of facts, in skeleton shapes. */
export function PageSkeleton({ tabs = 0 }: PageSkeletonProps) {
  const { t } = useTranslation()

  return (
    <div aria-busy className={cn('flex flex-col gap-6', LOADING_REVEAL_CLASS)}>
      <Section
        variant="page"
        title={<Skeleton className="my-1 h-6 w-64 max-w-full" />}
        description={<Skeleton className="h-4 w-96 max-w-full" />}
      />
      {tabs > 0 && (
        <div className="flex gap-6">
          {range(tabs).map((tab) => (
            <Skeleton key={tab} className="h-4 w-20" />
          ))}
        </div>
      )}
      <Section card title={<Skeleton className="my-0.5 h-4 w-40" />}>
        <div className="grid gap-4 sm:grid-cols-3">
          {range(PAGE_SKELETON_FACTS).map((fact) => (
            <div key={fact} className="flex flex-col gap-2">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="h-4 w-40 max-w-full" />
            </div>
          ))}
        </div>
      </Section>
      <span className="sr-only" role="status">
        {t('Loading…')}
      </span>
    </div>
  )
}
