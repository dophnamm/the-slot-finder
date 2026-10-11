import { useState } from 'react'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { ProgramFormDialog } from '@/shared/components/program-form-dialog'
import { PROGRAMS_PAGE_SIZE, usePrograms } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Card, CardAction, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { ProgramRow } from './components/program-row'
import type { ProgramsTableProps } from './types'
import { pageCount } from './utils'

/** Every program with its status, timeline and lead; planners and leads create programs here. */
export function ProgramsTable({ onCreated }: ProgramsTableProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const [page, setPage] = useState(1)
  const [formOpen, setFormOpen] = useState(false)
  const { data, isPending } = usePrograms(page)
  const programs = data?.data ?? []
  const pages = pageCount(data?.meta.total, PROGRAMS_PAGE_SIZE)

  const handleNew = () => setFormOpen(true)
  const handlePrevious = () => setPage((current) => Math.max(1, current - 1))
  const handleNext = () => setPage((current) => Math.min(pages, current + 1))

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
      <CardFooter className="justify-end gap-2">
        <span className="text-sm text-muted-foreground">{t('Page {page} of {pages}', { page, pages })}</span>
        <Button variant="outline" size="sm" onClick={handlePrevious} disabled={page <= 1}>
          {t('Previous')}
        </Button>
        <Button variant="outline" size="sm" onClick={handleNext} disabled={page >= pages}>
          {t('Next')}
        </Button>
      </CardFooter>
      {formOpen && <ProgramFormDialog open={formOpen} program={null} onOpenChange={setFormOpen} onSaved={onCreated} />}
    </Card>
  )
}
