import { BriefCard } from './components/brief-card'
import { CostCard } from './components/cost-card'
import { ProgramHeader } from './components/program-header'
import { ScopeCard } from './components/scope-card'
import { StaffingByExpertise } from './components/staffing-by-expertise'
import type { ProgramOverviewProps } from './types'

/** Header, brief, scope and cost cards, and staffing by expertise for one program. */
export function ProgramOverview({ programId }: ProgramOverviewProps) {
  return (
    <div className="flex flex-col gap-4">
      <ProgramHeader programId={programId} />
      <BriefCard />
      <div className="grid gap-4 md:grid-cols-2">
        <ScopeCard />
        <CostCard />
      </div>
      <StaffingByExpertise />
    </div>
  )
}
