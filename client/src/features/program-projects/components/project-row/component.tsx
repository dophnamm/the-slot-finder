import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { LinkRow } from '@/shared/components/link-row'
import { MetricBar } from '@/shared/components/metric-bar'
import { ProposeSquadButton } from '@/shared/components/propose-squad-button'
import { StatusBadge } from '@/shared/components/status-badge'
import { TextLink } from '@/shared/components/text-link'
import { activeWarnings } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { TableCell } from '@/shared/ui/table'
import { formatIsoWeekSpan } from '@/utils/functions'
import type { ProjectRowProps } from './types'

export function ProjectRow({ project }: ProjectRowProps) {
  const { t } = useTranslation()
  const detailPath = generatePath(ROUTES.projectDetail, { programId: project.program_id, id: project.id })
  const phase = project.current_phase
  const { metrics } = project

  return (
    <LinkRow to={detailPath}>
      <TableCell>
        <div className="flex flex-col">
          <TextLink to={detailPath}>{project.name}</TextLink>
          <span className="text-xs text-muted-foreground">
            {t('Priority {priority}', { priority: project.priority })}
          </span>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          <StatusBadge status={project.status} kind="project" />
          {metrics
            ? activeWarnings(metrics).map((warning) => (
                <Badge key={warning.flag} variant={warning.badge}>
                  {t(warning.label)}
                </Badge>
              ))
            : project.at_risk && <Badge variant="red">{t('At risk')}</Badge>}
        </div>
      </TableCell>
      <TableCell className="text-sm">
        <div className="flex flex-col">
          <span>{t('{start} – {end}', { start: project.start_week, end: project.deadline_week })}</span>
          <span className="text-xs text-muted-foreground">
            {formatIsoWeekSpan(project.start_week, project.deadline_week)}
          </span>
        </div>
      </TableCell>
      <TableCell className="text-sm">
        {phase ? (
          <div className="flex flex-col">
            <span>{phase.name}</span>
            <span className="text-xs text-muted-foreground">
              {t('{start} – {end}', { start: phase.start_week, end: phase.end_week })}
            </span>
          </div>
        ) : (
          <span className="text-muted-foreground">{t('No phases yet')}</span>
        )}
      </TableCell>
      <TableCell>
        {metrics && (
          <div className="grid min-w-56 grid-cols-2 gap-x-4 gap-y-1">
            <MetricBar compact label={t('Time')} value={metrics.time_percent} warning={metrics.past_end} />
            <MetricBar compact label={t('Scope')} value={metrics.scope_percent} warning={metrics.behind_on_delivery} />
            <MetricBar compact label={t('Planned burn')} value={metrics.planned_burn_percent} />
            <MetricBar compact label={t('Staffing')} value={metrics.staffing_percent} warning={metrics.understaffed} />
          </div>
        )}
      </TableCell>
      <TableCell>{project.lead?.name ?? t('No lead')}</TableCell>
      <TableCell className="text-right">
        <ProposeSquadButton projectId={project.id} status={project.status} />
      </TableCell>
    </LinkRow>
  )
}
