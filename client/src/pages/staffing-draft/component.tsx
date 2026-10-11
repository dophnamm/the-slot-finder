import { useParams } from 'react-router'
import { DraftReview } from '@/features/draft-review'

export function StaffingDraftPage() {
  const { id = '' } = useParams()

  return <DraftReview draftId={id} />
}
