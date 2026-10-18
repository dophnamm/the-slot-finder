import { FolderOpen } from 'lucide-react'
import { useState } from 'react'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { ProjectFormDialog } from '@/shared/components/project-form-dialog'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { canEdit, isOpenStatus, useProgram, useProgramProjects } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { HIDDEN_FIGURE_COLUMNS, PROJECT_TABLE_COLUMNS } from './constants'
import { IntakeDialog } from './components/intake-dialog'
import { ProjectRow } from './components/project-row'
import type { ProgramProjectsProps } from './types'
import { useProjectOrder } from './use-project-order'

/**
 * Projects tab of the program page: status, timeline, current phase and risk per project. Their order is
 * their priority; planners and the program's manager drag a row to change it.
 */
export function ProgramProjects({ programId }: ProgramProjectsProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: program } = useProgram(programId)
  const { data: projects = [], isPending } = useProgramProjects(programId)
  const [adding, setAdding] = useState(false)
  const [suggesting, setSuggesting] = useState(false)
  const canAdd = program !== undefined && canEdit(user, [program.lead]) && isOpenStatus(program.status)
  const showFigures = can(user?.role, 'program-overview:view')
  const columns = PROJECT_TABLE_COLUMNS - (showFigures ? 0 : HIDDEN_FIGURE_COLUMNS)
  const canReorder = canAdd && projects.length > 1
  const { rows, draggingId, handlers } = useProjectOrder(programId, projects)

  const handleAdd = () => setAdding(true)
  const handleSuggest = () => setSuggesting(true)

  return (
    <Section
      title={t('Projects')}
      description={
        canReorder
          ? t('Drag a row to change the order: the top project has the highest priority.')
          : t('Ordered by priority, then deadline.')
      }
      action={
        canAdd && (
          <div className="flex gap-2">
            <Button variant="outline" onClick={handleSuggest}>
              {t('New project from a brief')}
            </Button>
            <Button onClick={handleAdd}>{t('Add project')}</Button>
          </div>
        )
      }
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Project')}</TableHead>
            <TableHead>{t('Status')}</TableHead>
            <TableHead>{t('Start – deadline')}</TableHead>
            <TableHead>{t('Current phase')}</TableHead>
            {showFigures && (
              <>
                <TableHead>{t('Progress')}</TableHead>
                <TableHead className="min-w-16 text-right whitespace-normal">{t('Next week')}</TableHead>
                <TableHead className="min-w-16 text-right whitespace-normal">{t('Next 3 weeks')}</TableHead>
              </>
            )}
            <TableHead>{t('Lead')}</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((project) => (
            <ProjectRow
              key={project.id}
              project={project}
              showFigures={showFigures}
              reorder={canReorder ? { ...handlers, dragging: draggingId === project.id } : undefined}
            />
          ))}
          {isPending && <TableLoading columns={columns} />}
          {!isPending && projects.length === 0 && (
            <TableEmpty
              columns={columns}
              icon={FolderOpen}
              title={t('No projects yet.')}
              description={t('Add the program’s projects with their phases and demand.')}
            />
          )}
        </TableBody>
      </Table>
      {adding && <ProjectFormDialog open={adding} programId={programId} project={null} onOpenChange={setAdding} />}
      {suggesting && <IntakeDialog open={suggesting} programId={programId} onOpenChange={setSuggesting} />}
    </Section>
  )
}
