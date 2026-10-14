import { useParams } from 'react-router'
import { ProgramMembers } from '@/features/program-members'
import { ProgramOverview } from '@/features/program-overview'
import { ProgramProjects } from '@/features/program-projects'
import { ReplanBanner } from '@/features/replan-banner'

export function ProgramDetailPage() {
  const { id = '' } = useParams()

  return (
    <div className="flex flex-col gap-6">
      <ReplanBanner programId={id} />
      <ProgramOverview
        programId={id}
        projects={<ProgramProjects programId={id} />}
        members={<ProgramMembers programId={id} />}
      />
    </div>
  )
}
