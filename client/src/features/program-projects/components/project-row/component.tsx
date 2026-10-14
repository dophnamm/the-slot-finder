import { GripVertical } from 'lucide-react'
import type { DragEvent, KeyboardEvent } from 'react'
import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { LinkRow } from '@/shared/components/link-row'
import { MetricBar } from '@/shared/components/metric-bar'
import { ProposeSquadButton } from '@/shared/components/propose-squad-button'
import { StatusBadge } from '@/shared/components/status-badge'
import { TextLink } from '@/shared/components/text-link'
import { activeWarnings, burnMetric } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { TableCell } from '@/shared/ui/table'
import { cn, formatIsoWeekSpan, shortWeekLabel } from '@/utils/functions'
import { DRAG_DATA_TYPE, MOVE_KEYS, NO_UPCOMING_HOURS } from './constants'
import type { ProjectRowProps } from './types'

export function ProjectRow({ project, reorder }: ProjectRowProps) {
  const { t } = useTranslation()
  const detailPath = generatePath(ROUTES.projectDetail, { programId: project.program_id, id: project.id })
  const phase = project.current_phase
  const { metrics } = project
  const burn = metrics ? burnMetric(metrics) : null

  const handleDragStart = (event: DragEvent<HTMLTableRowElement>) => {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData(DRAG_DATA_TYPE, project.id)
    reorder?.onDragStart(project.id)
  }
  const handleDragOver = (event: DragEvent<HTMLTableRowElement>) => {
    event.preventDefault()
    reorder?.onDragOver(project.id)
  }
  const handleDrop = (event: DragEvent<HTMLTableRowElement>) => event.preventDefault()
  const handleDragEnd = () => reorder?.onDragEnd()
  const handleKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    const delta = MOVE_KEYS[event.key]
    if (delta === undefined) return
    event.preventDefault()
    reorder?.onMove(project.id, delta)
  }

  return (
    <LinkRow
      to={detailPath}
      draggable={reorder !== undefined}
      className={cn(reorder?.dragging && 'opacity-50')}
      onDragStart={reorder && handleDragStart}
      onDragOver={reorder && handleDragOver}
      onDrop={reorder && handleDrop}
      onDragEnd={reorder && handleDragEnd}
    >
      <TableCell>
        <div className="flex items-start gap-1.5">
          {reorder && (
            <button
              type="button"
              aria-label={t('Move {name}', { name: project.name })}
              title={t('Drag to change the order, or use the arrow keys.')}
              className="mt-0.5 flex cursor-grab items-center rounded-sm text-muted-foreground outline-none hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring active:cursor-grabbing"
              onKeyDown={handleKeyDown}
            >
              <GripVertical className="size-4" />
            </button>
          )}
          <div className="flex flex-col">
            <TextLink to={detailPath}>{project.name}</TextLink>
            <span className="text-xs text-muted-foreground">
              {t('Priority {priority}', { priority: project.priority })}
            </span>
          </div>
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
          <span title={t('{start} – {end}', { start: project.start_week, end: project.deadline_week })}>
            {t('{start} – {end}', {
              start: shortWeekLabel(project.start_week),
              end: shortWeekLabel(project.deadline_week),
            })}
          </span>
          <span className="text-xs text-muted-foreground">
            {formatIsoWeekSpan(project.start_week, project.deadline_week)}
          </span>
        </div>
      </TableCell>
      <TableCell className="text-sm">
        {phase ? (
          <div className="flex flex-col">
            <span>{phase.name}</span>
            <span
              className="text-xs text-muted-foreground"
              title={t('{start} – {end}', { start: phase.start_week, end: phase.end_week })}
            >
              {t('{start} – {end}', { start: shortWeekLabel(phase.start_week), end: shortWeekLabel(phase.end_week) })}
            </span>
          </div>
        ) : (
          <span className="text-muted-foreground">{t('No phases yet')}</span>
        )}
      </TableCell>
      <TableCell>
        {metrics && burn && (
          <div className="grid min-w-56 grid-cols-2 gap-x-4 gap-y-1">
            <MetricBar compact label={t('Time')} value={metrics.time_percent} warning={metrics.past_end} />
            <MetricBar compact label={t('Scope')} value={metrics.scope_percent} warning={metrics.behind_on_delivery} />
            <MetricBar compact label={t(burn.label)} value={burn.value} />
            <MetricBar compact label={t('Staffing')} value={metrics.staffing_percent} warning={metrics.understaffed} />
          </div>
        )}
      </TableCell>
      <TableCell className="text-right text-sm tabular-nums">
        {metrics &&
          (metrics.next_week_hours > 0 ? t('{hours} h', { hours: metrics.next_week_hours }) : NO_UPCOMING_HOURS)}
      </TableCell>
      <TableCell className="text-right text-sm tabular-nums">
        {metrics &&
          (metrics.next_three_weeks_hours > 0
            ? t('{hours} h', { hours: metrics.next_three_weeks_hours })
            : NO_UPCOMING_HOURS)}
      </TableCell>
      <TableCell>{project.lead?.name ?? t('No lead')}</TableCell>
      <TableCell className="text-right">
        <ProposeSquadButton projectId={project.id} status={project.status} programLeadId={project.program_lead_id} />
      </TableCell>
    </LinkRow>
  )
}
