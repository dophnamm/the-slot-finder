import { FolderOpen } from 'lucide-react'
import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { MetricBar } from '@/shared/components/metric-bar'
import { StatusBadge } from '@/shared/components/status-badge'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { PREVIEW_PROJECT_COLUMNS } from './constants'
import type { PreviewProjectsProps } from './types'
import { projectWarnings } from './utils'

/** The program's projects in short: status and warnings, deadline and staffing. */
export function PreviewProjects({ projects, isPending, programId }: PreviewProjectsProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Project')}</TableHead>
            <TableHead>{t('Status')}</TableHead>
            <TableHead>{t('Deadline')}</TableHead>
            <TableHead className="w-48">{t('Staffing')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {projects.map((project) => (
            <TableRow key={project.id}>
              <TableCell>
                <TextLink to={generatePath(ROUTES.projectDetail, { programId, id: project.id })}>
                  {project.name}
                </TextLink>
              </TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  <StatusBadge status={project.status} kind="project" />
                  {projectWarnings(project).map((warning) => (
                    <Badge key={warning.flag} variant={warning.badge}>
                      {t(warning.label)}
                    </Badge>
                  ))}
                </div>
              </TableCell>
              <TableCell className="text-sm">{project.deadline_week}</TableCell>
              <TableCell>
                {project.metrics && (
                  <MetricBar
                    compact
                    label={t('Staffing')}
                    value={project.metrics.staffing_percent}
                    warning={project.metrics.understaffed}
                  />
                )}
              </TableCell>
            </TableRow>
          ))}
          {isPending && <TableLoading columns={PREVIEW_PROJECT_COLUMNS} />}
          {!isPending && projects.length === 0 && (
            <TableEmpty columns={PREVIEW_PROJECT_COLUMNS} icon={FolderOpen} title={t('No projects yet.')} />
          )}
        </TableBody>
      </Table>
    </div>
  )
}
