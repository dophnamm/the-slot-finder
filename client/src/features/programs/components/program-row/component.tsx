import { generatePath, Link } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { StatusBadge } from '@/shared/components/status-badge'
import { weekSpan } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { TableCell, TableRow } from '@/shared/ui/table'
import type { ProgramRowProps } from './types'

export function ProgramRow({ program }: ProgramRowProps) {
  const { t } = useTranslation()

  return (
    <TableRow>
      <TableCell>
        <div className="flex flex-col">
          <Link className="font-medium hover:underline" to={generatePath(ROUTES.programDetail, { id: program.id })}>
            {program.name}
          </Link>
          <span className="text-xs text-muted-foreground">{program.client}</span>
        </div>
      </TableCell>
      <TableCell>
        <StatusBadge status={program.status} kind="program" />
      </TableCell>
      <TableCell className="text-sm">{weekSpan(program.start_week, program.end_week)}</TableCell>
      <TableCell>{program.lead?.name ?? t('No lead')}</TableCell>
    </TableRow>
  )
}
