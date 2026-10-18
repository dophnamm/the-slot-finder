import { useSearchParams } from 'react-router'
import { can } from '@/resources/rbac'
import { PROGRAM_TABS, TAB_PARAM } from '@/resources/routes'
import { NOT_FOUND_STATUS } from '@/shared/api/constants'
import { useCurrentUser } from '@/shared/auth'
import { LoadError } from '@/shared/components/load-error'
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
  const { data: program, error, isPending, refetch } = useProgram(programId)
  const { data: user } = useCurrentUser()
  const hasOverview = can(user?.role, 'program-overview:view')
  const { data: overview } = useProgramOverview(programId, hasOverview)
  const [searchParams, setSearchParams] = useSearchParams()
  const tab = programTab(searchParams.get(TAB_PARAM), hasOverview)

  const handleRetry = () => void refetch()
  const handleTabChange = (value: string) => setSearchParams({ [TAB_PARAM]: value }, { replace: true })

  if (isPending) return <PageSkeleton tabs={Object.keys(PROGRAM_TABS).length} />
  if (!program && error && error.status !== NOT_FOUND_STATUS)
    return <LoadError framed error={error} onRetry={handleRetry} />
  if (!program)
    return <PagePlaceholder title={t('Program not found')} description={t('There is no program at this address.')} />

  return (
    <div className="flex flex-col gap-6">
      <ProgramHeader program={program} />
      <Tabs value={tab} onValueChange={handleTabChange} className="gap-5">
        <TabsList variant="line">
          {hasOverview && <TabsTrigger value={PROGRAM_TABS.overview}>{t('Overview')}</TabsTrigger>}
          <TabsTrigger value={PROGRAM_TABS.projects}>{t('Projects')}</TabsTrigger>
          <TabsTrigger value={PROGRAM_TABS.team}>{t('Team')}</TabsTrigger>
        </TabsList>
        {hasOverview && (
          <TabsContent value={PROGRAM_TABS.overview} className="flex flex-col gap-5">
            <MetricsCard overview={overview} />
            <ExpertiseStaffing rows={overview?.staffing_by_expertise ?? []} isPending={!overview} />
            <BriefCard program={program} />
          </TabsContent>
        )}
        <TabsContent value={PROGRAM_TABS.projects}>{projects}</TabsContent>
        <TabsContent value={PROGRAM_TABS.team}>{team}</TabsContent>
      </Tabs>
    </div>
  )
}
