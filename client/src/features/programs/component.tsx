import { FolderKanban } from 'lucide-react'
import { useState } from 'react'
import { can } from '@/resources/rbac'
import { useCurrentUser } from '@/shared/auth'
import { ProgramFormDialog } from '@/shared/components/program-form-dialog'
import { PROGRAMS_PAGE_SIZE, usePrograms } from '@/shared/programs'
import { Pager } from '@/shared/components/pager'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { PROGRAM_TABLE_COLUMNS } from './constants'
import { ProgramRow } from './components/program-row'
import type { ProgramsTableProps } from './types'

/** Every program with its status, timeline and lead; planners and leads create programs here. */
export function ProgramsTable({ onCreated }: ProgramsTableProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const [page, setPage] = useState(1)
  const [formOpen, setFormOpen] = useState(false)
  const { data, isPending, isPlaceholderData } = usePrograms(page)
  const programs = data?.data ?? []

  const handleNew = () => setFormOpen(true)

  return (
    <Section
      refreshing={isPlaceholderData}
      className="min-h-0 flex-1"
      title={t('Programs')}
      description={t('Customer engagements and their projects.')}
      action={can(user?.role, 'programs:manage') && <Button onClick={handleNew}>{t('New program')}</Button>}
    >
      <Table
        scroll="both"
        footer={<Pager page={page} total={data?.meta.total} pageSize={PROGRAMS_PAGE_SIZE} onPageChange={setPage} />}
      >
        <TableHeader>
          <TableRow>
            <TableHead>{t('Program')}</TableHead>
            <TableHead>{t('Status')}</TableHead>
            <TableHead>{t('Timeline')}</TableHead>
            <TableHead>{t('Manager')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {programs.map((program) => (
            <ProgramRow key={program.id} program={program} />
          ))}
          {isPending && <TableLoading columns={PROGRAM_TABLE_COLUMNS} />}
          {!isPending && programs.length === 0 && (
            <TableEmpty
              columns={PROGRAM_TABLE_COLUMNS}
              icon={FolderKanban}
              title={t('No programs yet.')}
              description={t('A program groups a customer’s projects under one timeline.')}
            />
          )}
        </TableBody>
      </Table>
      {formOpen && <ProgramFormDialog open={formOpen} program={null} onOpenChange={setFormOpen} onSaved={onCreated} />}
    </Section>
  )
}
