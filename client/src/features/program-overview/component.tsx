import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useProgram, useProgramOverview } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { BriefCard } from './components/brief-card'
import { ExpertiseStaffing } from './components/expertise-staffing'
import { MetricsCard } from './components/metrics-card'
import { ProgramHeader } from './components/program-header'
import type { ProgramOverviewProps } from './types'

/** Header (status, timeline, lead, actions), progress and cost figures, staffing by expertise, and the brief. */
export function ProgramOverview({ programId }: ProgramOverviewProps) {
  const { t } = useTranslation()
  const { data: program, isPending } = useProgram(programId)
  const { data: overview } = useProgramOverview(programId)

  if (isPending) return null
  if (!program) return <PagePlaceholder title={t('Program not found')} description={t('There is no program at this address.')} />

  return (
    <div className="flex flex-col gap-4">
      <ProgramHeader program={program} />
      {overview && <MetricsCard overview={overview} />}
      {overview && <ExpertiseStaffing rows={overview.staffing_by_expertise} />}
      <BriefCard program={program} />
    </div>
  )
}
