import { Inbox } from 'lucide-react'
import { useState } from 'react'
import { Pager } from '@/shared/components/pager'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { PROGRAMS_PAGE_SIZE, useProjects } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { IntakeRow } from './components/intake-row'
import { INTAKE_STATUS, INTAKE_TABLE_COLUMNS } from './constants'

/** Incoming projects across programs, highest priority and earliest deadline first. */
export function IntakeQueue() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const { data, isPending } = useProjects(INTAKE_STATUS, page)
  const projects = data?.data ?? []

  return (
    <Section
      title={t('Intake queue')}
      description={t('Incoming projects across programs, highest priority and earliest deadline first.')}
    >
      <Table
        footer={<Pager page={page} total={data?.meta.total} pageSize={PROGRAMS_PAGE_SIZE} onPageChange={setPage} />}
      >
        <TableHeader>
          <TableRow>
            <TableHead>{t('Project')}</TableHead>
            <TableHead>{t('Program')}</TableHead>
            <TableHead>{t('Start – deadline')}</TableHead>
            <TableHead>{t('Current phase')}</TableHead>
            <TableHead>{t('Lead')}</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {projects.map((project) => (
            <IntakeRow key={project.id} project={project} />
          ))}
          {isPending && <TableLoading columns={INTAKE_TABLE_COLUMNS} />}
          {!isPending && projects.length === 0 && (
            <TableEmpty
              columns={INTAKE_TABLE_COLUMNS}
              icon={Inbox}
              title={t('No incoming projects.')}
              description={t('Projects added to a program wait here until a staffing draft covers them.')}
            />
          )}
        </TableBody>
      </Table>
    </Section>
  )
}
