import { generatePath, Link } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { LinkRow } from '@/shared/components/link-row'
import { ProposeSquadButton } from '@/shared/components/propose-squad-button'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { TableCell } from '@/shared/ui/table'
import { formatIsoWeekSpan } from '@/utils/functions'
import type { IntakeRowProps } from './types'

export function IntakeRow({ project }: IntakeRowProps) {
  const { t } = useTranslation()
  const detailPath = generatePath(ROUTES.projectDetail, { programId: project.program_id, id: project.id })
  const phase = project.current_phase

  return (
    <LinkRow to={detailPath}>
      <TableCell>
        <div className="flex flex-col gap-1">
          <div className="flex flex-wrap items-center gap-2">
            <Link className="font-medium hover:underline" to={detailPath}>
              {project.name}
            </Link>
            {project.at_risk && <Badge variant="red">{t('At risk')}</Badge>}
          </div>
          <span className="text-xs text-muted-foreground">{t('Priority {priority}', { priority: project.priority })}</span>
        </div>
      </TableCell>
      <TableCell>
        <Link className="hover:underline" to={generatePath(ROUTES.programDetail, { id: project.program_id })}>
          {project.program_name ?? t('Program')}
        </Link>
      </TableCell>
      <TableCell className="text-sm">
        <div className="flex flex-col">
          <span>{t('{start} – {end}', { start: project.start_week, end: project.deadline_week })}</span>
          <span className="text-xs text-muted-foreground">{formatIsoWeekSpan(project.start_week, project.deadline_week)}</span>
        </div>
      </TableCell>
      <TableCell className="text-sm">
        {phase ? (
          <div className="flex flex-col">
            <span>{phase.name}</span>
            <span className="text-xs text-muted-foreground">{t('{start} – {end}', { start: phase.start_week, end: phase.end_week })}</span>
          </div>
        ) : (
          <span className="text-muted-foreground">{t('No phases yet')}</span>
        )}
      </TableCell>
      <TableCell>{project.lead?.name ?? t('No lead')}</TableCell>
      <TableCell className="text-right">
        <ProposeSquadButton projectId={project.id} status={project.status} />
      </TableCell>
    </LinkRow>
  )
}
