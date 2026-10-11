import { useState } from 'react'
import { useCurrentUser } from '@/shared/auth'
import { ProjectFormDialog } from '@/shared/components/project-form-dialog'
import { Section } from '@/shared/components/section'
import { canEdit, isOpenStatus, useProgram, useProgramProjects } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
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
            <TableHead>{t('Lead')}</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {projects.map((project) => (
            <ProjectRow key={project.id} project={project} />
          ))}
        </TableBody>
      </Table>
      {isPending && <p className="py-6 text-center text-sm text-muted-foreground">{t('Loading…')}</p>}
      {!isPending && projects.length === 0 && (
        <p className="py-6 text-center text-sm text-muted-foreground">{t('No projects yet.')}</p>
      )}
      {adding && <ProjectFormDialog open={adding} programId={programId} project={null} onOpenChange={setAdding} />}
    </Section>
  )
}
