import { useParams } from 'react-router'
import { ScenarioReview } from '@/features/scenario-review'

export function ScenarioDetailPage() {
  const { id = '' } = useParams()

  return <ScenarioReview scenarioId={id} />
}
