import { useParams } from 'react-router'
import { ProgramOverview } from '@/features/program-overview'
import { ProgramProjects } from '@/features/program-projects'
import { ProgramTeam } from '@/features/program-team'
import { ReplanBanner } from '@/features/replan-banner'
import { can } from '@/resources/rbac'
import { PROGRAM_TABS } from '@/resources/routes'
import { FORBIDDEN_STATUS } from '@/shared/api/constants'
import { useCurrentUser } from '@/shared/auth'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { PageSkeleton } from '@/shared/components/page-skeleton'
import { useTranslation } from '@/shared/i18n'
import { useProgram } from '@/shared/programs'

/** One program. Waits for the program, so one the user may not see costs one request, not several. */
export function ProgramDetailPage() {
  const { t } = useTranslation()
  const { id = '' } = useParams()
  const { data: user } = useCurrentUser()
  const { error, isPending } = useProgram(id)

  if (isPending) return <PageSkeleton tabs={Object.keys(PROGRAM_TABS).length} />
  if (error?.status === FORBIDDEN_STATUS)
    return (
      <PagePlaceholder
        title={t('You can’t see this program')}
        description={t('A manager sees the programs they manage or lead a project in.')}
      />
    )

  return (
    <div className="flex flex-col gap-6">
      <ReplanBanner programId={id} />
      <ProgramOverview
        programId={id}
        projects={<ProgramProjects programId={id} />}
        team={<ProgramTeam programId={id} canManage={can(user?.role, 'program-teams:manage')} />}
      />
    </div>
  )
}
