import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useProgram } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { BriefCard } from './components/brief-card'
import { ProgramHeader } from './components/program-header'
import type { ProgramOverviewProps } from './types'

/** Header (status, timeline, lead, actions) and brief of one program. */
export function ProgramOverview({ programId }: ProgramOverviewProps) {
  const { t } = useTranslation()
  const { data: program, isPending } = useProgram(programId)

  if (isPending) return null
  if (!program) return <PagePlaceholder title={t('Program not found')} description={t('There is no program at this address.')} />

  return (
    <div className="flex flex-col gap-4">
      <ProgramHeader program={program} />
      <BriefCard program={program} />
    </div>
  )
}
