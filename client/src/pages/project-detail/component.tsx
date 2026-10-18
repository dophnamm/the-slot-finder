import { useParams } from 'react-router'
import { DemandForm } from '@/features/demand-form'
import { ProjectPhases } from '@/features/project-phases'
import { ProjectSummary } from '@/features/project-summary'
import { ReplanBanner } from '@/features/replan-banner'
import { can } from '@/resources/rbac'
import { FORBIDDEN_STATUS } from '@/shared/api/constants'
import { useCurrentUser } from '@/shared/auth'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { PageSkeleton } from '@/shared/components/page-skeleton'
import { useTranslation } from '@/shared/i18n'
import { useProject } from '@/shared/programs'

/** One project. Waits for the project, so one the user may not see costs one request, not several. */
export function ProjectDetailPage() {
  const { t } = useTranslation()
  const { id = '' } = useParams()
  const { data: user } = useCurrentUser()
  const { data: project, error, isPending } = useProject(id)

  if (isPending) return <PageSkeleton />
  if (error?.status === FORBIDDEN_STATUS)
    return (
      <PagePlaceholder
        title={t('You can’t see this project')}
        description={
          can(user?.role, 'project-demands:view')
            ? t('A manager sees the projects of the programs they manage or lead a project in.')
            : t('Ask a planner if you need access.')
        }
      />
    )

  return (
    <div className="flex flex-col gap-8">
      {project && can(user?.role, 'programs:manage') && <ReplanBanner programId={project.program_id} />}
      <ProjectSummary projectId={id} />
      <ProjectPhases projectId={id} />
      {can(user?.role, 'project-demands:view') && <DemandForm projectId={id} />}
    </div>
  )
}
