import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useProgram, useProgramOverview } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'
import { BriefCard } from './components/brief-card'
import { ExpertiseStaffing } from './components/expertise-staffing'
import { MetricsCard } from './components/metrics-card'
import { ProgramHeader } from './components/program-header'
import { DEFAULT_PROGRAM_TAB } from './constants'
import type { ProgramOverviewProps } from './types'

/** The header (status, timeline, manager, actions), then tabs: the overview, the projects and the team. */
export function ProgramOverview({ programId, projects, team }: ProgramOverviewProps) {
  const { t } = useTranslation()
  const { data: program, isPending } = useProgram(programId)
  const { data: overview } = useProgramOverview(programId)

  if (isPending) return null
  if (!program)
    return <PagePlaceholder title={t('Program not found')} description={t('There is no program at this address.')} />

  return (
    <div className="flex flex-col gap-6">
      <ProgramHeader program={program} />
      <Tabs defaultValue={DEFAULT_PROGRAM_TAB} className="gap-5">
        <TabsList variant="line">
          <TabsTrigger value="overview">{t('Overview')}</TabsTrigger>
          <TabsTrigger value="projects">{t('Projects')}</TabsTrigger>
          <TabsTrigger value="team">{t('Team')}</TabsTrigger>
        </TabsList>
        <TabsContent value="overview" className="flex flex-col gap-5">
          {overview && <MetricsCard overview={overview} />}
          {overview && <ExpertiseStaffing rows={overview.staffing_by_expertise} />}
          <BriefCard program={program} />
        </TabsContent>
        <TabsContent value="projects">{projects}</TabsContent>
        <TabsContent value="team">{team}</TabsContent>
      </Tabs>
    </div>
  )
}
