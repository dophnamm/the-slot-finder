import { RefreshCw, Trash2 } from 'lucide-react'
import { useNavigate } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { NOT_FOUND_STATUS } from '@/shared/api/constants'
import { LoadError } from '@/shared/components/load-error'
import { AssumptionText } from '@/shared/components/assumption-text'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { PageSkeleton } from '@/shared/components/page-skeleton'
import { Section } from '@/shared/components/section'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { confirm } from '@/shared/stores/confirm-store'
import { useDeleteScenario, useRerunScenario, useScenario } from '@/shared/scenarios'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { cn } from '@/utils/functions'
import { ScenarioChanges } from './components/scenario-changes'
import { ScenarioUnmet } from './components/scenario-unmet'
import type { ScenarioReviewProps } from './types'
import { projectPath } from './utils'

/** A what-if's result: its assumptions, how much it changes the plan, who gains or loses hours, and what nobody can cover. */
export function ScenarioReview({ scenarioId }: ScenarioReviewProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data: scenario, error, isPending, refetch } = useScenario(scenarioId)
  const rerun = useRerunScenario(scenarioId)
  const remove = useDeleteScenario(scenarioId)
  const busy = rerun.isPending || remove.isPending

  const handleRerun = () => rerun.mutate()
  const handleRetry = () => void refetch()
  const handleDeleted = () => navigate(ROUTES.scenarios, { replace: true })
  const handleDelete = async () => {
    const confirmed = await confirm({
      title: t('Delete this scenario?'),
      description: t('The real plan is not affected.'),
      confirmLabel: t('Delete'),
      destructive: true,
    })
    if (!confirmed) return
    remove.mutate(undefined, { onSuccess: handleDeleted })
  }

  if (isPending) return <PageSkeleton />
  if (!scenario && error && error.status !== NOT_FOUND_STATUS)
    return <LoadError framed error={error} onRetry={handleRetry} />
  if (!scenario)
    return <PagePlaceholder title={t('Scenario not found')} description={t('There is no scenario at this address.')} />
  const { summary } = scenario

  return (
    <div className="flex flex-col gap-8">
      <Section
        variant="page"
        card
        title={
          <>
            {scenario.name}
            {scenario.stale && <Badge variant="amber">{t('Stale')}</Badge>}
          </>
        }
        description={
          scenario.stale
            ? t(
                'Built on baseline version {base}; version {current} is in force now. Run it again to see today’s effect.',
                {
                  base: scenario.base_version,
                  current: scenario.current_version,
                },
              )
            : t('Built on baseline version {version}.', { version: scenario.base_version })
        }
        action={
          <>
            <Button variant="outline" size="sm" disabled={busy} isLoading={rerun.isPending} onClick={handleRerun}>
              <RefreshCw />
              {t('Run again')}
            </Button>
            <Button variant="outline" size="sm" disabled={busy} isLoading={remove.isPending} onClick={handleDelete}>
              <Trash2 />
              {t('Delete')}
            </Button>
          </>
        }
      >
        <div className="grid gap-6 lg:grid-cols-2">
          <div className="flex flex-col gap-2">
            <span className="text-sm font-medium">{t('Assumptions')}</span>
            <ul className="flex list-disc flex-col gap-1 pl-5 text-sm">
              {scenario.assumptions.map((assumption, index) => (
                <li key={`${index}:${assumption.type}`}>
                  <AssumptionText assumption={assumption} />
                </li>
              ))}
            </ul>
          </div>
          <dl className="grid grid-cols-3 gap-4">
            <div className="flex flex-col">
              <dt className="text-xs text-muted-foreground">{t('People changed')}</dt>
              <dd className="text-lg font-semibold tabular-nums">{summary.people_changed}</dd>
            </div>
            <div className="flex flex-col">
              <dt className="text-xs text-muted-foreground">{t('Hours moved')}</dt>
              <dd className="text-lg font-semibold tabular-nums">{t('{hours} hrs', { hours: summary.hours_moved })}</dd>
            </div>
            <div className="flex flex-col">
              <dt className="text-xs text-muted-foreground">{t('Unmet')}</dt>
              <dd className={cn('text-lg font-semibold tabular-nums', summary.unmet_hours > 0 && 'text-destructive')}>
                {t('{hours} hrs', { hours: summary.unmet_hours })}
              </dd>
            </div>
          </dl>
        </div>
        <div className="flex flex-col gap-1 text-sm text-muted-foreground">
          <p>
            {t(
              'Nothing real has changed. To make it real, move the phases on the project’s page or add the leave on the People page; a re-plan draft follows.',
            )}
          </p>
          {scenario.projects.length > 0 && (
            <p className="flex flex-wrap gap-x-3">
              <span>{t('Planned again:')}</span>
              {scenario.projects.map((project) => (
                <TextLink key={project.id} to={projectPath(project)}>
                  {project.name}
                </TextLink>
              ))}
            </p>
          )}
        </div>
      </Section>
      <ScenarioChanges changes={scenario.changes} projects={scenario.projects} />
      <ScenarioUnmet unmet={scenario.unmet} />
    </div>
  )
}
