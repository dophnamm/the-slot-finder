import { ChevronRight } from 'lucide-react'
import { useState } from 'react'
import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { LinkRow } from '@/shared/components/link-row'
import { StatusBadge } from '@/shared/components/status-badge'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { weekSpan } from '@/shared/programs'
import { Button } from '@/shared/ui/button'
import { TableCell, TableRow } from '@/shared/ui/table'
import { cn } from '@/utils/functions'
import { PROGRAM_TABLE_COLUMNS } from '../../constants'
import { ProgramPreview } from '../program-preview'
import type { ProgramRowProps } from './types'

/** A program; the arrow opens a preview of its figures, projects and members under the row. */
export function ProgramRow({ program }: ProgramRowProps) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)
  const detailPath = generatePath(ROUTES.programDetail, { id: program.id })

  const handleToggle = () => setExpanded((open) => !open)

  return (
    <>
      <LinkRow to={detailPath}>
        <TableCell>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon-sm"
              aria-expanded={expanded}
              aria-label={expanded ? t('Hide details') : t('Show details')}
              onClick={handleToggle}
            >
              <ChevronRight className={cn('transition-transform', expanded && 'rotate-90')} />
            </Button>
            <div className="flex flex-col">
              <TextLink to={detailPath}>{program.name}</TextLink>
              <span className="text-xs text-muted-foreground">{program.client}</span>
            </div>
          </div>
        </TableCell>
        <TableCell>
          <StatusBadge status={program.status} kind="program" />
        </TableCell>
        <TableCell className="text-sm">{weekSpan(program.start_week, program.end_week)}</TableCell>
        <TableCell>{program.lead?.name ?? t('No manager')}</TableCell>
      </LinkRow>
      {expanded && (
        <TableRow className="hover:bg-transparent">
          <TableCell colSpan={PROGRAM_TABLE_COLUMNS} className="bg-muted/30 p-4 whitespace-normal">
            <ProgramPreview programId={program.id} />
          </TableCell>
        </TableRow>
      )}
    </>
  )
}
