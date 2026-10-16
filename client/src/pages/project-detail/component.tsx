import { useParams } from 'react-router'
import { DemandForm } from '@/features/demand-form'
import { ProjectPhases } from '@/features/project-phases'
import { ProjectSummary } from '@/features/project-summary'
import { ReplanBanner } from '@/features/replan-banner'
import { FORBIDDEN_STATUS } from '@/shared/api/constants'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { PageSkeleton } from '@/shared/components/page-skeleton'
import { useTranslation } from '@/shared/i18n'
import { useProject } from '@/shared/programs'

/** One project. Waits for the project, so one the user may not see costs one request, not several. */
export function ProjectDetailPage() {
  const { t } = useTranslation()
  const { id = '' } = useParams()
  const { data: project, error, isPending } = useProject(id)

  if (isPending) return <PageSkeleton />
  if (error?.status === FORBIDDEN_STATUS)
    return (
      <PagePlaceholder
        title={t('You can’t see this project')}
        description={t('A manager sees the projects of the programs they manage or lead a project in.')}
      />
    )

  return (
    <div className="flex flex-col gap-8">
      {project && <ReplanBanner programId={project.program_id} />}
      <ProjectSummary projectId={id} />
      <ProjectPhases projectId={id} />
      <DemandForm projectId={id} />
    </div>
  )
}
