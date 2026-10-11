import { useState } from 'react'
import { useCurrentUser } from '@/shared/auth'
import { ProjectFormDialog } from '@/shared/components/project-form-dialog'
import { canEdit, isOpenStatus, useProgram, useProgramProjects } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
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
    <Card>
      <CardHeader>
        <CardTitle>{t('Projects')}</CardTitle>
        <CardDescription>{t('Ordered by priority, then deadline.')}</CardDescription>
        {canAdd && (
          <CardAction>
            <Button onClick={handleAdd}>{t('Add project')}</Button>
          </CardAction>
        )}
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Project')}</TableHead>
              <TableHead>{t('Status')}</TableHead>
              <TableHead>{t('Start – deadline')}</TableHead>
              <TableHead>{t('Current phase')}</TableHead>
              <TableHead>{t('Lead')}</TableHead>
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
      </CardContent>
      {adding && <ProjectFormDialog open={adding} programId={programId} project={null} onOpenChange={setAdding} />}
    </Card>
  )
}
