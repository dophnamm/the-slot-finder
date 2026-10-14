import { useSearchParams } from 'react-router'
import { PROGRAM_TABS, TAB_PARAM } from '@/resources/routes'
import { PageSkeleton } from '@/shared/components/page-skeleton'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useProgram, useProgramOverview } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'
import { BriefCard } from './components/brief-card'
import { ExpertiseStaffing } from './components/expertise-staffing'
import { MetricsCard } from './components/metrics-card'
import { ProgramHeader } from './components/program-header'
import type { ProgramOverviewProps } from './types'
import { programTab } from './utils'

/**
 * The header (status, timeline, manager, actions), then tabs: the overview, the projects and the team.
 * The open tab is in the URL (`?tab=`), so a link such as the breadcrumb's Projects opens it.
 */
export function ProgramOverview({ programId, projects, team }: ProgramOverviewProps) {
  const { t } = useTranslation()
  const { data: program, isPending } = useProgram(programId)
  const { data: overview } = useProgramOverview(programId)
  const [searchParams, setSearchParams] = useSearchParams()
  const tab = programTab(searchParams.get(TAB_PARAM))

  const handleTabChange = (value: string) => setSearchParams({ [TAB_PARAM]: value }, { replace: true })

  if (isPending) return <PageSkeleton tabs={Object.keys(PROGRAM_TABS).length} />
  if (!program)
    return <PagePlaceholder title={t('Program not found')} description={t('There is no program at this address.')} />

  return (
    <div className="flex flex-col gap-6">
      <ProgramHeader program={program} />
      <Tabs value={tab} onValueChange={handleTabChange} className="gap-5">
        <TabsList variant="line">
          <TabsTrigger value={PROGRAM_TABS.overview}>{t('Overview')}</TabsTrigger>
          <TabsTrigger value={PROGRAM_TABS.projects}>{t('Projects')}</TabsTrigger>
          <TabsTrigger value={PROGRAM_TABS.team}>{t('Team')}</TabsTrigger>
        </TabsList>
        <TabsContent value={PROGRAM_TABS.overview} className="flex flex-col gap-5">
          <MetricsCard overview={overview} />
          <ExpertiseStaffing rows={overview?.staffing_by_expertise ?? []} isPending={!overview} />
          <BriefCard program={program} />
        </TabsContent>
        <TabsContent value={PROGRAM_TABS.projects}>{projects}</TabsContent>
        <TabsContent value={PROGRAM_TABS.team}>{team}</TabsContent>
      </Tabs>
    </div>
  )
}
