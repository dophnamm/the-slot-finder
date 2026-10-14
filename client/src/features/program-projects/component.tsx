import { FolderOpen } from 'lucide-react'
import { useState } from 'react'
import { useCurrentUser } from '@/shared/auth'
import { ProjectFormDialog } from '@/shared/components/project-form-dialog'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { canEdit, isOpenStatus, useProgram, useProgramProjects } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { PROJECT_TABLE_COLUMNS } from './constants'
import { ProjectRow } from './components/project-row'
import type { ProgramProjectsProps } from './types'

/** Projects tab of the program page: status, timeline, current phase and risk per project. */
export function ProgramProjects({ programId }: ProgramProjectsProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: program } = useProgram(programId)
  const { data: projects = [], isPending } = useProgramProjects(programId)
  const [adding, setAdding] = useState(false)
  const canAdd = program !== undefined && canEdit(user, [program.lead]) && isOpenStatus(program.status)

  const handleAdd = () => setAdding(true)

  return (
    <Section
      title={t('Projects')}
      description={t('Ordered by priority, then deadline.')}
      action={canAdd && <Button onClick={handleAdd}>{t('Add project')}</Button>}
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Project')}</TableHead>
            <TableHead>{t('Status')}</TableHead>
            <TableHead>{t('Start – deadline')}</TableHead>
            <TableHead>{t('Current phase')}</TableHead>
            <TableHead>{t('Progress')}</TableHead>
            <TableHead className="min-w-16 text-right whitespace-normal">{t('Next week')}</TableHead>
            <TableHead className="min-w-16 text-right whitespace-normal">{t('Next 3 weeks')}</TableHead>
            <TableHead>{t('Lead')}</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {projects.map((project) => (
            <ProjectRow key={project.id} project={project} />
          ))}
          {isPending && <TableLoading columns={PROJECT_TABLE_COLUMNS} />}
          {!isPending && projects.length === 0 && (
            <TableEmpty
              columns={PROJECT_TABLE_COLUMNS}
              icon={FolderOpen}
              title={t('No projects yet.')}
              description={t('Add the program’s projects with their phases and demand.')}
            />
          )}
        </TableBody>
      </Table>
      {adding && <ProjectFormDialog open={adding} programId={programId} project={null} onOpenChange={setAdding} />}
    </Section>
  )
}
