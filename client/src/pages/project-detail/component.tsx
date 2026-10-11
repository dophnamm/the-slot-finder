import { useParams } from 'react-router'
import { DemandForm } from '@/features/demand-form'
import { ProjectPhases } from '@/features/project-phases'
import { ProjectSummary } from '@/features/project-summary'

export function ProjectDetailPage() {
  const { id = '' } = useParams()

  return (
    <div className="flex flex-col gap-4">
      <ProjectSummary projectId={id} />
      <ProjectPhases projectId={id} />
      <DemandForm projectId={id} />
    </div>
  )
}
