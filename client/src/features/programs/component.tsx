import { useState } from 'react'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { ProgramFormDialog } from '@/shared/components/program-form-dialog'
import { PROGRAMS_PAGE_SIZE, usePrograms } from '@/shared/programs'
import { Pager } from '@/shared/components/pager'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Card, CardAction, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { ProgramRow } from './components/program-row'
import type { ProgramsTableProps } from './types'

/** Every program with its status, timeline and lead; planners and leads create programs here. */
export function ProgramsTable({ onCreated }: ProgramsTableProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const [page, setPage] = useState(1)
  const [formOpen, setFormOpen] = useState(false)
  const { data, isPending } = usePrograms(page)
  const programs = data?.data ?? []

  const handleNew = () => setFormOpen(true)

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Programs')}</CardTitle>
        <CardDescription>{t('Customer engagements and their projects.')}</CardDescription>
        {can(user?.role, 'programs:manage') && (
          <CardAction>
            <Button onClick={handleNew}>{t('New program')}</Button>
          </CardAction>
        )}
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Program')}</TableHead>
              <TableHead>{t('Status')}</TableHead>
              <TableHead>{t('Timeline')}</TableHead>
              <TableHead>{t('Lead')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {programs.map((program) => (
              <ProgramRow key={program.id} program={program} />
            ))}
          </TableBody>
        </Table>
        {isPending && <p className="py-6 text-center text-sm text-muted-foreground">{t('Loading…')}</p>}
        {!isPending && programs.length === 0 && (
          <p className="py-6 text-center text-sm text-muted-foreground">{t('No programs yet.')}</p>
        )}
      </CardContent>
      <CardFooter className="justify-end">
        <Pager page={page} total={data?.meta.total} pageSize={PROGRAMS_PAGE_SIZE} onPageChange={setPage} />
      </CardFooter>
      {formOpen && <ProgramFormDialog open={formOpen} program={null} onOpenChange={setFormOpen} onSaved={onCreated} />}
    </Card>
  )
}
