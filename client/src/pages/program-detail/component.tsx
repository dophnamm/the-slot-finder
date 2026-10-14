import { useParams } from 'react-router'
import { ProgramOverview } from '@/features/program-overview'
import { ProgramProjects } from '@/features/program-projects'
import { ProgramTeam } from '@/features/program-team'
import { ReplanBanner } from '@/features/replan-banner'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'

export function ProgramDetailPage() {
  const { id = '' } = useParams()
  const { data: user } = useCurrentUser()

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
