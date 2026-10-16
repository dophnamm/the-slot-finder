import { NOT_FOUND_STATUS } from '@/shared/api/constants'
import { useCurrentUser } from '@/shared/auth'
import { LoadError } from '@/shared/components/load-error'
import { PageSkeleton } from '@/shared/components/page-skeleton'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'
import { canEditDraft, useDraft } from '@/shared/staffing'
import { AssignmentGrid } from './components/assignment-grid'
import { DemandCoverage } from './components/demand-coverage'
import { DraftHeader } from './components/draft-header'
import { PlanChanges } from './components/plan-changes'
import type { DraftReviewProps } from './types'

/** Review of one staffing draft: header and actions, coverage, the assignment grid and changes. */
export function DraftReview({ draftId }: DraftReviewProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: draft, error, isPending, refetch } = useDraft(draftId)

  const handleRetry = () => void refetch()

  if (isPending) return <PageSkeleton />
  if (!draft && error && error.status !== NOT_FOUND_STATUS)
    return <LoadError framed error={error} onRetry={handleRetry} />
  if (!draft)
    return (
      <PagePlaceholder
        title={t('Staffing draft not found')}
        description={t('There is no staffing draft at this address.')}
      />
    )

  const editable = canEditDraft(user, draft)

  return (
    <div className="flex flex-col gap-8">
      <DraftHeader draft={draft} editable={editable} />
      <DemandCoverage demands={draft.demands} />
      <AssignmentGrid draft={draft} editable={editable} />
      <PlanChanges changes={draft.changes} />
    </div>
  )
}
