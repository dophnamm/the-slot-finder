import { useCurrentUser } from '@/shared/auth'
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
  const { data: draft, isPending } = useDraft(draftId)

  if (isPending) return null
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
