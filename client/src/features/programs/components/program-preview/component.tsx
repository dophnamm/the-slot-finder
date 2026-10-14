import { MembersTable } from '@/shared/components/members-table'
import { MetricBar } from '@/shared/components/metric-bar'
import { useTranslation } from '@/shared/i18n'
import { useProgramMembers, useProgramOverview, useProgramProjects } from '@/shared/programs'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'
import { Skeleton } from '@/shared/ui/skeleton'
import { LOADING_REVEAL_CLASS } from '@/utils/constants'
import { cn, range } from '@/utils/functions'
import { PreviewProjects } from './components/preview-projects'
import type { ProgramPreviewProps } from './types'
import { PREVIEW_FIGURES } from './constants'
import { previewFigures } from './utils'

/**
 * A program at a glance, under its row: its key figures, then Projects and Members tabs like the
 * program page. Loads when opened.
 */
export function ProgramPreview({ programId }: ProgramPreviewProps) {
  const { t } = useTranslation()
  const { data: overview } = useProgramOverview(programId)
  const projects = useProgramProjects(programId)
  const members = useProgramMembers(programId)
  const projectList = projects.data ?? []
  const memberList = members.data ?? []

  return (
    <div className="flex flex-col gap-4">
      {overview && (
        <dl className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {previewFigures(overview.metrics).map((figure) => (
            <div key={figure.label} className="flex flex-col">
              <dt className="text-xs text-muted-foreground">{t(figure.label)}</dt>
              <dd className={cn('font-semibold tabular-nums', figure.warning && 'text-destructive')}>
                {t('{hours} h', { hours: figure.hours })}
              </dd>
            </div>
          ))}
          <MetricBar
            label={t('Staffing')}
            value={overview.metrics.staffing_percent}
            warning={overview.metrics.understaffed}
          />
        </dl>
      )}
      {!overview && (
        <div aria-busy className={cn('grid grid-cols-2 gap-4 sm:grid-cols-4', LOADING_REVEAL_CLASS)}>
          {range(PREVIEW_FIGURES + 1).map((figure) => (
            <div key={figure} className="flex flex-col gap-2">
              <Skeleton className="h-3 w-16" />
              <Skeleton className="h-4 w-20" />
            </div>
          ))}
        </div>
      )}
      <Tabs defaultValue="projects">
        <TabsList>
          <TabsTrigger value="projects">
            {projects.isPending ? t('Projects') : t('Projects ({count})', { count: projectList.length })}
          </TabsTrigger>
          <TabsTrigger value="members">
            {members.isPending ? t('Members') : t('Members ({count})', { count: memberList.length })}
          </TabsTrigger>
        </TabsList>
        <TabsContent value="projects">
          <PreviewProjects projects={projectList} isPending={projects.isPending} programId={programId} />
        </TabsContent>
        <TabsContent value="members">
          <MembersTable members={memberList} isPending={members.isPending} programId={programId} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
